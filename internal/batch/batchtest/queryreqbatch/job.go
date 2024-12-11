package queryreqbatch

import (
	"context"
	"fmt"

	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/api/response"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/auth"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
)

type FindJobFactory struct{}

func (f FindJobFactory) Factory(
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
	var jobId string

	req := executor.GetJobReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if jobId, ok = request.PathVariables["jobId"]; !ok {
		return nil, nil, fmt.Errorf("jobId not found in pathVariables")
	}
	req.Path.JobID = jobId

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.JobResponseDto])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.GetJobReq, response.JobResponseDto]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
