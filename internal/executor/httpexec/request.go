package httpexec

import (
	"context"
	"net/http"

	"github.com/ablankz/bloader/internal/container"
)

// ExecReq represents the request executor
type ExecReq interface {
	// CreateRequest creates the http.Request object for the query
	CreateRequest(ctx context.Context, ctr *container.Container, count int) (*http.Request, error)
}
