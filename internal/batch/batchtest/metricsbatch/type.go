package metricsbatch

import (
	"context"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/app"
)

type BatchConfigWithRawMetrics struct {
	Metrics MetricsBatchRawConfig `yaml:"metrics"`
}

type MetricsBatchRawConfig struct {
	Enabled  bool  `yaml:"enabled"`
	Requests []any `yaml:"requests"`
}

type MetricsBatchConfig struct {
	Enabled  bool                        `yaml:"enabled"`
	Requests []MetricsBatchRequestConfig `yaml:"requests"`
}

type MetricsBatchRequestConfig struct {
	ID   string                          `yaml:"id"`
	Type string                          `yaml:"type"`
	Data []MetricsBatchRequestDataConfig `yaml:"data"`
}

type MetricsBatchRequestDataConfig struct {
	Key      string `yaml:"key"`
	JMESPath string `yaml:"jmesPath"`
	OnNil    string `yaml:"onNil"`
}

type MetricsType string

const (
	MetricsTypePrometheus MetricsType = "prometheus"
)

func NewMetricsTypeFromStr(s string) MetricsType {
	return MetricsType(s)
}

type MetricsWriter func(
	ctx context.Context,
	ctr *app.Container,
	data executor.ResponseContent[any],
) error

type MetricsFetcher interface {
	Fetch(ctx context.Context, ctr *app.Container, writer MetricsWriter) (<-chan TermType, error)
}

type TermType int

const (
	_ TermType = iota
	TermTypeContext
	TermWriteError
)
