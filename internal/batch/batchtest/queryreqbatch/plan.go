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

type FindTaskFactory struct{}

func (f FindTaskFactory) Factory(
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

	req := executor.FindTaskReq{
		AuthToken:    authToken,
		BaseEndpoint: apiEndpoint,
	}

	if taskId, ok = request.PathVariables["taskId"]; !ok {
		return nil, nil, fmt.Errorf("taskId not found in pathVariables")
	}
	req.Path.TaskID = taskId

	for key, param := range request.QueryParam {
		switch key {
		case "with":
			req.Param.With = param
		}
	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ResponseDto[domain.TaskDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.FindTaskReq, response.ResponseDto[domain.TaskDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}

type GetTasksFactory struct{}

func (f GetTasksFactory) Factory(
	ctx context.Context,
	ctr *app.Container,
	id int,
	request *execbatch.ValidatedExecRequest,
	termChan chan<- execbatch.TerminateType,
	authToken *auth.AuthToken,
	apiEndpoint string,
	consumer execbatch.ResponseDataConsumer,
) (executor.RequestExecutor, func(), error) {
	req := executor.GetTasksReq{
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
		case "hasTeamFilter":
			hasTeamFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasTeamFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasTeamFilter = hasTeamFilterBool
		case "filterTeamIDs":
			req.Param.FilterTeamIDs = param
		case "hasStatusFilter":
			hasStatusFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasStatusFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasStatusFilter = hasStatusFilterBool
		case "filterStatuses":
			req.Param.FilterStatuses = param
		case "hasChargeUserFilter":
			hasChargeUserFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasChargeUserFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasChargeUserFilter = hasChargeUserFilterBool
		case "filterChargeUserIDs":
			req.Param.FilterChargeUserIDs = param
		case "filterStartDatetimeEarlierThan":
			req.Param.FilterStartDatetimeEarlierThan = param[0]
		case "filterStartDatetimeLaterThan":
			req.Param.FilterStartDatetimeLaterThan = param[0]
		case "filterDueDatetimeEarlierThan":
			req.Param.FilterDueDatetimeEarlierThan = param[0]
		case "filterDueDatetimeLaterThan":
			req.Param.FilterDueDatetimeLaterThan = param[0]
		case "hasFileObjectFilter":
			hasFileObjectFilterBool, err := strconv.ParseBool(param[0])
			if err != nil {
				ctr.Logger.Warn(ctx, "Failed to convert hasFileObjectFilter to bool",
					logger.Value("error", err))
				continue
			}
			req.Param.HasFileObjectFilter = hasFileObjectFilterBool
		case "filterFileObjectIDs":
			req.Param.FilterFileObjectIDs = param
		case "fileObjectFilterType":
			req.Param.FileObjectFilterType = param[0]
		case "with":
			req.Param.With = param
		}

	}

	// INFO: close on executor, because only it will write to this channel
	resChan := make(chan executor.ResponseContent[response.ListResponseDto[domain.TaskDto]])

	resChanCloser := func() {
		close(resChan)
	}

	execbatch.RunAsyncProcessing(ctx, ctr, id, request, termChan, resChan, consumer)

	return executor.RequestContent[executor.GetTasksReq, response.ListResponseDto[domain.TaskDto]]{
		Req:          req,
		Interval:     request.Interval,
		ResponseWait: request.AwaitPrevResp,
		ResChan:      resChan,
		CountLimit:   request.Break.Count,
	}, resChanCloser, nil
}
