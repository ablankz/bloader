package slave

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/runner"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
)

// SlaveAuthenticatorFactor represents the slave authenticator factor
type SlaveAuthenticatorFactor struct {
	auth                          *slcontainer.Auth
	connectionID                  string
	receiveChanelRequestContainer *slcontainer.ReceiveChanelRequestContainer
	mapper                        *slcontainer.RequestConnectionMapper
}

// Factorize returns the factorized authenticator
func (s *SlaveAuthenticatorFactor) Factorize(
	ctx context.Context,
	authID string,
	isDefault bool,
) (auth.SetAuthor, error) {
	if isDefault {
		if s.auth.DefaultAuthenticator == "" {
			term := s.receiveChanelRequestContainer.SendAuthResourceRequests(
				ctx,
				s.connectionID,
				s.mapper,
				slcontainer.AuthResourceRequest{
					IsDefault: true,
				},
			)
			if term == nil {
				return nil, fmt.Errorf("failed to send auth resource request")
			}
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context canceled")
			case <-term:
			}
		}
		authID = s.auth.DefaultAuthenticator
	}

	a, ok := s.auth.Find(authID)
	if ok {
		return a, nil
	}

	term := s.receiveChanelRequestContainer.SendAuthResourceRequests(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.AuthResourceRequest{
			AuthID: authID,
		},
	)
	if term == nil {
		return nil, fmt.Errorf("failed to send auth resource request")
	}
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context canceled")
	case <-term:
	}

	a, ok = s.auth.Find(authID)
	if !ok {
		return nil, fmt.Errorf("auth not found: %s", authID)
	}

	return a, nil
}

var _ runner.AuthenticatorFactor = &SlaveAuthenticatorFactor{}
