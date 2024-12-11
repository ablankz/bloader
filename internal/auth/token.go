package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

type AuthToken struct {
	AccessToken  string    `yaml:"access_token"`
	RefreshToken string    `yaml:"refresh_token"`
	TokenType    string    `yaml:"token_type"`
	Expiry       time.Time `yaml:"expiry"`
}

func NewAuthToken(accessToken, refreshToken, tokenType string, expiry time.Time) *AuthToken {
	return &AuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    tokenType,
		Expiry:       expiry,
	}
}

func (t *AuthToken) SetAuthHeader(r *http.Request) {
	r.Header.Set("Authorization", t.TokenType+" "+t.AccessToken)
}

func (t *AuthToken) IsExpired() bool {
	return t.Expiry.Before(time.Now())
}

func (t *AuthToken) Refresh(ctx context.Context, config *oauth2.Config, credentialFilePath string) error {

	file, err := os.OpenFile(credentialFilePath, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := file.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate file: %w", err)
	}

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	tokenSource := config.TokenSource(ctx, &oauth2.Token{
		RefreshToken: t.RefreshToken,
	})

	newToken, err := tokenSource.Token()

	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	t.AccessToken = newToken.AccessToken
	t.RefreshToken = newToken.RefreshToken
	t.TokenType = newToken.TokenType
	t.Expiry = newToken.Expiry

	if err := t.Save(encoder); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	return nil
}

func (t *AuthToken) Save(encoder *yaml.Encoder) error {
	return encoder.Encode(t)
}

func (t *AuthToken) Load(decoder *yaml.Decoder) error {
	return decoder.Decode(t)
}
