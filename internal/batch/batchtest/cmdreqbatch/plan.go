package cmdreqbatch

import (
	"context"
	"fmt"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/api/response"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/auth"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
)

type CreateTaskFactory struct{}

func (f CreateTaskFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	req := executor.CreateTaskReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	req.Body = request.Body

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.CommandResponseDto])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.CreateTaskReq, response.CommandResponseDto]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}

type UpdateStatusTaskFactory struct{}

func (f UpdateStatusTaskFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	var ok bool
	var taskId string

	req := executor.UpdateStatusTaskReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if taskId, ok = request.PathVariables["taskId"]; !ok {
		return nil, nil, fmt.Errorf("taskId not found in pathVariables")
	}
	req.Path.TaskID = taskId

	req.Body = request.Body

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.CommandResponseDto])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.UpdateStatusTaskReq, response.CommandResponseDto]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
