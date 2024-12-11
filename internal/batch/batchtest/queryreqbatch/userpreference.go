package queryreqbatch

import (
	"context"
	"fmt"

	"github.com/LabGroupware/go-measure-tui/internal/api/domain"
	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/api/response"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/auth"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
)

type FindUserPreferenceFactory struct{}

func (f FindUserPreferenceFactory) Factory(
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
	var userPreferenceId string

	req := executor.FindUserPreferenceReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if userPreferenceId, ok = request.PathVariables["userPreferenceId"]; !ok {
		return nil, nil, fmt.Errorf("userPreferenceId not found in pathVariables")
	}
	req.Path.UserPreferenceID = userPreferenceId

	for key, param := range request.QueryParam {
		switch key {
		case "with":
			req.Param.With = param
		}
	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ResponseDto[domain.UserPreferenceDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.FindUserPreferenceReq, response.ResponseDto[domain.UserPreferenceDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
