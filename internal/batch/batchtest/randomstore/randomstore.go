package randomstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"gopkg.in/yaml.v3"
)

func RandomStoreValueBatch(
	ctx context.Context,
	ctr *app.Container,
	conf RandomStoreValueConfig,
	rowConf io.Reader,
	store *sync.Map,
) (map[string]string, error) {
	newStore := sync.Map{}

	var rawData RandomStoreValueRawConfig
	decoder := yaml.NewDecoder(rowConf)
	if err := decoder.Decode(&rawData); err != nil {
		return nil, fmt.Errorf("failed to decode yaml: %v", err)
	}

	for i, data := range conf.Data {
		factor := GetRandomGeneratorFactory(data.Type)
		if factor == nil {
			ctr.Logger.Error(ctx, "failed to find error")
			return nil, fmt.Errorf("failed to find generator factory: %s", data.Type)
		}
		var buf bytes.Buffer
		if err := yaml.NewEncoder(&buf).Encode(rawData.Data[i]); err != nil {
			return nil, fmt.Errorf("failed to encode yaml: %v", err)
		}

		if err := factor.Init(&buf); err != nil {
			ctr.Logger.Error(ctx, "failed to find error",
				logger.Value("error", err))
			return nil, fmt.Errorf("failed to init factory: %v", err)
		}

		generator, err := factor.GeneratorFactory(ctx, ctr)

		if err != nil {
			ctr.Logger.Error(ctx, "failed to find error",
				logger.Value("error", err))
			return nil, fmt.Errorf("failed to create generator: %v", err)
		}

		if err := generator.Generate(ctx, ctr, &newStore); err != nil {
			ctr.Logger.Error(ctx, "failed to find error",
				logger.Value("error", err))
			return nil, fmt.Errorf("failed to generate: %v", err)
		}
	}

	newMap := make(map[string]string)

	newStore.Range(func(key, value interface{}) bool {
		store.Store(key, value)
		newMap[key.(string)] = value.(string)
		return true
	})

	return newMap, nil
}
