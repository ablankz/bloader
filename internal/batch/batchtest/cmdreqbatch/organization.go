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

type CreateOrganizationFactory struct{}

func (f CreateOrganizationFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	req := executor.CreateOrganizationReq{
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

	return executor.RequestContent[executor.CreateOrganizationReq, response.CommandResponseDto]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}

type AddUsersOrganizationFactory struct{}

func (f AddUsersOrganizationFactory) Factory(
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
	var organizationId string

	req := executor.AddUsersOrganizationReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if organizationId, ok = request.PathVariables["organizationId"]; !ok {
		return nil, nil, fmt.Errorf("organizationId not found in pathVariables")
	}
	req.Path.OrganizationID = organizationId

	req.Body = request.Body

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.CommandResponseDto])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.AddUsersOrganizationReq, response.CommandResponseDto]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
