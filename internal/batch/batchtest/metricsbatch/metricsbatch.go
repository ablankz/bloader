package metricsbatch

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"github.com/jmespath/go-jmespath"
	"gopkg.in/yaml.v3"
)

func MetricsFetchBatch(
	ctx context.Context,
	ctr *app.Container,
	conf MetricsBatchConfig,
	rowConf io.Reader,
	testType string,
	outputRoot string,
) error {
	concurrentCount := len(conf.Requests)
	threadExecutors := make([]*MetricsThreadExecutor, concurrentCount)

	var rawData BatchConfigWithRawMetrics
	decoder := yaml.NewDecoder(rowConf)
	if err := decoder.Decode(&rawData); err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	for i, req := range conf.Requests {
		factor := GetMetricsFetcherFactory(NewMetricsTypeFromStr(req.Type))
		if factor == nil {
			return fmt.Errorf("unknown test type: %s", req.Type)
		}
		byteData, err := yaml.Marshal(rawData.Metrics.Requests[i])
		if err != nil {
			return fmt.Errorf("failed to marshal yaml: %v", err)
		}

		if err := factor.Init(byteData); err != nil {
			return fmt.Errorf("failed to init factory: %v", err)
		}
		fetcher, err := factor.FetcherFactory(ctx, ctr)
		if err != nil {
			return fmt.Errorf("failed to create fetcher: %v", err)
		}

		logFilePath := fmt.Sprintf("%s/metrics_%s.csv", outputRoot, req.ID)
		file, err := os.Create(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}

		writer := csv.NewWriter(file)
		header := make([]string, 0, len(req.Data)+5)
		header = append(header, "Success", "SendDatetime", "ReceivedDatetime", "ResponseTime", "StatusCode")
		for _, d := range req.Data {
			header = append(header, d.Key)
		}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write header: %v", err)
		}
		writer.Flush()

		writeFunc := func(
			ctx context.Context,
			ctr *app.Container,
			data executor.ResponseContent[any],
		) error {
			records := make([]string, 0, len(req.Data)+5)
			records = append(
				records,
				fmt.Sprintf("%v", data.Success),
				data.StartTime.Format(time.RFC3339Nano),
				data.EndTime.Format(time.RFC3339Nano),
				fmt.Sprintf("%v", data.ResponseTime),
				fmt.Sprintf("%v", data.StatusCode),
			)
			for _, d := range req.Data {
				jmesPathQuery := d.JMESPath
				result, err := jmespath.Search(jmesPathQuery, data.Res)
				if err != nil {
					ctr.Logger.Error(ctx, "failed to search jmespath",
						logger.Value("error", err), logger.Value("on", "metricsFetchBatch"))
					return fmt.Errorf("failed to search jmespath: %v", err)
				}
				if result == nil {
					switch d.OnNil {
					case "cancel":
						ctr.Logger.Warn(ctx, "cancel nil value",
							logger.Value("on", "metricsFetchBatch"))
						return fmt.Errorf("cancel nil value")
					default:
						ctr.Logger.Warn(ctx, "ignore nil value",
							logger.Value("on", "metricsFetchBatch"))
					}
				}

				records = append(records, fmt.Sprintf("%v", result))
			}

			writer := csv.NewWriter(file)
			if err := writer.Write(records); err != nil {
				ctr.Logger.Error(ctx, "failed to write data to csv",
					logger.Value("error", err), logger.Value("on", "metricsFetchBatch"))
			}
			writer.Flush()

			return nil
		}

		threadExecutors[i] = &MetricsThreadExecutor{
			fetcher: fetcher,
			writer:  writeFunc,
			closer:  func() { file.Close() },
		}
	}

	go func() {
		defer func(threadExecutors []*MetricsThreadExecutor) {
			for _, executor := range threadExecutors {
				executor.closer()
			}
		}(threadExecutors)

		var wg sync.WaitGroup

		for _, executor := range threadExecutors {
			wg.Add(1)

			go func(executor *MetricsThreadExecutor) {
				ctx, threadCancel := context.WithCancel(ctx)
				defer wg.Done()
				defer threadCancel()

				termChan, err := executor.fetcher.Fetch(ctx, ctr, executor.writer)
				if err != nil {
					ctr.Logger.Error(ctx, "failed to start fetch metrics",
						logger.Value("error", err), logger.Value("on", "metricsFetchBatch"))
					threadCancel()
				}

				select {
				case <-ctx.Done():
					ctr.Logger.Info(ctx, "Metrics Fetch End For Term",
						logger.Value("on", "metricsFetchBatch"))
					return
				case termType := <-termChan:
					switch termType {
					case TermTypeContext:
						ctr.Logger.Info(ctx, "Metrics Fetch End For Term",
							logger.Value("on", "metricsFetchBatch"))
					case TermWriteError:
						ctr.Logger.Error(ctx, "failed to write metrics",
							logger.Value("on", "metricsFetchBatch"))
						threadCancel()
						return
					}
				}
			}(executor)
		}

		wg.Wait()
	}()

	return nil
}
