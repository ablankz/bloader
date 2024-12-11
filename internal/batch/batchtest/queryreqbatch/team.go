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

type FindTeamFactory struct{}

func (f FindTeamFactory) Factory(
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
	var teamId string

	req := executor.FindTeamReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if teamId, ok = request.PathVariables["teamId"]; !ok {
		return nil, nil, fmt.Errorf("teamId not found in pathVariables")
	}
	req.Path.TeamID = teamId

	for key, param := range request.QueryParam {
		switch key {
		case "with":
			req.Param.With = param
		}
	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ResponseDto[domain.TeamDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.FindTeamReq, response.ResponseDto[domain.TeamDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}

type GetTeamsFactory struct{}

func (f GetTeamsFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	req := executor.GetTeamsReq{
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
		case "hasIsDefaultFilter":
			hasIsDefaultFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasIsDefaultFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasIsDefaultFilter = hasIsDefaultFilterBool
		case "filterIsDefault":
			filterIsDefault, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert filterIsDefault to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.FilterIsDefault = filterIsDefault
		case "hasOrganizationFilter":
			hasOrganizationFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasOrganizationFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasOrganizationFilter = hasOrganizationFilterBool
		case "filterOrganizationIDs":
			req.Param.FilterOrganizationIDs = param
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
	resChan := make(chan executor.ResponseContent[response.ListResponseDto[domain.TeamDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.GetTeamsReq, response.ListResponseDto[domain.TeamDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
