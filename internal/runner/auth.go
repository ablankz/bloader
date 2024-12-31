package runner

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/auth"
)

// AuthenticatorFactor represents the authenticator factor
type AuthenticatorFactor interface {
	// Factorize returns the factorized authenticator
	Factorize(ctx context.Context, authID string, isDefault bool) (auth.SetAuthor, error)
}

// LocalAuthenticatorFactor represents the local authenticator factor
type LocalAuthenticatorFactor struct {
	authCtr auth.AuthenticatorContainer
}

// NewLocalAuthenticatorFactor creates a new local authenticator factor
func NewLocalAuthenticatorFactor(authCtr auth.AuthenticatorContainer) *LocalAuthenticatorFactor {
	return &LocalAuthenticatorFactor{
		authCtr: authCtr,
	}
}

// Factorize returns the factorized authenticator
func (l LocalAuthenticatorFactor) Factorize(
	ctx context.Context,
	authID string,
	isDefault bool,
) (auth.SetAuthor, error) {
	if isDefault {
		authID = l.authCtr.DefaultAuthenticator
	}
	authenticator, exists := l.authCtr.Container[authID]
	if !exists {
		return nil, fmt.Errorf("auth_id: %s does not exist", authID)
	}

	return *authenticator, nil
}
