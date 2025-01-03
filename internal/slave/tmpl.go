package slave

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/runner"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
)

type SlaveTmplFactor struct {
	loader                        *slcontainer.Loader
	connectionID                  string
	receiveChanelRequestContainer *slcontainer.ReceiveChanelRequestContainer
	mapper                        *slcontainer.RequestConnectionMapper
}

// TmplFactorize is a function that factorizes the template
func (s *SlaveTmplFactor) TmplFactorize(ctx context.Context, path string) (string, error) {
	l, ok := s.loader.GetLoader(path)
	if ok {
		return l, nil
	}

	term := s.receiveChanelRequestContainer.SendLoaderResourceRequests(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.LoaderResourceRequest{
			LoaderID: path,
		},
	)
	if term == nil {
		return "", fmt.Errorf("failed to send loader resource request")
	}
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context canceled")
	case <-term:
	}

	l, ok = s.loader.GetLoader(path)
	if ok {
		return l, nil
	}

	return "", fmt.Errorf("failed to factorize the template")
}

var _ runner.TmplFactor = &SlaveTmplFactor{}
