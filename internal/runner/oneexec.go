package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/output"
)

// OneExecType represents the type of OneExec
type OneExecType string

const (
	// OneExecTypeHTTP represents the HTTP type
	OneExecTypeHTTP OneExecType = "http"
)

// OneExec represents the OneExec runner
type OneExec struct {
	Type   *string       `yaml:"type"`
	Output OneExecOutput `yaml:"output"`
	Auth   OneExecAuth   `yaml:"auth"`
}

// ValidOneExec represents the valid OneExec runner
type ValidOneExec struct {
	Type   OneExecType
	Output []output.Output
	Auth   *auth.Authenticator
}

// Validate validates the OneExec
func (r OneExec) Validate(ctr *container.Container, oCtr output.OutputContainer) (ValidOneExec, error) {
	var oneExecType OneExecType
	if r.Type == nil {
		return ValidOneExec{}, fmt.Errorf("type is required")
	}
	switch OneExecType(*r.Type) {
	case OneExecTypeHTTP:
		oneExecType = OneExecType(*r.Type)
	default:
		return ValidOneExec{}, fmt.Errorf("invalid type value: %s", *r.Type)
	}
	var validOutput []output.Output
	var validAuth *auth.Authenticator
	validOutput, err := r.Output.Validate(oCtr)
	if err != nil {
		return ValidOneExec{}, fmt.Errorf("failed to validate output: %v", err)
	}
	validAuth, err = r.Auth.Validate(ctr)
	if err != nil {
		return ValidOneExec{}, fmt.Errorf("failed to validate auth: %v", err)
	}
	return ValidOneExec{
		Type:   oneExecType,
		Output: validOutput,
		Auth:   validAuth,
	}, nil
}

// OneExecOutput represents the output configuration for the OneExec runner
type OneExecOutput struct {
	Enabled bool     `yaml:"enabled"`
	IDs     []string `yaml:"ids"`
}

// Validate validates the OneExecOutput
func (o OneExecOutput) Validate(oCtr output.OutputContainer) ([]output.Output, error) {
	if !o.Enabled {
		return nil, nil
	}
	var outputs []output.Output
	for _, id := range o.IDs {
		output, exists := oCtr[id]
		if !exists {
			return nil, fmt.Errorf("output id does not exist")
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

// OneExecAuth represents the auth configuration for the OneExec runner
type OneExecAuth struct {
	Enabled bool    `yaml:"enabled"`
	AuthID  *string `yaml:"auth_id"`
}

// Validate validates the OneExecAuth
func (a OneExecAuth) Validate(ctr *container.Container) (*auth.Authenticator, error) {
	if !a.Enabled {
		return nil, nil
	}
	var auth *auth.Authenticator
	var exists bool
	var authID string
	if a.AuthID == nil {
		authID = ctr.AuthenticatorContainer.DefaultAuthenticator
	} else {
		authID = *a.AuthID
	}
	auth, exists = ctr.AuthenticatorContainer.Container[authID]
	if !exists {
		return nil, fmt.Errorf("auth_id: %s does not exist", authID)
	}
	return auth, nil
}

// Run runs the OneExec runner
func (r ValidOneExec) Run(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
	threadOnlyStore *sync.Map,
) error {
	return nil
}
