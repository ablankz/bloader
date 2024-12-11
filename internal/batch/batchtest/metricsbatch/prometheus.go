package metricsbatch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"gopkg.in/yaml.v3"
)

type PrometheusMetricsFetcher struct {
	req      *http.Request
	interval time.Duration
}

func (r PrometheusMetricsFetcher) CreateRequest(ctx context.Context, ctr *app.Container) (*http.Request, error) {
	return r.req, nil
}

func (p *PrometheusMetricsFetcher) Fetch(
	ctx context.Context,
	ctr *app.Container,
	writer MetricsWriter,
) (<-chan TermType, error) {

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[any])
	termChan := make(chan TermType)

	req := executor.RequestContent[PrometheusMetricsFetcher, any]{
		Req:          *p,
		Interval:     p.interval,
		ResponseWait: false,
		ResChan:      resChan,
		CountLimit:   executor.RequestCountLimit{},
	}

	err := req.RequestExecute(ctx, ctr)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	go func() {
		defer close(termChan)

		for {
			select {
			case <-ctx.Done():
				ctr.Logger.Info(ctx, "Prometheus Query End For Term",
					logger.Value("on", "PrometheusMetricsFetcher.Fetch"))
				termChan <- TermTypeContext
				return
			case res := <-resChan:
				err := writer(ctx, ctr, res)
				if err != nil {
					ctr.Logger.Error(ctx, "failed to write metrics",
						logger.Value("error", err), logger.Value("on", "PrometheusMetricsFetcher.Fetch"))
					termChan <- TermWriteError
					return
				}
			}
		}
	}()

	return termChan, nil
}

type PrometheusMetricsBatchRequestConfig struct {
	ID       string                          `yaml:"id"`
	Type     string                          `yaml:"type"`
	URL      string                          `yaml:"url"`
	Query    string                          `yaml:"query"`
	Interval string                          `yaml:"interval"`
	Data     []MetricsBatchRequestDataConfig `yaml:"data"`
}

func (p *PrometheusMetricsBatchRequestConfig) Init(conf []byte) error {
	err := yaml.Unmarshal(conf, p)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %w", err)
	}
	return nil
}

func (p *PrometheusMetricsBatchRequestConfig) FetcherFactory(ctx context.Context, ctr *app.Container) (MetricsFetcher, error) {
	baseURL := p.URL
	fullURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	queryParams := fullURL.Query()
	queryParams.Add("query", p.Query)
	fullURL.RawQuery = queryParams.Encode()

	ctr.Logger.Debug(ctx, "GET request to Prometheus query URL created",
		logger.Value("url", fullURL.String()), logger.Value("on", "PrometheusMetricsBatchRequestConfig.FetcherFactory"))

	req, err := http.NewRequest(http.MethodGet, fullURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	var interval time.Duration

	if p.Interval != "" {
		interval, err = time.ParseDuration(p.Interval)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interval: %w", err)
		}
	} else {
		return nil, fmt.Errorf("interval is required")
	}

	return &PrometheusMetricsFetcher{
		req:      req,
		interval: interval,
	}, nil
}

var _ MetricsFetcher = &PrometheusMetricsFetcher{}

var _ MetricsFetcherFactory = &PrometheusMetricsBatchRequestConfig{}
