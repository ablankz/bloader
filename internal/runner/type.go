package runner

import (
	"fmt"
	"time"

	"github.com/jmespath/go-jmespath"
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
	Enabled bool                    `yaml:"enabled"`
	Data    []RunnerStoreImportData `yaml:"data"`
}

// ValidRunnerStoreImport represents the valid RunnerStoreImport runner
type ValidRunnerStoreImport struct {
	Enabled bool
	Data    []ValidRunnerStoreImportData
}

// Validate validates the RunnerStoreImport
func (r RunnerStoreImport) Validate() (ValidRunnerStoreImport, error) {
	if !r.Enabled {
		return ValidRunnerStoreImport{}, nil
	}
	var validData []ValidRunnerStoreImportData
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

// RunnerStoreImportData represents the data for the StoreImport runner
type RunnerStoreImportData struct {
	BucketID   *string                 `yaml:"bucket_id"`
	Key        *string                 `yaml:"key"`
	StoreKey   *string                 `yaml:"store_key"`
	ThreadOnly bool                    `yaml:"thread_only"`
	Encrypt    CredentialEncryptConfig `yaml:"encrypt"`
}

// ValidRunnerStoreImportData represents the valid data for the StoreImport runner
type ValidRunnerStoreImportData struct {
	BucketID   string
	Key        string
	StoreKey   string
	ThreadOnly bool
	Encrypt    ValidCredentialEncryptConfig
}

// Validate validates the StoreImportData
func (d RunnerStoreImportData) Validate() (ValidRunnerStoreImportData, error) {
	if d.BucketID == nil {
		return ValidRunnerStoreImportData{}, fmt.Errorf("bucket_id is required")
	}
	if d.Key == nil {
		return ValidRunnerStoreImportData{}, fmt.Errorf("key is required")
	}
	if d.StoreKey == nil {
		return ValidRunnerStoreImportData{}, fmt.Errorf("store_key is required")
	}
	validEncrypt, err := d.Encrypt.Validate()
	if err != nil {
		return ValidRunnerStoreImportData{}, fmt.Errorf("failed to validate encrypt: %v", err)
	}
	return ValidRunnerStoreImportData{
		BucketID:   *d.BucketID,
		Key:        *d.Key,
		StoreKey:   *d.StoreKey,
		ThreadOnly: d.ThreadOnly,
		Encrypt:    validEncrypt,
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

// DataExtractorOnNilType represents the on nil type for the data extractor
type DataExtractorOnNilType string

const (
	// DataExtractorOnNilTypeEmpty represents the empty on nil type for the data extractor
	DataExtractorOnNilTypeEmpty DataExtractorOnNilType = "empty"
	// DataExtractorOnNilTypeNull represents the null on nil type for the data extractor
	DataExtractorOnNilTypeNull DataExtractorOnNilType = "null"
	// DataExtractorOnNilTypeError represents the error on nil type for the data extractor
	DataExtractorOnNilTypeError DataExtractorOnNilType = "error"

	// DefaultDataExtractorOnNilType represents the default on nil type for the data extractor
	DefaultDataExtractorOnNilType DataExtractorOnNilType = DataExtractorOnNilTypeNull
)

// DataExtractorType represents the type for the data extractor
type DataExtractorType string

const (
	// DataExtractorTypeJMESPath represents the JMESPath type for the data extractor
	DataExtractorTypeJMESPath DataExtractorType = "jmesPath"
)

// DataExtractor represents the data extractor for the OneExec runner
type DataExtractor struct {
	Type     *string `yaml:"type"`
	JMESPath *string `yaml:"jmesPath"`
	OnNil    *string `yaml:"onNil"`
}

// ValidDataExtractor represents the valid data extractor for the OneExec runner
type ValidDataExtractor struct {
	Type     DataExtractorType
	JMESPath *jmespath.JMESPath
	OnNil    DataExtractorOnNilType
}

// Validate validates the data extractor
func (d DataExtractor) Validate() (ValidDataExtractor, error) {
	if d.Type == nil {
		return ValidDataExtractor{}, fmt.Errorf("type is required")
	}
	var valid ValidDataExtractor
	switch DataExtractorType(*d.Type) {
	case DataExtractorTypeJMESPath:
		valid.Type = DataExtractorType(*d.Type)
		if d.JMESPath == nil {
			return ValidDataExtractor{}, fmt.Errorf("jmesPath is required")
		}
		jPath, err := jmespath.Compile(*d.JMESPath)
		if err != nil {
			return ValidDataExtractor{}, fmt.Errorf("failed to compile jmesPath: %v", err)
		}
		valid.JMESPath = jPath
		if d.OnNil == nil {
			valid.OnNil = DefaultDataExtractorOnNilType
		} else {
			switch DataExtractorOnNilType(*d.OnNil) {
			case DataExtractorOnNilTypeEmpty, DataExtractorOnNilTypeNull, DataExtractorOnNilTypeError:
				valid.OnNil = DataExtractorOnNilType(*d.OnNil)
			default:
				valid.OnNil = DefaultDataExtractorOnNilType
			}
		}
	default:
		return ValidDataExtractor{}, fmt.Errorf("invalid type value: %s", *d.Type)
	}
	return valid, nil
}

// Extract extracts the data from the response
func (d ValidDataExtractor) Extract(data any) (any, error) {
	switch d.Type {
	case DataExtractorTypeJMESPath:
		result, err := d.JMESPath.Search(data)
		if err != nil {
			return nil, fmt.Errorf("failed to search jmesPath: %v", err)
		}
		if result == nil {
			switch d.OnNil {
			case DataExtractorOnNilTypeEmpty:
				return "", nil
			case DataExtractorOnNilTypeNull:
				return nil, nil
			case DataExtractorOnNilTypeError:
				return nil, fmt.Errorf("nil value")
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported data extractor type: %s", d.Type)
	}
}
