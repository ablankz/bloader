package randomstore

import (
	"context"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
)

type RandomStoreValueConfig struct {
	Type string                       `yaml:"type"`
	Data []RandomStoreValueDataConfig `yaml:"data"`
}

type RandomStoreValueRawConfig struct {
	Type string `yaml:"type"`
	Data []any  `yaml:"data"`
}

type RandomStoreValueDataConfig struct {
	Key  string `yaml:"key"`
	Type string `yaml:"type"`
}

type RadomGenerator interface {
	Generate(ctx context.Context, ctr *app.Container, store *sync.Map) error
}
