package runner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"os"
	"sync"
	"time"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/runner/massexec"
	"github.com/ablankz/bloader/internal/runner/memoryvalue"
	"github.com/ablankz/bloader/internal/runner/storeimport"
	"github.com/ablankz/bloader/internal/runner/storevalue"
)

func baseExecute(
	ctx context.Context,
	ctr *container.Container,
	filepath string,
	str *sync.Map,
	threadOnlyStr *sync.Map,
	outputRoot string,
	outputCtr output.OutputContainer,
	index int,
	callCount int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	yamlTemplate := buffer.String()

	preTmpl, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(yamlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}

	replacedValuesData := make(map[string]any)
	replaceThreadValuesData := make(map[string]any)

	str.Range(func(key, value any) bool {
		if byteData, ok := value.([]byte); ok {
			var v any
			json.Unmarshal(byteData, &v)
			replacedValuesData[key.(string)] = v
			return true
		}
		replacedValuesData[key.(string)] = value
		return true
	})

	threadOnlyStr.Range(func(key, value any) bool {
		if byteData, ok := value.([]byte); ok {
			var v any
			json.Unmarshal(byteData, &v)
			replaceThreadValuesData[key.(string)] = v
			return true
		}
		replaceThreadValuesData[key.(string)] = value
		return true
	})

	preData := map[string]any{
		"Values":       replacedValuesData,
		"ThreadValues": replaceThreadValuesData,
	}

	var preYamlBuf *bytes.Buffer = new(bytes.Buffer)
	if err := preTmpl.Execute(preYamlBuf, preData); err != nil {
		return fmt.Errorf("failed to execute yaml: %v", err)
	}

	var preRunner Runner
	if err := yaml.Unmarshal(preYamlBuf.Bytes(), &preRunner); err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	validPreRunner, err := preRunner.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate runner: %v", err)
	}

	if validPreRunner.StoreImport.Enabled {
		for _, d := range validPreRunner.StoreImport.Data {
			valBytes, err := ctr.Store.GetObject(d.BucketID, d.StoreKey)
			if err != nil {
				return fmt.Errorf("failed to get object: %v", err)
			}
			if d.Encrypt.Enabled {
				encryptor, ok := ctr.EncypterContainer[d.Encrypt.EncryptID]
				if !ok {
					return fmt.Errorf("encryptor not found: %s", d.Encrypt.EncryptID)
				}
				decryptedVal, err := encryptor.Decrypt(string(valBytes))
				if err != nil {
					return fmt.Errorf("failed to decrypt value: %v", err)
				}
				valBytes = []byte(decryptedVal)
			}
			var val any
			if err := json.Unmarshal(valBytes, &val); err != nil {
				return fmt.Errorf("failed to unmarshal value: %v, if the value encrypted, please make sure the value is decrypted", err)
			}
			if d.ThreadOnly {
				threadOnlyStr.Store(d.Key, valBytes)
				replaceThreadValuesData[d.Key] = val
			} else {
				str.Store(d.Key, valBytes)
				replacedValuesData[d.Key] = val
			}
		}
	}

	tmpl, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(yamlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}

	data := map[string]any{
		"Values":       replacedValuesData,
		"ThreadValues": replaceThreadValuesData,
		"Dynamic": map[string]any{
			"OutputRoot": outputRoot,
			"LoopCount":  index,
			"CallCount":  callCount,
		},
	}

	var yamlBuf *bytes.Buffer = new(bytes.Buffer)
	if err := tmpl.Execute(yamlBuf, data); err != nil {
		return fmt.Errorf("failed to execute yaml: %v", err)
	}

	var rawData bytes.Buffer
	reader := io.TeeReader(yamlBuf, &rawData)

	var conf Runner
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&conf); err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	validRunner, err := conf.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate runner: %v", err)
	}

	if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterInit); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	fmt.Println("rawData", rawData.String())

	switch validRunner.Kind {
	case RunnerKindStoreValue:
		var storeValue storevalue.StoreValue
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&storeValue); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validStoreValue storevalue.ValidStoreValue
		if validStoreValue, err = storeValue.Validate(); err != nil {
			return fmt.Errorf("failed to validate store value: %v", err)
		}
		if err := validStoreValue.Run(ctx, ctr); err != nil {
			if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute store value: %v", err)
		}
		ctr.Logger.Info(ctx, "executed store value")
	case RunnerKindMemoryValue:
		var memoryStoreValue memoryvalue.MemoryValue
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&memoryStoreValue); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validMemoryValue memoryvalue.ValidMemoryValue
		if validMemoryValue, err = memoryStoreValue.Validate(); err != nil {
			return fmt.Errorf("failed to validate memory store value: %v", err)
		}
		if err := validMemoryValue.Run(ctx, ctr, str); err != nil {
			if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute memory store value: %v", err)
		}
		ctr.Logger.Info(ctx, "executed memory store value")
	case RunnerKindStoreImport:
		var storeImport storeimport.StoreImport
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&storeImport); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validStoreImport storeimport.ValidStoreImport
		if validStoreImport, err = storeImport.Validate(); err != nil {
			return fmt.Errorf("failed to validate store import: %v", err)
		}
		if err := validStoreImport.Run(ctx, ctr, str); err != nil {
			if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute store import: %v", err)
		}
		ctr.Logger.Info(ctx, "executed store import")
	case RunnerKindMassExecute:
		var massExec massexec.MassExec
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&massExec); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validMassExec massexec.ValidMassExec
		if validMassExec, err = massExec.Validate(ctr, outputCtr); err != nil {
			return fmt.Errorf("failed to validate mass exec: %v", err)
		}
		if err := validMassExec.Run(ctx, ctr, outputRoot, str, threadOnlyStr); err != nil {
			if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute mass exec: %v", err)
		}
		ctr.Logger.Info(ctx, "executed mass exec")
	default:
		return fmt.Errorf("invalid runner kind: %s", validRunner.Kind)
	}

	if err := wait(ctx, ctr, validRunner, RunnerSleepValueAfterExec); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	return nil
}

func wait(
	ctx context.Context,
	ctr *container.Container,
	conf ValidRunner,
	after RunnerSleepValueAfter,
) error {
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
