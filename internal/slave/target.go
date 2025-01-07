package slave

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/runner"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
	"github.com/ablankz/bloader/internal/target"
)

// SlaveTargetFactor represents the factory for the slave target
type SlaveTargetFactor struct {
	target                        *slcontainer.Target
	connectionID                  string
	receiveChanelRequestContainer *slcontainer.ReceiveChanelRequestContainer
	mapper                        *slcontainer.RequestConnectionMapper
}

// Factorize returns the factorized target
func (s *SlaveTargetFactor) Factorize(ctx context.Context, targetID string) (target.Target, error) {
	t, ok := s.target.Find(targetID)
	if ok {
		return t, nil
	}

	term := s.receiveChanelRequestContainer.SendTargetResourceRequests(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.TargetResourceRequest{
			TargetID: targetID,
		},
	)
	if term == nil {
		return target.Target{}, fmt.Errorf("failed to send target resource request")
	}
	select {
	case <-ctx.Done():
		return target.Target{}, nil
	case <-term:
	}

	t, ok = s.target.Find(targetID)
	if !ok {
		return target.Target{}, fmt.Errorf("target not found: %s", targetID)
	}

	return t, nil
}

var _ runner.TargetFactor = &SlaveTargetFactor{}
