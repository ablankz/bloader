package massexecutorbatch

import (
	"context"
	"fmt"
	"os"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
)

type MassiveExecThreadExecutor struct {
	ID              int
	outputFile      *os.File
	RequestExecutor executor.RequestExecutor
	TermChan        chan execbatch.TerminateType
	successBreak    []string
}

func NewMassiveRequestThreadExecutor(
	id int,
	outputFile *os.File,
	req executor.RequestExecutor,
) *MassiveExecThreadExecutor {
	return &MassiveExecThreadExecutor{
		ID:              id,
		outputFile:      outputFile,
		RequestExecutor: req,
	}
}

func (e *MassiveExecThreadExecutor) Execute(
	ctx context.Context,
	ctr *app.Container,
	startChan <-chan struct{},
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	<-startChan

	ctr.Logger.Info(ctx, "Execute Start",
		logger.Value("ExecutorID", e.ID))
	err := e.RequestExecutor.RequestExecute(ctx, ctr)
	if err != nil {
		return err
	}

	termType := <-e.TermChan
	ctr.Logger.Info(ctx, "Execute End For Break",
		logger.Value("ExecuteID", e.ID))
	for _, breakType := range e.successBreak {
		if termType == execbatch.NewTerminateTypeFromString(breakType) {
			ctr.Logger.Info(ctx, "Execute End For Success Break", logger.Value("ExecuteID", e.ID))
			return nil
		}
	}
	return fmt.Errorf("execute End For Fail Break: %s", termType.String())
}

func (e *MassiveExecThreadExecutor) Close(ctx context.Context) error {
	e.outputFile.Close()
	return nil
}
