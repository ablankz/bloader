package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/executor/httpexec"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/utils"
)

// OneExecType represents the type of OneExec
type OneExecType string

const (
	// OneExecTypeHTTP represents the HTTP type
	OneExecTypeHTTP OneExecType = "http"
)

// OneExec represents the OneExec runner
type OneExec struct {
	Type    *string        `yaml:"type"`
	Output  OneExecOutput  `yaml:"output"`
	Auth    OneExecAuth    `yaml:"auth"`
	Request OneExecRequest `yaml:"request"`
}

// ValidOneExec represents the valid OneExec runner
type ValidOneExec struct {
	Type    OneExecType
	Output  []output.Output
	Auth    *auth.Authenticator
	Request ValidOneExecRequest
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
	validRequest, err := r.Request.Validate(ctr, oneExecType)
	if err != nil {
		return ValidOneExec{}, fmt.Errorf("failed to validate request: %v", err)
	}
	return ValidOneExec{
		Type:    oneExecType,
		Output:  validOutput,
		Auth:    validAuth,
		Request: validRequest,
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

// OneExecRequest represents the request configuration for the OneExec runner
type OneExecRequest struct {
	TargetID      *string                `yaml:"target_id"`
	Endpoint      *string                `yaml:"endpoint"`
	Method        *string                `yaml:"method"`
	QueryParam    map[string]any         `yaml:"query_param"`
	PathVariables map[string]string      `yaml:"path_variables"`
	Headers       map[string]any         `yaml:"headers"`
	BodyType      *string                `yaml:"body_type"`
	Body          any                    `yaml:"body"`
	ResponseType  *string                `yaml:"response_type"`
	Data          []ExecRequestData      `yaml:"data"`
	MemoryData    []ExecRequestData      `yaml:"memory_data"`
	StoreData     []ExecRequestStoreData `yaml:"store_data"`
}

// ValidOneExecRequest represents the valid request configuration for the OneExec runner
type ValidOneExecRequest struct {
	URL           string
	Method        string
	QueryParam    map[string]any
	PathVariables map[string]string
	Headers       map[string]any
	BodyType      HTTPRequestBodyType
	Body          any
	ResponseType  string
	Data          ValidExecRequestDataSlice
	MemoryData    ValidExecRequestDataSlice
	StoreData     []ValidExecRequestStoreData
}

// Validate validates the OneExecRequest
func (r OneExecRequest) Validate(ctr *container.Container, targetType OneExecType) (ValidOneExecRequest, error) {
	var valid ValidOneExecRequest
	var err error
	if r.TargetID == nil {
		return ValidOneExecRequest{}, fmt.Errorf("target_id is required")
	}
	if r.Endpoint == nil {
		return ValidOneExecRequest{}, fmt.Errorf("endpoint is required")
	}
	var urlRoot string
	if urlRoot, err = ctr.TargetContainer.FindTarget(*r.TargetID, config.TargetType(targetType)); err != nil {
		return ValidOneExecRequest{}, fmt.Errorf("failed to find target: %v", err)
	}
	valid.URL = fmt.Sprintf("%s%s", urlRoot, *r.Endpoint)
	if r.Method == nil {
		return ValidOneExecRequest{}, fmt.Errorf("method is required")
	}
	valid.Method = *r.Method

	valid.QueryParam = r.QueryParam
	valid.PathVariables = r.PathVariables
	valid.Headers = r.Headers
	valid.Body = r.Body
	if r.BodyType == nil {
		valid.BodyType = DefaultHTTPRequestBodyType
	} else {
		switch HTTPRequestBodyType(*r.BodyType) {
		case HTTPRequestBodyTypeJSON, HTTPRequestBodyTypeForm, HTTPRequestBodyTypeMultipart:
			valid.BodyType = HTTPRequestBodyTypeJSON
		default:
			return ValidOneExecRequest{}, fmt.Errorf("invalid body_type value: %s", *r.BodyType)
		}
	}
	if r.ResponseType == nil {
		return ValidOneExecRequest{}, fmt.Errorf("response_type is required")
	}
	valid.ResponseType = *r.ResponseType
	for _, d := range r.Data {
		validData, err := d.Validate()
		if err != nil {
			return ValidOneExecRequest{}, fmt.Errorf("failed to validate data: %v", err)
		}
		valid.Data = append(valid.Data, validData)
	}
	for _, d := range r.MemoryData {
		validData, err := d.Validate()
		if err != nil {
			return ValidOneExecRequest{}, fmt.Errorf("failed to validate memory data: %v", err)
		}
		valid.MemoryData = append(valid.MemoryData, validData)
	}
	for _, d := range r.StoreData {
		validData, err := d.Validate()
		if err != nil {
			return ValidOneExecRequest{}, fmt.Errorf("failed to validate store data: %v", err)
		}
		valid.StoreData = append(valid.StoreData, validData)
	}
	return valid, nil
}

// Run runs the OneExec runner
func (r ValidOneExec) Run(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
) error {
	switch r.Type {
	case OneExecTypeHTTP:
		return r.runHTTP(ctx, ctr, outputRoot, str)
	}
	return nil
}

func (r ValidOneExec) runHTTP(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
) error {
	req := HTTPRequest{
		Method:        r.Request.Method,
		URL:           r.Request.URL,
		Headers:       r.Request.Headers,
		QueryParams:   r.Request.QueryParam,
		PathVariables: r.Request.PathVariables,
		BodyType:      r.Request.BodyType,
		Body:          r.Request.Body,
		AttachRequestInfo: func(ctx context.Context, ctr *container.Container, req *http.Request) error {
			(*r.Auth).SetOnRequest(ctx, ctr.Store, req)
			return nil
		},
	}
	exe := httpexec.RequestContent[HTTPRequest]{
		Req:          req,
		ResponseType: httpexec.ResponseType(r.Request.ResponseType),
	}

	writers := make([]output.HTTPDataWrite, 0)
	uniqueName := fmt.Sprintf("%s/%s", outputRoot, utils.GenerateUniqueID())
	for _, o := range r.Output {
		writer, close, err := o.HTTPDataWriteFactory(
			ctx,
			ctr,
			true,
			uniqueName,
			append(
				[]string{
					"Success",
					"SendDatetime",
					"ReceivedDatetime",
					"Count",
					"ResponseTime",
					"StatusCode",
				},
				r.Request.Data.ExtractHeader()...,
			),
		)
		if err != nil {
			return fmt.Errorf("failed to create writer: %v", err)
		}
		defer close()
		writers = append(writers, writer)
	}

	resp, err := exe.RequestExecute(ctx, ctr)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	var data []string
	for _, d := range r.Request.Data {
		result, err := d.Extractor.Extract(resp.Res)
		if err != nil {
			return fmt.Errorf("failed to extract data: %v", err)
		}
		data = append(data, fmt.Sprint(result))
	}
	for _, w := range writers {
		if err := w(ctx, ctr, append(resp.ToWriteHTTPData(1).ToSlice(), data...)); err != nil {
			return fmt.Errorf("failed to write data: %v", err)
		}
	}

	for _, d := range r.Request.MemoryData {
		result, err := d.Extractor.Extract(resp.Res)
		if err != nil {
			return fmt.Errorf("failed to extract memory data: %v", err)
		}
		str.Store(d.Key, result)
	}

	for _, d := range r.Request.StoreData {
		result, err := d.Extractor.Extract(resp.Res)
		if err != nil {
			return fmt.Errorf("failed to extract store data: %v", err)
		}
		byteData, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal store data: %v", err)
		}
		if d.Encrypt.Enabled {
			enc, ok := ctr.EncypterContainer[d.Encrypt.EncryptID]
			if !ok {
				return fmt.Errorf("encrypt_id: %s does not exist", d.Encrypt.EncryptID)
			}
			strData, err := enc.Encrypt(byteData)
			if err != nil {
				return fmt.Errorf("failed to encrypt store data: %v", err)
			}
			byteData = []byte(strData)
		}
		if err := ctr.Store.PutObject(d.BucketID, d.StoreKey, byteData); err != nil {
			return fmt.Errorf("failed to put store data: %v", err)
		}
	}

	return nil
}
