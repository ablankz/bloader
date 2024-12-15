package massexec

import (
	"context"
	"fmt"
	"sync"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/output"
)

// MassExecType represents the type of MassExec
type MassExecType string

const (
	// MassExecTypeHTTP represents the HTTP type
	MassExecTypeHTTP MassExecType = "http"
)

// MassExec represents the MassExec runner
type MassExec struct {
	Type   *string        `yaml:"type"`
	Output MassExecOutput `yaml:"output"`
	Auth   MassExecAuth   `yaml:"auth"`
}

// ValidMassExec represents the valid MassExec runner
type ValidMassExec struct {
	Type   MassExecType
	Output []output.Output
	Auth   *auth.Authenticator
}

// Validate validates the MassExec
func (r MassExec) Validate(ctr *container.Container, oCtr output.OutputContainer) (ValidMassExec, error) {
	var massExecType MassExecType
	if r.Type == nil {
		return ValidMassExec{}, fmt.Errorf("type is required")
	}
	switch MassExecType(*r.Type) {
	case MassExecTypeHTTP:
		massExecType = MassExecType(*r.Type)
	default:
		return ValidMassExec{}, fmt.Errorf("invalid type value: %s", *r.Type)
	}
	var validOutput []output.Output
	var validAuth *auth.Authenticator
	validOutput, err := r.Output.Validate(oCtr)
	if err != nil {
		return ValidMassExec{}, fmt.Errorf("failed to validate output: %v", err)
	}
	validAuth, err = r.Auth.Validate(ctr)
	if err != nil {
		return ValidMassExec{}, fmt.Errorf("failed to validate auth: %v", err)
	}
	return ValidMassExec{
		Type:   massExecType,
		Output: validOutput,
		Auth:   validAuth,
	}, nil
}

// MassExecOutput represents the output configuration for the MassExec runner
type MassExecOutput struct {
	Enabled bool     `yaml:"enabled"`
	IDs     []string `yaml:"ids"`
}

// Validate validates the MassExecOutput
func (o MassExecOutput) Validate(oCtr output.OutputContainer) ([]output.Output, error) {
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

// MassExecAuth represents the auth configuration for the MassExec runner
type MassExecAuth struct {
	Enabled bool    `yaml:"enabled"`
	AuthID  *string `yaml:"auth_id"`
}

// Validate validates the MassExecAuth
func (a MassExecAuth) Validate(ctr *container.Container) (*auth.Authenticator, error) {
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

// Run runs the MassExec runner
func (r ValidMassExec) Run(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
	threadOnlyStore *sync.Map,
) error {
	return nil
}
