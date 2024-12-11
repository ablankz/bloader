package randomstore

import (
	"context"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"gopkg.in/yaml.v3"
)

type RandomFloatValueGenerator struct {
	Key       string
	From      float64
	To        float64
	Precision int
}

func (p *RandomFloatValueGenerator) Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error {

	n := rand.Float64()
	n = n*(p.To-p.From) + p.From

	format := fmt.Sprintf("%%.%df", p.Precision)
	value := fmt.Sprintf(format, n)
	store.Store(p.Key, value)
	return nil
}

type RandomStoreValueFloatDataConfig struct {
	Key       string  `yaml:"key"`
	Type      string  `yaml:"type"`
	From      float64 `yaml:"from"`
	To        float64 `yaml:"to"`
	Precision int     `yaml:"precision"`
}

func (p *RandomStoreValueFloatDataConfig) Init(conf io.Reader) error {
	decoder := yaml.NewDecoder(conf)
	if err := decoder.Decode(p); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}
	return nil
}

func (p *RandomStoreValueFloatDataConfig) GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error) {
	return &RandomFloatValueGenerator{
		Key:       p.Key,
		From:      p.From,
		To:        p.To,
		Precision: p.Precision,
	}, nil
}
