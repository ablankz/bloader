package auth

import (
	"context"
	"net/http"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"

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

// GetAuthValue returns the authentication value
func (a *APIKeyAuthenticator) GetAuthValue() *pb.Auth {
	return &pb.Auth{
		Type: pb.AuthType_AUTH_TYPE_API_KEY,
		Auth: &pb.Auth_ApiKey{
			ApiKey: &pb.AuthApiKey{
				ApiKey:     a.APIKey,
				HeaderName: a.HeaderName,
			},
		},
	}
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
