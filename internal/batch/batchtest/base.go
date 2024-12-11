package batchtest

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/massexecutorbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/metricsbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/oneexecbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/prefetchbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/randomstore"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/socketsubscribe"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"gopkg.in/yaml.v3"
)

func baseExecute(
	ctx context.Context,
	ctr *app.Container,
	filename string,
	store *sync.Map,
	threadOnlyStore *sync.Map,
	outputRoot string,
	metricsOutputRoot string,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fmt.Println("Execute...", filename)

	selfLoopCount := ""
	if v, exists := threadOnlyStore.Load("loopCount"); exists {
		selfLoopCount = v.(string)
	}

	file, err := os.Open(filepath.Join(ctr.Config.Batch.Test.Path, filename))
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var conf BatchTestType
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&conf); err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
			return fmt.Errorf("failed to wait: %v", err)
		}
		return fmt.Errorf("failed to seek file: %v", err)
	}

	if err := wait(ctx, ctr, conf, SleepAfterInit); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	if conf.Prefetch.Enabled {
		var replacements = make(map[string]string)
		if replacements, err = prefetchbatch.PrefetchBatch(ctx, ctr, conf.Prefetch, store); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute prefetch: %v", err)
		}

		ctr.Logger.Debug(ctx, "replacements set",
			logger.Value("replacements", replacements))
	}

	if err := wait(ctx, ctr, conf, SleepAfterPrefetch); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
			return fmt.Errorf("failed to wait: %v", err)
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	content := buffer.String()
	placeholderRegex := regexp.MustCompile(`\<\.\.\<\s*(\w+)\s*\>\.\.\>`)

	result := placeholderRegex.ReplaceAllStringFunc(content, func(match string) string {
		originKey := placeholderRegex.FindStringSubmatch(match)[1]
		keys := strings.Split(originKey, "_")
		var builder strings.Builder
		for i, k := range keys {
			if v, exists := threadOnlyStore.Load(k); exists {
				builder.WriteString(v.(string))
			} else {
				builder.WriteString(k)
			}

			if i != len(keys)-1 {
				builder.WriteString("_")
			}
		}
		key := builder.String()

		if v, exists := store.Load(key); exists {
			return v.(string)
		}

		if key != originKey {
			return key
		}

		return match
	})

	var yamlData map[string]interface{}

	if err := yaml.Unmarshal([]byte(result), &yamlData); err != nil {
		if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
			return fmt.Errorf("failed to wait: %v", err)
		}
		return fmt.Errorf("failed to parse as YAML: %w", err)
	}

	ctr.Logger.Debug(ctx, "replaced content",
		logger.Value("content", yamlData))

	if err := wait(ctx, ctr, conf, SleepAfterReplace); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	reader := bytes.NewReader([]byte(result))

	if conf.Metrics.Enabled {
		ctr.Logger.Debug(ctx, "metrics enabled",
			logger.Value("metrics", conf.Metrics))
		err = os.MkdirAll(metricsOutputRoot, os.ModePerm)
		if err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to create directory: %v", err)
		}
		if err := metricsbatch.MetricsFetchBatch(
			ctx,
			ctr,
			conf.Metrics,
			bytes.NewReader([]byte(result)),
			conf.Type,
			metricsOutputRoot,
		); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute metrics fetch: %v", err)
		}
	}

	if err := wait(ctx, ctr, conf, SleepAfterMetrics); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	if conf.Output.Enabled {
		switch conf.Type {
		case "MassExecute", "Pipeline", "SocketSubscribe", "SocketConnectAndSubscribe":
			err := os.MkdirAll(outputRoot, os.ModePerm)
			if err != nil {
				if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
					return fmt.Errorf("failed to wait: %v", err)
				}
				return fmt.Errorf("failed to create directory: %v", err)
			}
		}
	}

	fmt.Println("Executing...", conf.Type)

	switch conf.Type {
	case "RandomStoreValue":
		var randomStoreValue randomstore.RandomStoreValueConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&randomStoreValue); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var values map[string]string
		if values, err = randomstore.RandomStoreValueBatch(
			ctx,
			ctr,
			randomStoreValue,
			bytes.NewReader([]byte(result)),
			store,
		); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute random store value: %v", err)
		}
		ctr.Logger.Info(ctx, "newValues",
			logger.Value("values", values))
	case "OneExecute":
		var oneExec oneexecbatch.OneExecuteConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&oneExec); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var values map[string]string
		if values, err = oneexecbatch.OneExecuteBatch(ctx, ctr, oneExec, store); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute one execute: %v", err)
		}
		ctr.Logger.Info(ctx, "newValues",
			logger.Value("values", values))
		fmt.Println("newValues set complete", values)
	case "MassExecute":
		var massExec massexecutorbatch.MassExecute
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&massExec); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		if err := massexecutorbatch.MassExecuteBatch(ctx, ctr, massExec, outputRoot); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute mass execute: %v", err)
		}
	case "SocketConnect":
		var socketConnect socketsubscribe.SocketConnectConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&socketConnect); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		if err := socketsubscribe.SocketConnect(ctx, ctr, socketConnect, store, outputRoot); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute socket connect: %v", err)
		}
	case "SocketSubscribe":
		var socketSubscribe socketsubscribe.SocketSubscribeConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&socketSubscribe); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		if err := socketsubscribe.SocketSubscribe(ctx, ctr, socketSubscribe, store, selfLoopCount, outputRoot); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute socket subscribe: %v", err)
		}
	case "SocketConnectAndSubscribe":
		var socketSubscribe socketsubscribe.SocketConnectAndSubscribeConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&socketSubscribe); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		if err := socketsubscribe.SocketConnectAndSubscribe(ctx, ctr, socketSubscribe, store, outputRoot); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute socket subscribe: %v", err)
		}
	case "Pipeline":
		var pipeline PipelineConfig
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&pipeline); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		if err := pipelineBatch(ctx, ctr, pipeline, store, selfLoopCount, outputRoot, metricsOutputRoot); err != nil {
			if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute pipeline: %v", err)
		}
	default:
		if err := wait(ctx, ctr, conf, SleepAfterFailedExec); err != nil {
			return fmt.Errorf("failed to wait: %v", err)
		}
		return fmt.Errorf("unknown type")
	}

	fmt.Println("Execute complete", conf.Type, filename)

	if err := wait(ctx, ctr, conf, SleepAfterSuccessExec); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	return nil
}

func wait(ctx context.Context, ctr *app.Container, conf BatchTestType, after SleepAfterValue) error {
	if v, wait := conf.RetrieveSleepValue(after); wait {
		ctr.Logger.Debug(ctx, "sleeping after execute",
			logger.Value("duration", v))
		fmt.Println("sleeping for", v, "...")
		select {
		case <-time.After(v):
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		}
		fmt.Println("sleeping complete")
	}

	return nil
}
