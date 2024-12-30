package output

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/utils"
)

// GrpcOutput represents the local output service
type GrpcOutput struct {
	// BaseID of the output
	BaseID string
}

// NewGrpcOutput creates a new GrpcOutput
func NewGrpcOutput(baseID string) GrpcOutput {
	return GrpcOutput{
		BaseID: baseID,
	}
}

// HTTPDataWriteFactory returns the HTTPDataWrite function
func (o GrpcOutput) HTTPDataWriteFactory(
	ctx context.Context,
	ctr *container.Container,
	enabled bool,
	uniqueName string,
	header []string,
) (HTTPDataWrite, Close, error) {
	var filePath string
	f, err := utils.CreateFileWithDir(filePath)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to create file",
			logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
		return nil, nil, fmt.Errorf("failed to create file: %w", err)
	}

	return func(
			ctx context.Context,
			ctr *container.Container,
			data []string,
		) error {
			if !enabled {
				return nil
			}
			return nil
		}, func() error {
			return f.Close()
		}, nil
}

var _ Output = GrpcOutput{}
