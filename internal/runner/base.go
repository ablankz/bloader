package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"github.com/ablankz/bloader/internal/logger"
)

// BaseExecutor represents the base executor
type BaseExecutor struct {
	Logger       logger.Logger
	TmplFactor   TmplFactor
	Store        Store
	AuthFactor   AuthenticatorFactor
	OutputFactor OutputFactor
	TargetFactor TargetFactor
}

// Execute executes the base executor
func (e BaseExecutor) Execute(
	ctx context.Context,
	filename string,
	str *sync.Map,
	threadOnlyStr *sync.Map,
	outputRoot string,
	index int,
	callCount int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tmplStr, err := e.TmplFactor.TmplFactorize(ctx, filename)
	if err != nil {
		return fmt.Errorf("failed to factorize template: %v", err)
	}

	tmpl, err := template.New("yaml").Funcs(sprig.TxtFuncMap()).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse yaml: %v", err)
	}
	replacedValuesData := make(map[string]any)
	replaceThreadValuesData := make(map[string]any)

	str.Range(func(key, value any) bool {
		replacedValuesData[key.(string)] = value
		return true
	})

	threadOnlyStr.Range(func(key, value any) bool {
		replaceThreadValuesData[key.(string)] = value
		return true
	})

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

	var runner Runner
	decoder := yaml.NewDecoder(reader)
	if err := decoder.Decode(&runner); err != nil {
		return fmt.Errorf("failed to decode yaml: %v", err)
	}

	validRunner, err := runner.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate runner: %v", err)
	}

	if validRunner.StoreImport.Enabled {
		e.Store.Import(
			ctx,
			validRunner.StoreImport.Data,
			func(ctx context.Context, data ValidStoreImportData, val any) error {
				if data.ThreadOnly {
					threadOnlyStr.Store(data.Key, val)
					replaceThreadValuesData[data.Key] = val
				} else {
					str.Store(data.Key, val)
					replacedValuesData[data.Key] = val
				}
				return nil
			},
		)

		data = map[string]any{
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
		rawData.Reset()
		reader := io.TeeReader(yamlBuf, &rawData)
		decoder := yaml.NewDecoder(reader)
		if err := decoder.Decode(&runner); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}

		validRunner, err = runner.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate runner: %v", err)
		}
	}

	if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterInit); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	switch validRunner.Kind {
	case RunnerKindStoreValue:
		var storeValue StoreValue
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&storeValue); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validStoreValue ValidStoreValue
		if validStoreValue, err = storeValue.Validate(); err != nil {
			return fmt.Errorf("failed to validate store value: %v", err)
		}
		if err := validStoreValue.Run(ctx, e.Store); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute store value: %v", err)
		}
		e.Logger.Info(ctx, "executed store value")
	case RunnerKindMemoryValue:
		var memoryStoreValue MemoryValue
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&memoryStoreValue); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validMemoryValue ValidMemoryValue
		if validMemoryValue, err = memoryStoreValue.Validate(); err != nil {
			return fmt.Errorf("failed to validate memory store value: %v", err)
		}
		if err := validMemoryValue.Run(ctx, str); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute memory store value: %v", err)
		}
		e.Logger.Info(ctx, "executed memory store value")
	case RunnerKindStoreImport:
		var storeImport StoreImport
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&storeImport); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validStoreImport ValidStoreImport
		if validStoreImport, err = storeImport.Validate(); err != nil {
			return fmt.Errorf("failed to validate store import: %v", err)
		}
		if err := validStoreImport.Run(ctx, e.Store, str); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute store import: %v", err)
		}
		e.Logger.Info(ctx, "executed store import")
	case RunnerKindOneExecute:
		var oneExec OneExec
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&oneExec); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validOneExec ValidOneExec
		if validOneExec, err = oneExec.Validate(ctx, e.AuthFactor, e.OutputFactor, e.TargetFactor); err != nil {
			return fmt.Errorf("failed to validate one exec: %v", err)
		}
		if err := validOneExec.Run(ctx, outputRoot, str, e.Logger, e.Store); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute one exec: %v", err)
		}
		e.Logger.Info(ctx, "executed one exec")
	case RunnerKindMassExecute:
		var massExec MassExec
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&massExec); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validMassExec ValidMassExec
		if validMassExec, err = massExec.Validate(
			ctx,
			e.Logger,
			e.AuthFactor,
			e.OutputFactor,
			e.TargetFactor,
			tmplStr,
			data,
		); err != nil {
			return fmt.Errorf("failed to validate mass exec: %v", err)
		}
		if err := validMassExec.Run(
			ctx,
			e.Logger,
			outputRoot,
			e.AuthFactor,
			e.OutputFactor,
			e.TargetFactor,
		); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute mass exec: %v", err)
		}
		e.Logger.Info(ctx, "executed mass exec")
	case RunnerKindFlow:
		var flow Flow
		decoder := yaml.NewDecoder(&rawData)
		if err := decoder.Decode(&flow); err != nil {
			return fmt.Errorf("failed to decode yaml: %v", err)
		}
		var validFlow ValidFlow
		if validFlow, err = flow.Validate(); err != nil {
			return fmt.Errorf("failed to validate flow: %v", err)
		}
		if err := validFlow.Run(
			ctx,
			e.Logger,
			e.TmplFactor,
			e.Store,
			e.AuthFactor,
			e.OutputFactor,
			e.TargetFactor,
			str,
			outputRoot,
			callCount,
		); err != nil {
			if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterFailedExec); err != nil {
				return fmt.Errorf("failed to wait: %v", err)
			}
			return fmt.Errorf("failed to execute flow: %v", err)
		}
		e.Logger.Info(ctx, "executed flow")
	default:
		return fmt.Errorf("invalid runner kind: %s", validRunner.Kind)
	}

	if err := wait(ctx, e.Logger, validRunner, RunnerSleepValueAfterExec); err != nil {
		return fmt.Errorf("failed to wait: %v", err)
	}

	return nil
}
