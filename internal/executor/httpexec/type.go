package httpexec

import (
	"context"
	"fmt"
	"strings"

	"github.com/ablankz/bloader/internal/container"
)

// TerminateType represents the type of terminate
type TerminateType string

const (
	// TerminateTypeByContext represents the context type
	TerminateTypeByContext TerminateType = "context"
	// TerminateTypeByCount represents the count type
	TerminateTypeByCount TerminateType = "count"
	// TerminateTypeBySystemError represents the system error type
	TerminateTypeBySystemError TerminateType = "sysError"
	// TerminateTypeByCreateRequestErr represents the create request error type
	TerminateTypeByCreateRequestErr TerminateType = "createRequestError"
	// TerminateTypeByParseResponseErr represents the parse response error type
	TerminateTypeByParseResponseErr TerminateType = "parseError"
	// TerminateTypeByWriteError represents the write error type
	TerminateTypeByWriteError TerminateType = "writeError"
	// TerminateTypeByTimeout represents the timeout type
	TerminateTypeByTimeout TerminateType = "time"
	// TerminateTypeByResponseBody represents the response body type
	TerminateTypeByResponseBody TerminateType = "responseBody"
	// TerminateTypeByStatusCode represents the status code type
	TerminateTypeByStatusCode TerminateType = "statusCode"
)

func NewTerminateTypeFromString(s string) (TerminateType, []string, error) {
	strs := strings.Split(s, "/")
	params := strings.Split(strs[1], ",")
	switch TerminateType(strs[0]) {
	case TerminateTypeByContext:
		return TerminateTypeByContext, nil, nil
	case TerminateTypeByCount:
		return TerminateTypeByCount, nil, nil
	case TerminateTypeBySystemError:
		return TerminateTypeBySystemError, nil, nil
	case TerminateTypeByCreateRequestErr:
		return TerminateTypeByCreateRequestErr, nil, nil
	case TerminateTypeByParseResponseErr:
		return TerminateTypeByParseResponseErr, nil, nil
	case TerminateTypeByWriteError:
		return TerminateTypeByWriteError, nil, nil
	case TerminateTypeByTimeout:
		return TerminateTypeByTimeout, nil, nil
	case TerminateTypeByResponseBody:
		return TerminateTypeByResponseBody, params, nil
	case TerminateTypeByStatusCode:
		return TerminateTypeByStatusCode, params, nil
	default:
		return "", nil, fmt.Errorf("invalid terminate type: %s", s)
	}
}

// ResponseType represents the response type
type ResponseType string

const (
	// ResponseTypeJSON represents the JSON response type
	ResponseTypeJSON ResponseType = "json"
	// ResponseTypeXML represents the XML response type
	ResponseTypeXML ResponseType = "xml"
	// ResponseTypeYAML represents the YAML response type
	ResponseTypeYAML ResponseType = "yaml"
	// ResponseTypeText represents the text response type
	ResponseTypeText ResponseType = "text"
	// ResponseTypeHTML represents the HTML response type
	ResponseTypeHTML ResponseType = "html"
)

// RequestExecutor represents the request executor
type RequestExecutor interface {
	// RequestExecute executes the request
	RequestExecute(ctx context.Context, ctr *container.Container) (ResponseContent, error)
}

// MassRequestExecutor represents the request executor
type MassRequestExecutor interface {
	// MassRequestExecute executes the request
	MassRequestExecute(ctx context.Context, ctr *container.Container) error
}
