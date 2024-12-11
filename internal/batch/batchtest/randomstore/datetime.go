package randomstore

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"gopkg.in/yaml.v3"
)

type RandomDatetimeValueGenerator struct {
	Key    string
	Format string
}

func (p *RandomDatetimeValueGenerator) Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error {
	startTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	endTime := time.Date(2030, 12, 31, 23, 59, 59, 0, time.UTC)

	duration := endTime.Sub(startTime)
	randomDuration := time.Duration(rand.N(int64(duration)))

	randomTime := startTime.Add(randomDuration)
	store.Store(p.Key, randomTime.Format(p.Format))

	return nil
}

type RandomStoreValueDatetimeDataConfig struct {
	Key    string `yaml:"key"`
	Type   string `yaml:"type"`
	Format string `yaml:"format"`
}

func (p *RandomStoreValueDatetimeDataConfig) Init(conf io.Reader) error {
	decoder := yaml.NewDecoder(conf)
	if err := decoder.Decode(p); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}
	return nil
}

func (p *RandomStoreValueDatetimeDataConfig) GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error) {
	return &RandomDatetimeValueGenerator{
		Key:    p.Key,
		Format: p.Format,
	}, nil
}
