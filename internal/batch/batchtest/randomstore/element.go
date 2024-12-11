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

type RandomElementValueGenerator struct {
	Key    string
	Values []interface{}
}

func (p *RandomElementValueGenerator) Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error {
	if len(p.Values) == 0 {
		return fmt.Errorf("values is empty")
	}

	value := p.Values[rand.N(len(p.Values))]
	store.Store(p.Key, fmt.Sprintf("%v", value))
	return nil
}

type RandomStoreValueElementDataConfig struct {
	Key   string `yaml:"key"`
	Type  string `yaml:"type"`
	Value []any  `yaml:"value"`
}

func (p *RandomStoreValueElementDataConfig) Init(conf io.Reader) error {
	decoder := yaml.NewDecoder(conf)
	if err := decoder.Decode(p); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}
	return nil
}

func (p *RandomStoreValueElementDataConfig) GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error) {
	return &RandomElementValueGenerator{
		Key:    p.Key,
		Values: p.Value,
	}, nil
}
