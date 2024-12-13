package auth

import (
	"context"
	"net/http"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/store"
)

// JWTAuthenticator is an Authenticator that uses Private Key
type JWTAuthenticator struct {
	credentialConf config.ValidAuthCredentialConfig
}

// Authenticate authenticates the user
func NewJWTAuthenticator(conf config.ValidAuthJWTConfig) (Authenticator, error) {
	return &JWTAuthenticator{
		credentialConf: conf.Credential,
	}, nil
}

// Authenticate authenticates the user
func (a *JWTAuthenticator) Authenticate(ctx context.Context, str store.Store) error {
	return nil
}

// SetOnRequest sets the authentication information on the request
func (a *JWTAuthenticator) SetOnRequest(ctx context.Context, str store.Store, r *http.Request) {
}

// IsExpired checks if the authentication information is expired
func (a *JWTAuthenticator) IsExpired(ctx context.Context, str store.Store) bool {
	return false
}

// Refresh refreshes the authentication information
func (a *JWTAuthenticator) Refresh(ctx context.Context, str store.Store) error {
	return nil
}
