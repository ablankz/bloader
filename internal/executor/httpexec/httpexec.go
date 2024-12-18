package httpexec

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/utils"
	"gopkg.in/yaml.v3"
)

// RequestContent represents the request content
type RequestContent[Req ExecReq] struct {
	Req          Req
	ResponseType ResponseType
}

// RequestExecute executes the request
func (q RequestContent[Req]) RequestExecute(
	ctx context.Context,
	ctr *container.Container,
) (ResponseContent, error) {
	req, err := q.Req.CreateRequest(ctx, ctr)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to create request",
			logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"))
		return ResponseContent{}, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Minute,
		Transport: &utils.DelayedTransport{
			Transport: http.DefaultTransport,
			// Delay:     2 * time.Second,
		},
	}

	ctr.Logger.Debug(ctx, "sending request",
		logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
	startTime := time.Now()
	resp, err := client.Do(req)
	endTime := time.Now()
	ctr.Logger.Debug(ctx, "received response",
		logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))

	if err != nil {
		ctr.Logger.Error(ctx, "response error",
			logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
		return ResponseContent{
			Success:      false,
			StartTime:    startTime,
			EndTime:      endTime,
			ResponseTime: endTime.Sub(startTime).Milliseconds(),
			HasSystemErr: true,
		}, nil
	}
	defer resp.Body.Close()

	statusCode := resp.StatusCode
	var response any
	responseByte, err := io.ReadAll(resp.Body)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to read response",
			logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))

		return ResponseContent{
			Success:        false,
			Res:            response,
			StartTime:      startTime,
			EndTime:        endTime,
			ResponseTime:   endTime.Sub(startTime).Milliseconds(),
			StatusCode:     statusCode,
			ParseResHasErr: true,
		}, nil
	}
	switch ResponseType(q.ResponseType) {
	case ResponseTypeJSON:
		err = json.Unmarshal(responseByte, &response)
	case ResponseTypeXML:
		err = xml.Unmarshal(responseByte, &response)
	case ResponseTypeYAML:
		err = yaml.Unmarshal(responseByte, &response)
	case ResponseTypeText:
		response = string(responseByte)
	case ResponseTypeHTML:
		response = string(responseByte)
	default:
		err = fmt.Errorf("invalid response type: %s", q.ResponseType)
	}
	if err != nil {
		ctr.Logger.Error(ctx, "failed to parse response",
			logger.Value("error", err), logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
		return ResponseContent{
			Success:        false,
			Res:            response,
			ByteResponse:   responseByte,
			StartTime:      startTime,
			EndTime:        endTime,
			ResponseTime:   endTime.Sub(startTime).Milliseconds(),
			StatusCode:     statusCode,
			ParseResHasErr: true,
		}, nil
	}
	ctr.Logger.Debug(ctx, "response OK",
		logger.Value("on", "RequestContent.QueryExecute"), logger.Value("url", req.URL))
	return ResponseContent{
		Success:      true,
		ByteResponse: responseByte,
		Res:          response,
		StartTime:    startTime,
		EndTime:      endTime,
		ResponseTime: endTime.Sub(startTime).Milliseconds(),
		StatusCode:   statusCode,
	}, nil
}

var _ RequestExecutor = RequestContent[ExecReq]{} // ensure that RequestContent implements RequestExecutor
