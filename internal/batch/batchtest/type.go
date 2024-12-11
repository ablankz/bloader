package batchtest

import (
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/metricsbatch"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/prefetchbatch"
)

type BatchTestType struct {
	Type     string                          `yaml:"type"`
	Sleep    BatchSleep                      `yaml:"sleep"`
	Prefetch prefetchbatch.PrefetchConfig    `yaml:"prefetch"`
	Metrics  metricsbatch.MetricsBatchConfig `yaml:"metrics"`
	Data     any                             `yaml:"data"`
	Output   BatchTestOutput                 `yaml:"output"`
}

type BatchTestOutput struct {
	Enabled bool `yaml:"enabled"`
}

type BatchSleep struct {
	Enabled bool              `yaml:"enabled"`
	Values  []BatchSleepValue `yaml:"values"`
}

type BatchSleepValue struct {
	Duration string
	After    string // init, prefetch, replace, metricsBoot, request
}

type SleepAfterValue string

const (
	SleepAfterInit        SleepAfterValue = "init"
	SleepAfterPrefetch    SleepAfterValue = "prefetch"
	SleepAfterReplace     SleepAfterValue = "replace"
	SleepAfterMetrics     SleepAfterValue = "metricsBoot"
	SleepAfterSuccessExec SleepAfterValue = "successExec"
	SleepAfterFailedExec  SleepAfterValue = "failedExec"
)

func (b BatchTestType) RetrieveSleepValue(after SleepAfterValue) (time.Duration, bool) {
	for _, v := range b.Sleep.Values {
		if v.After == string(after) {
			d, err := time.ParseDuration(v.Duration)
			if err != nil {
				return time.Duration(0), false
			}
			return d, true
		}
	}
	return time.Duration(0), false
}
