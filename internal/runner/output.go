package runner

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/output"
)

// OutputFactor represents the output factor
type OutputFactor interface {
	// Factorize returns the factorized output
	Factorize(ctx context.Context, outputID string) (output.Output, error)
}

// LocalOutputFactor represents the local output factor
type LocalOutputFactor struct {
	outputCtr output.OutputContainer
}

// NewLocalOutputFactor creates a new LocalOutputFactor
func NewLocalOutputFactor(outputCtr output.OutputContainer) LocalOutputFactor {
	return LocalOutputFactor{
		outputCtr: outputCtr,
	}
}

// Factorize returns the factorized output
func (f LocalOutputFactor) Factorize(ctx context.Context, outputID string) (output.Output, error) {
	o, ok := f.outputCtr[outputID]
	if !ok {
		return nil, fmt.Errorf("output not found: %s", outputID)
	}
	return o, nil
}
