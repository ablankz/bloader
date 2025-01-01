package auth

import (
	"context"
	"net/http"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"

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
func (a *JWTAuthenticator) SetOnRequest(ctx context.Context, r *http.Request) {
}

// GetAuthValue returns the authentication value
func (a *JWTAuthenticator) GetAuthValue() *pb.Auth {
	return &pb.Auth{
		Type: pb.AuthType_AUTH_TYPE_JWT,
		Auth: &pb.Auth_Jwt{
			Jwt: &pb.AuthJwt{
				Jwt: "", // TODO: Implement
			},
		},
	}
}

// IsExpired checks if the authentication information is expired
func (a *JWTAuthenticator) IsExpired(ctx context.Context, str store.Store) bool {
	return false
}

// Refresh refreshes the authentication information
func (a *JWTAuthenticator) Refresh(ctx context.Context, str store.Store) error {
	return nil
}

var _ Authenticator = &JWTAuthenticator{}
var _ SetAuthor = &JWTAuthenticator{}
