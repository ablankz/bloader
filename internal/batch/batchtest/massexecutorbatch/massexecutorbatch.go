package massexecutorbatch

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/batchexecmap"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
)

func MassExecuteBatch(ctx context.Context, ctr *app.Container, massExec MassExecute, outputRoot string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	concurrentCount := len(massExec.Data.Requests)

	threadExecutors := make([]*MassiveExecThreadExecutor, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		request := massExec.Data.Requests[i]
		threadExecutors[i] = &MassiveExecThreadExecutor{
			ID: i + 1,
		}

		if massExec.Output.Enabled {
			logFilePath := fmt.Sprintf("%s/massive_execute_%010d.csv", outputRoot, i+1)
			file, err := os.Create(logFilePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			threadExecutors[i].outputFile = file
			defer threadExecutors[i].Close(ctx)

			writer := csv.NewWriter(file)
			header := []string{"Success", "SendDatetime", "ReceivedDatetime", "Count", "ResponseTime", "StatusCode", "Data"}
			if err := writer.Write(header); err != nil {
				return fmt.Errorf("failed to write header: %v", err)
			}
			writer.Flush()
		}

		endType := massExec.Data.Requests[i].EndpointType
		execType := batchexecmap.NewExecTypeFromString(endType)
		if execType == 0 {
			return fmt.Errorf("failed to parse exec type: %s", endType)
		}
		factor := batchexecmap.TypeFactoryMap[execType]
		// INFO: close on factor.Factory(response handler), because only it will write to this channel
		termChan := make(chan execbatch.TerminateType)
		validatedReq := &execbatch.ValidatedExecRequest{}
		if err := execbatch.ValidateExecReq(ctx, ctr, request.ExecRequest, validatedReq); err != nil {
			return fmt.Errorf("failed to validate execute request: %v", err)
		}
		writeFunc := func(
			ctx context.Context,
			ctr *app.Container,
			id int,
			data execbatch.WriteData,
		) error {
			if !massExec.Output.Enabled {
				return nil
			}
			writer := csv.NewWriter(threadExecutors[i].outputFile)
			ctr.Logger.Debug(ctx, "Writing data to csv",
				logger.Value("id", id), logger.Value("data", data), logger.Value("on", "runAsyncProcessing"))
			if err := writer.Write(data.ToSlice()); err != nil {
				ctr.Logger.Error(ctx, "failed to write data to csv",
					logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
			}
			writer.Flush()
			return nil
		}
		executor, _, err := factor.Factory(
			ctx,
			ctr,
			i+1,
			validatedReq,
			termChan,
			ctr.AuthToken,
			ctr.Config.Web.API.Url,
			writeFunc,
		)
		ctr.Logger.Info(ctx, "created executor",
			logger.Value("id", i+1), logger.Value("type", endType), logger.Value("executor", executor))
		if err != nil {
			return fmt.Errorf("failed to create executor: %v", err)
		}
		threadExecutors[i].RequestExecutor = executor
		threadExecutors[i].TermChan = termChan
		threadExecutors[i].successBreak = request.SuccessBreak
	}

	var wg sync.WaitGroup
	atomicErr := atomic.Value{}
	var startChan = make(chan struct{})
	for _, executor := range threadExecutors {
		wg.Add(1)
		go func(exec *MassiveExecThreadExecutor) {
			defer wg.Done()
			if err := exec.Execute(ctx, ctr, startChan); err != nil {
				atomicErr.Store(err)
				ctr.Logger.Error(ctx, "failed to execute",
					logger.Value("error", err), logger.Value("id", exec.ID))
				cancel()
				return
			}
		}(executor)
	}
	close(startChan)
	wg.Wait()

	if err := atomicErr.Load(); err != nil {
		ctr.Logger.Error(ctx, "failed to find error",
			logger.Value("error", err.(error)), logger.Value("on", "MassExecuteBatch"))
		return err.(error)
	}

	return nil
}
