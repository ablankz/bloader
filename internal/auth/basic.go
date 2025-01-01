package auth

import (
	"context"
	"net/http"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/store"
)

// BasicAuthenticator is an Authenticator that uses Basic Auth
type BasicAuthenticator struct {
	Username string
	Password string
}

// NewBasicAuthenticator creates a new BasicAuthenticator
func NewBasicAuthenticator(conf config.ValidAuthBasicConfig) (Authenticator, error) {
	return &BasicAuthenticator{
		Username: conf.Username,
		Password: conf.Password,
	}, nil
}

// Authenticate authenticates the user
func (a *BasicAuthenticator) Authenticate(ctx context.Context, str store.Store) error {
	return nil
}

// SetOnRequest sets the authentication information on the request
func (a *BasicAuthenticator) SetOnRequest(ctx context.Context, r *http.Request) {
	r.SetBasicAuth(a.Username, a.Password)
}

// GetAuthValue returns the authentication value
func (a *BasicAuthenticator) GetAuthValue() *pb.Auth {
	return &pb.Auth{
		Type: pb.AuthType_AUTH_TYPE_BASIC,
		Auth: &pb.Auth_Basic{
			Basic: &pb.AuthBasic{
				Username: a.Username,
				Password: a.Password,
			},
		},
	}
}

// IsExpired checks if the authentication information is expired
func (a *BasicAuthenticator) IsExpired(ctx context.Context, str store.Store) bool {
	return false
}

// Refresh refreshes the authentication information
func (a *BasicAuthenticator) Refresh(ctx context.Context, str store.Store) error {
	return nil
}

var _ Authenticator = &BasicAuthenticator{}
var _ SetAuthor = &BasicAuthenticator{}
