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
	// RunnerKindMemoryValue represents the memory store value runner
	RunnerKindMemoryValue RunnerKind = "MemoryValue"
	// RunnerKindStoreImport represents the store import runner
	RunnerKindStoreImport RunnerKind = "StoreImport"
	// RunnerKindOneExecute represents execute one request runner
	RunnerKindOneExecute RunnerKind = "OneExecute"
	// RunnerKindMassExecute represents execute multiple requests runner
	RunnerKindMassExecute RunnerKind = "MassExecute"
	// RunnerKindFlow represents the flow runner
	RunnerKindFlow RunnerKind = "Flow"
)

// Runner represents a runner
type Runner struct {
	Kind        *string           `yaml:"kind"`
	Sleep       RunnerSleep       `yaml:"sleep"`
	StoreImport RunnerStoreImport `yaml:"store_import"`
}

// ValidRunner represents a valid runner
type ValidRunner struct {
	Kind        RunnerKind
	Sleep       ValidRunnerSleep
	StoreImport ValidRunnerStoreImport
}

// Validate validates a runner
func (r Runner) Validate() (ValidRunner, error) {
	if r.Kind == nil {
		return ValidRunner{}, fmt.Errorf("kind is required")
	}
	var kind RunnerKind
	switch RunnerKind(*r.Kind) {
	case RunnerKindStoreValue,
		RunnerKindMemoryValue,
		RunnerKindStoreImport,
		RunnerKindOneExecute,
		RunnerKindMassExecute,
		RunnerKindFlow:
		kind = RunnerKind(*r.Kind)
	default:
		return ValidRunner{}, fmt.Errorf("invalid kind value: %s", *r.Kind)
	}
	validSleep, err := r.Sleep.Validate()
	if err != nil {
		return ValidRunner{}, fmt.Errorf("failed to validate sleep: %v", err)
	}
	validStoreImport, err := r.StoreImport.Validate()
	if err != nil {
		return ValidRunner{}, fmt.Errorf("failed to validate store import: %v", err)
	}
	return ValidRunner{
		Kind:        kind,
		Sleep:       validSleep,
		StoreImport: validStoreImport,
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
	// RunnerSleepValueAfterExec represents the success after value for a runner sleep value
	RunnerSleepValueAfterExec RunnerSleepValueAfter = "exec"
	// RunnerSleepValueAfterFailedExec represents the failed after value for a runner sleep value
	RunnerSleepValueAfterFailedExec RunnerSleepValueAfter = "failedExec"
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
	case RunnerSleepValueAfterInit, RunnerSleepValueAfterExec, RunnerSleepValueAfterFailedExec:
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
func (r ValidRunner) RetrieveSleepValue(after RunnerSleepValueAfter) (time.Duration, bool) {
	for _, v := range r.Sleep.Values {
		if v.After == after {
			return v.Duration, true
		}
	}
	return time.Duration(0), false
}

// RunnerStoreImport represents the StoreImport runner
type RunnerStoreImport struct {
	Enabled bool              `yaml:"enabled"`
	Data    []StoreImportData `yaml:"data"`
}

// ValidRunnerStoreImport represents the valid RunnerStoreImport runner
type ValidRunnerStoreImport struct {
	Enabled bool
	Data    []ValidStoreImportData
}

// Validate validates the RunnerStoreImport
func (r RunnerStoreImport) Validate() (ValidRunnerStoreImport, error) {
	if !r.Enabled {
		return ValidRunnerStoreImport{}, nil
	}
	var validData []ValidStoreImportData
	for i, d := range r.Data {
		valid, err := d.Validate()
		if err != nil {
			return ValidRunnerStoreImport{}, fmt.Errorf("failed to validate data at index %d: %v", i, err)
		}
		validData = append(validData, valid)
	}
	return ValidRunnerStoreImport{
		Enabled: r.Enabled,
		Data:    validData,
	}, nil
}

// CredentialEncryptConfig is the configuration for the credential encrypt.
type CredentialEncryptConfig struct {
	Enabled   bool    `yaml:"enabled"`
	EncryptID *string `yaml:"encrypt_id"`
}

// ValidCredentialEncryptConfig represents the valid auth credential encrypt configuration
type ValidCredentialEncryptConfig struct {
	Enabled   bool
	EncryptID string
}

// Validate validates the credential encrypt configuration
func (c CredentialEncryptConfig) Validate() (ValidCredentialEncryptConfig, error) {
	if !c.Enabled {
		return ValidCredentialEncryptConfig{}, nil
	}
	if c.EncryptID == nil {
		return ValidCredentialEncryptConfig{}, fmt.Errorf("encrypt_id is required")
	}
	return ValidCredentialEncryptConfig{
		Enabled:   c.Enabled,
		EncryptID: *c.EncryptID,
	}, nil
}
