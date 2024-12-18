package output

import (
	"context"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
)

// HTTPDataWrite writes the data to the output
type HTTPDataWrite func(ctx context.Context, ctr *container.Container, data []string) error

// Output represents a output to be scanned
type Output interface {
	// HTTPDataWriteFactory returns the HTTPDataWrite function
	HTTPDataWriteFactory(
		ctx context.Context,
		ctr *container.Container,
		enabled bool,
		uniqueName string,
		header []string,
	) (HTTPDataWrite, Close, error)
}

// OutputContainer is a map of outputs
type OutputContainer map[string]Output

// NewOutputContainer creates a new OutputContainer
func NewOutputContainer(env string, cfg config.ValidOutputConfig) OutputContainer {
	outputs := make(OutputContainer)
	for _, output := range cfg {
		var t Output
		var ok bool
		for _, val := range output.Values {
			if val.Env == env {
				switch val.Type {
				case config.OutputTypeLocal:
					t = NewLocalOutput(val)
				}
				ok = true
				break
			}
		}
		if !ok {
			continue
		}
		outputs[output.ID] = t
	}
	return outputs
}
