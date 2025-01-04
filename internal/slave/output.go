package slave

import (
	"context"
	"fmt"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/runner"
)

// SlaveOutput represents the slave output service
type SlaveOutput struct {
	// OutputID represents the output ID
	OutputID string
	// outputChan represents the output channel
	outputChan chan<- *pb.CallExecResponse
}

// NewSlaveOutput creates a new SlaveOutput
func NewSlaveOutput(outputID string, outputChan chan<- *pb.CallExecResponse) SlaveOutput {
	return SlaveOutput{
		OutputID:   outputID,
		outputChan: outputChan,
	}
}

// HTTPDataWriteFactory returns the HTTPDataWrite function
func (o SlaveOutput) HTTPDataWriteFactory(
	ctx context.Context,
	log logger.Logger,
	enabled bool,
	uniqueName string,
	header []string,
) (output.HTTPDataWrite, output.Close, error) {

	fmt.Println("HTTPDataWriteFactory", uniqueName)

	select {
	case <-ctx.Done():
		return nil, nil, fmt.Errorf("context canceled")
	case o.outputChan <- &pb.CallExecResponse{
		OutputId:   o.OutputID,
		OutputType: pb.CallExecOutputType_CALL_EXEC_OUTPUT_TYPE_HTTP,
		OutputRoot: uniqueName,
		Output: &pb.CallExecResponse_OutputHttp{
			OutputHttp: &pb.CallExecOutputHTTP{
				Data: header,
			},
		},
	}: // do nothing
	}

	return func(
			ctx context.Context,
			log logger.Logger,
			data []string,
		) error {
			if !enabled {
				return nil
			}
			select {
			case <-ctx.Done():
				return fmt.Errorf("context canceled")
			case o.outputChan <- &pb.CallExecResponse{
				OutputId:   o.OutputID,
				OutputType: pb.CallExecOutputType_CALL_EXEC_OUTPUT_TYPE_HTTP,
				OutputRoot: uniqueName,
				Output: &pb.CallExecResponse_OutputHttp{
					OutputHttp: &pb.CallExecOutputHTTP{
						Data: data,
					},
				},
			}: // do nothing
			}
			return nil
		}, func() error {
			return nil
		}, nil
}

var _ output.Output = SlaveOutput{}

// SlaveOutputFactor represents the factory
type SlaveOutputFactor struct {
	outputChan chan<- *pb.CallExecResponse
}

// Factorize returns the factorized output
func (f *SlaveOutputFactor) Factorize(ctx context.Context, outputID string) (output.Output, error) {
	o := NewSlaveOutput(outputID, f.outputChan)
	return o, nil
}

var _ runner.OutputFactor = &SlaveOutputFactor{}
