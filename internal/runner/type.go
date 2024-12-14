package runner

import (
	"fmt"
	"time"
)

// RunnerKind represents the kind of runner
type RunnerKind string

const (
	// RunnerKindStoreValue represents the store value runner
	RunnerKindStoreValue RunnerKind = "StoreValue"
)

// Runner represents a runner
type Runner struct {
	Kind  *string     `yaml:"kind"`
	Sleep RunnerSleep `yaml:"sleep"`
}

// ValidRunner represents a valid runner
type ValidRunner struct {
	Kind  RunnerKind
	Sleep ValidRunnerSleep
}

// Validate validates a runner
func (r Runner) Validate() (ValidRunner, error) {
	if r.Kind == nil {
		return ValidRunner{}, fmt.Errorf("kind is required")
	}
	var kind RunnerKind
	switch RunnerKind(*r.Kind) {
	case RunnerKindStoreValue:
		kind = RunnerKind(*r.Kind)
	default:
		return ValidRunner{}, fmt.Errorf("invalid kind value: %s", *r.Kind)
	}
	validSleep, err := r.Sleep.Validate()
	if err != nil {
		return ValidRunner{}, fmt.Errorf("failed to validate sleep: %v", err)
	}
	return ValidRunner{
		Kind:  kind,
		Sleep: validSleep,
	}, nil
}

// RunnerSleep represents the sleep configuration for a runner
type RunnerSleep struct {
	Enabled bool               `yaml:"enabled"`
	Values  []RunnerSleepValue `yaml:"values"`
}

// ValidRunnerSleep represents a valid runner sleep configuration
type ValidRunnerSleep struct {
	Enabled bool
	Values  []ValidRunnerSleepValue
}

// Validate validates a runnerSleep
func (r RunnerSleep) Validate() (ValidRunnerSleep, error) {
	if !r.Enabled {
		return ValidRunnerSleep{}, nil
	}
	var values []ValidRunnerSleepValue
	for _, v := range r.Values {
		valid, err := v.Validate()
		if err != nil {
			return ValidRunnerSleep{}, fmt.Errorf("failed to validate sleep value: %v", err)
		}
		values = append(values, valid)
	}
	return ValidRunnerSleep{
		Enabled: r.Enabled,
		Values:  values,
	}, nil
}

// RunnerSleepValueAfter represents the after value for a runner sleep value
type RunnerSleepValueAfter string

const (
	// RunnerSleepValueAfterInit represents the init after value for a runner sleep value
	RunnerSleepValueAfterInit RunnerSleepValueAfter = "init"
	// RunnerSleepValueAfterReplace represents the replace after value for a runner sleep value
	RunnerSleepValueAfterReplace RunnerSleepValueAfter = "replace"
	// RunnerSleepValueAfterMetricsBoot represents the metricsBoot after value for a runner sleep value
	RunnerSleepValueAfterMetricsBoot RunnerSleepValueAfter = "metricsBoot"
	// RunnerSleepValueAfterRequest represents the request after value for a runner sleep value
	RunnerSleepValueAfterRequest RunnerSleepValueAfter = "request"
)

// RunnerSleepValue represents the sleep value for a runner
type RunnerSleepValue struct {
	Duration *string
	After    *string
}

// ValidRunnerSleepValue represents a valid runner sleep value
type ValidRunnerSleepValue struct {
	Duration time.Duration
	After    RunnerSleepValueAfter
}

// Validate validates a runner
func (r RunnerSleepValue) Validate() (ValidRunnerSleepValue, error) {
	if r.Duration == nil {
		return ValidRunnerSleepValue{}, fmt.Errorf("duration is required")
	}
	if r.After == nil {
		return ValidRunnerSleepValue{}, fmt.Errorf("after is required")
	}
	d, err := time.ParseDuration(*r.Duration)
	if err != nil {
		return ValidRunnerSleepValue{}, fmt.Errorf("failed to parse duration: %v", err)
	}
	var after RunnerSleepValueAfter
	switch RunnerSleepValueAfter(*r.After) {
	case RunnerSleepValueAfterInit, RunnerSleepValueAfterReplace, RunnerSleepValueAfterMetricsBoot, RunnerSleepValueAfterRequest:
		after = RunnerSleepValueAfter(*r.After)
	default:
		return ValidRunnerSleepValue{}, fmt.Errorf("invalid after value: %s", *r.After)
	}
	return ValidRunnerSleepValue{
		Duration: d,
		After:    after,
	}, nil
}

// RetrieveSleepValue retrieves the sleep value for a runner
// func (r Runner) RetrieveSleepValue(after RunnerSleepValueAfter) (time.Duration, bool) {
// 	for _, v := range r.Sleep.Values {
// 		if v.After != nil && *v.After == string(after) {
// 			return v.Duration, true
// 		}
// 	}
// 	return time.Duration(0), false
// }
