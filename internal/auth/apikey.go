package auth

import (
	"context"
	"net/http"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/store"
)

// APIKeyAuthenticator is an Authenticator that uses API Key
type APIKeyAuthenticator struct {
	APIKey     string
	HeaderName string
}

// Authenticate authenticates the user
func NewAPIKeyAuthenticator(conf config.ValidAuthAPIKeyConfig) (Authenticator, error) {
	return &APIKeyAuthenticator{
		APIKey:     conf.Key,
		HeaderName: conf.HeaderName,
	}, nil
}

// Authenticate authenticates the user
func (a *APIKeyAuthenticator) Authenticate(ctx context.Context, str store.Store) error {
	return nil
}

// SetOnRequest sets the authentication information on the request
func (a *APIKeyAuthenticator) SetOnRequest(ctx context.Context, r *http.Request) {
	r.Header.Set(a.HeaderName, a.APIKey)
}

// IsExpired checks if the authentication information is expired
func (a *APIKeyAuthenticator) IsExpired(ctx context.Context, str store.Store) bool {
	return false
}

// Refresh refreshes the authentication information
func (a *APIKeyAuthenticator) Refresh(ctx context.Context, str store.Store) error {
	return nil
}

var _ Authenticator = &APIKeyAuthenticator{}
var _ SetAuthor = &APIKeyAuthenticator{}
