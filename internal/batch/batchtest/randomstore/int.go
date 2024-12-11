package randomstore

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"gopkg.in/yaml.v3"
)

type RandomIntValueGenerator struct {
	Key  string
	From int
	To   int
}

func (p *RandomIntValueGenerator) Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error {
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return fmt.Errorf("failed to generate random number: %w", err)
	}
	value := int(n.Int64())
	store.Store(p.Key, fmt.Sprintf("%d", value))
	return nil
}

type RandomStoreValueIntDataConfig struct {
	Key  string `yaml:"key"`
	Type string `yaml:"type"`
	From int    `yaml:"from"`
	To   int    `yaml:"to"`
}

func (p *RandomStoreValueIntDataConfig) Init(conf io.Reader) error {
	decoder := yaml.NewDecoder(conf)
	if err := decoder.Decode(p); err != nil {
		return fmt.Errorf("failed to decode yaml: %w", err)
	}
	return nil
}

func (p *RandomStoreValueIntDataConfig) GeneratorFactory(ctx context.Context, ctr *app.Container) (RadomGenerator, error) {
	return &RandomIntValueGenerator{
		Key:  p.Key,
		From: p.From,
		To:   p.To,
	}, nil
}
