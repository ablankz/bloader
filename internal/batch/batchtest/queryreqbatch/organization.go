package queryreqbatch

import (
	"context"
	"fmt"
	"strconv"

	"github.com/LabGroupware/go-measure-tui/internal/api/domain"
	"github.com/LabGroupware/go-measure-tui/internal/api/request/executor"
	"github.com/LabGroupware/go-measure-tui/internal/api/response"
	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/auth"
	"github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
)

type FindOrganizationFactory struct{}

func (f FindOrganizationFactory) Factory(
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

	req := executor.FindOrganizationReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if organizationId, ok = request.PathVariables["organizationId"]; !ok {
		return nil, nil, fmt.Errorf("organizationId not found in pathVariables")
	}
	req.Path.OrganizationID = organizationId

	for key, param := range request.QueryParam {
		switch key {
		case "with":
			req.Param.With = param
		}
	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ResponseDto[domain.OrganizationDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.FindOrganizationReq, response.ResponseDto[domain.OrganizationDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}

type GetOrganizationsFactory struct{}

func (f GetOrganizationsFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	req := executor.GetOrganizationsReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	for key, param := range request.QueryParam {
		switch key {
		case "limit":
			limitInt, err := strconv.Atoi(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert limit to int",
					logger.Value("error", err))
				continue
			}
			req.Param.Limit = limitInt
		case "offset":
			offsetInt, err := strconv.Atoi(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert offset to int",
					logger.Value("error", err))
				continue
			}
			req.Param.Offset = offsetInt
		case "cursor":
			req.Param.Cursor = param[0]
		case "pagination":
			req.Param.Pagination = param[0]
		case "sortField":
			req.Param.SortField = param[0]
		case "sortOrder":
			req.Param.SortOrder = param[0]
		case "withCount":
			withCountBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert withCount to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.WithCount = withCountBool
		case "hasOwnerFilter":
			hasOwnerFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasOwnerFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasOwnerFilter = hasOwnerFilterBool
		case "filterOwnerIDs":
			req.Param.FilterOwnerIDs = param
		case "hasPlanFilter":
			hasPlanFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasPlanFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasPlanFilter = hasPlanFilterBool
		case "filterPlans":
			req.Param.FilterPlans = param
		case "hasUserFilter":
			hasUserFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasUserFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasUserFilter = hasUserFilterBool
		case "filterUserIDs":
			req.Param.FilterUserIDs = param
		case "userFilterType":
			req.Param.UserFilterType = param[0]
		case "with":
			req.Param.With = param
		}

	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ListResponseDto[domain.OrganizationDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.GetOrganizationsReq, response.ListResponseDto[domain.OrganizationDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
