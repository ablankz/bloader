// Package auth provides the authentication logic for the application
package auth

import (
	"context"
	"net/http"

	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/store"
)

// Authenticator is an interface for authenticating
type Authenticator interface {
	// Authenticate authenticates the user
	Authenticate(ctx context.Context, str store.Store) error
	// SetOnRequest sets the authentication information on the request
	SetOnRequest(ctx context.Context, str store.Store, r *http.Request)
	// IsExpired checks if the authentication information is expired
	IsExpired(ctx context.Context, str store.Store) bool
	// Refresh refreshes the authentication information
	Refresh(ctx context.Context, str store.Store) error
}

// AuthAuthenticatorContainer holds the dependencies for the Authenticator
type AuthAuthenticatorContainer struct {
	DefaultAuthenticator string
	Container            map[string]*Authenticator
}

// NewAuthAuthenticatorContainer creates a new AuthAuthenticatorContainer
func NewAuthAuthenticatorContainerFromConfig(str store.Store, conf config.ValidConfig) (AuthAuthenticatorContainer, error) {
	var ctr AuthAuthenticatorContainer = AuthAuthenticatorContainer{
		Container: make(map[string]*Authenticator),
	}

	for _, authConf := range conf.Auth {
		var authenticator Authenticator
		var err error
		switch authConf.Type {
		case config.AuthTypeOAuth2:
			var redirectPort int
			if conf.Server.RedirectPort.Enabled {
				redirectPort = conf.Server.RedirectPort.Port
			} else {
				redirectPort = conf.Server.Port
			}
			authenticator, err = NewOAuthAuthenticator(str, redirectPort, authConf.OAuth2)
			if err != nil {
				return AuthAuthenticatorContainer{}, err
			}
		case config.AuthTypeBasic:
			authenticator, err = NewBasicAuthenticator(authConf.Basic)
			if err != nil {
				return AuthAuthenticatorContainer{}, err
			}
		case config.AuthTypeAPIKey:
			authenticator, err = NewAPIKeyAuthenticator(authConf.APIKey)
			if err != nil {
				return AuthAuthenticatorContainer{}, err
			}
		case config.AuthTypeJWT:
			authenticator, err = NewJWTAuthenticator(authConf.JWT)
			if err != nil {
				return AuthAuthenticatorContainer{}, err
			}
		}

		ctr.Container[authConf.ID] = &authenticator
		if authConf.Default {
			ctr.DefaultAuthenticator = authConf.ID
		}
	}

	return ctr, nil
}
