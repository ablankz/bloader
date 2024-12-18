package runner

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/batch/batchtest/batchexecmap"
	"github.com/ablankz/bloader/internal/batch/batchtest/execbatch"
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/executor/httpexec"
	"github.com/ablankz/bloader/internal/logger"
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
	Type     *string           `yaml:"type"`
	Output   MassExecOutput    `yaml:"output"`
	Auth     MassExecAuth      `yaml:"auth"`
	Requests []MassExecRequest `yaml:"requests"`
}

// ValidMassExec represents the valid MassExec runner
type ValidMassExec struct {
	Type     MassExecType
	Output   []output.Output
	Auth     *auth.Authenticator
	Requests []ValidMassExecRequest
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
	var validRequests []ValidMassExecRequest
	for _, req := range r.Requests {
		validRequest, err := req.Validate(ctr, massExecType)
		if err != nil {
			return ValidMassExec{}, fmt.Errorf("failed to validate request: %v", err)
		}
		validRequests = append(validRequests, validRequest)
	}
	return ValidMassExec{
		Type:     massExecType,
		Output:   validOutput,
		Auth:     validAuth,
		Requests: validRequests,
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

// MassExecRequest represents the request configuration for the MassExec runner
type MassExecRequest struct {
	TargetID      *string           `yaml:"target_id"`
	Endpoint      *string           `yaml:"endpoint"`
	Method        *string           `yaml:"method"`
	QueryParam    map[string]any    `yaml:"query_param"`
	PathVariables map[string]string `yaml:"path_variables"`
	Headers       map[string]any    `yaml:"headers"`
	BodyType      *string           `yaml:"body_type"`
	Body          any               `yaml:"body"`
	ResponseType  *string           `yaml:"response_type"`
	Data          []ExecRequestData `yaml:"data"`
}

// ValidMassExecRequest represents the valid request configuration for the MassExec runner
type ValidMassExecRequest struct {
	URL           string
	Method        string
	QueryParam    map[string]any
	PathVariables map[string]string
	Headers       map[string]any
	BodyType      httpexec.HTTPRequestBodyType
	Body          any
	ResponseType  string
	Data          ValidExecRequestDataSlice
}

// Validate validates the MassExecRequest
func (r MassExecRequest) Validate(ctr *container.Container, targetType MassExecType) (ValidMassExecRequest, error) {
	var valid ValidMassExecRequest
	var err error
	if r.TargetID == nil {
		return ValidMassExecRequest{}, fmt.Errorf("target_id is required")
	}
	if r.Endpoint == nil {
		return ValidMassExecRequest{}, fmt.Errorf("endpoint is required")
	}
	var urlRoot string
	if urlRoot, err = ctr.TargetContainer.FindTarget(*r.TargetID, config.TargetType(targetType)); err != nil {
		return ValidMassExecRequest{}, fmt.Errorf("failed to find target: %v", err)
	}
	valid.URL = fmt.Sprintf("%s%s", urlRoot, *r.Endpoint)
	if r.Method == nil {
		return ValidMassExecRequest{}, fmt.Errorf("method is required")
	}
	valid.Method = *r.Method

	valid.QueryParam = r.QueryParam
	valid.PathVariables = r.PathVariables
	valid.Headers = r.Headers
	valid.Body = r.Body
	if r.BodyType == nil {
		valid.BodyType = httpexec.DefaultHTTPRequestBodyType
	} else {
		switch httpexec.HTTPRequestBodyType(*r.BodyType) {
		case httpexec.HTTPRequestBodyTypeJSON, httpexec.HTTPRequestBodyTypeForm, httpexec.HTTPRequestBodyTypeMultipart:
			valid.BodyType = httpexec.HTTPRequestBodyTypeJSON
		default:
			return ValidMassExecRequest{}, fmt.Errorf("invalid body_type value: %s", *r.BodyType)
		}
	}
	if r.ResponseType == nil {
		return ValidMassExecRequest{}, fmt.Errorf("response_type is required")
	}
	valid.ResponseType = *r.ResponseType
	for _, d := range r.Data {
		validData, err := d.Validate()
		if err != nil {
			return ValidMassExecRequest{}, fmt.Errorf("failed to validate data: %v", err)
		}
		valid.Data = append(valid.Data, validData)
	}
	return valid, nil
}

// Run runs the MassExec runner
func (r ValidMassExec) Run(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
	threadOnlyStore *sync.Map,
) error {
	switch r.Type {
	case MassExecTypeHTTP:
		return r.runHTTP(ctx, ctr, outputRoot, str)
	}
	return nil
}

func (r ValidMassExec) runHTTP(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
	str *sync.Map,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	concurrentCount := len(r.Requests)

	threadExecutors := make([]*MassiveExecThreadExecutor, concurrentCount)

	for i := 0; i < concurrentCount; i++ {
		request := massExec.Data.Requests[i]
		threadExecutors[i] = &MassiveExecThreadExecutor{
			ID: i + 1,
		}

		if massExec.Output.Enabled {
			logFilePath := fmt.Sprintf("%s/massive_execute_%010d.csv", outputRoot, i+1)
			file, err := os.Create(logFilePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			threadExecutors[i].outputFile = file
			defer threadExecutors[i].Close(ctx)

			writer := csv.NewWriter(file)
			header := []string{"Success", "SendDatetime", "ReceivedDatetime", "Count", "ResponseTime", "StatusCode", "Data"}
			if err := writer.Write(header); err != nil {
				return fmt.Errorf("failed to write header: %v", err)
			}
			writer.Flush()
		}

		endType := massExec.Data.Requests[i].EndpointType
		execType := batchexecmap.NewExecTypeFromString(endType)
		if execType == 0 {
			return fmt.Errorf("failed to parse exec type: %s", endType)
		}
		factor := batchexecmap.TypeFactoryMap[execType]
		// INFO: close on factor.Factory(response handler), because only it will write to this channel
		termChan := make(chan execbatch.TerminateType)
		validatedReq := &execbatch.ValidatedExecRequest{}
		if err := execbatch.ValidateExecReq(ctx, ctr, request.ExecRequest, validatedReq); err != nil {
			return fmt.Errorf("failed to validate execute request: %v", err)
		}
		writeFunc := func(
			ctx context.Context,
			ctr *app.Container,
			id int,
			data execbatch.WriteData,
		) error {
			if !massExec.Output.Enabled {
				return nil
			}
			writer := csv.NewWriter(threadExecutors[i].outputFile)
			ctr.Logger.Debug(ctx, "Writing data to csv",
				logger.Value("id", id), logger.Value("data", data), logger.Value("on", "runAsyncProcessing"))
			if err := writer.Write(data.ToSlice()); err != nil {
				ctr.Logger.Error(ctx, "failed to write data to csv",
					logger.Value("error", err), logger.Value("on", "runAsyncProcessing"))
			}
			writer.Flush()
			return nil
		}
		executor, _, err := factor.Factory(
			ctx,
			ctr,
			i+1,
			validatedReq,
			termChan,
			ctr.AuthToken,
			ctr.Config.Web.API.Url,
			writeFunc,
		)
		ctr.Logger.Info(ctx, "created executor",
			logger.Value("id", i+1), logger.Value("type", endType), logger.Value("executor", executor))
		if err != nil {
			return fmt.Errorf("failed to create executor: %v", err)
		}
		threadExecutors[i].RequestExecutor = executor
		threadExecutors[i].TermChan = termChan
		threadExecutors[i].successBreak = request.SuccessBreak
	}

	var wg sync.WaitGroup
	atomicErr := atomic.Value{}
	var startChan = make(chan struct{})
	for _, executor := range threadExecutors {
		wg.Add(1)
		go func(exec *MassiveExecThreadExecutor) {
			defer wg.Done()
			if err := exec.Execute(ctx, ctr, startChan); err != nil {
				atomicErr.Store(err)
				ctr.Logger.Error(ctx, "failed to execute",
					logger.Value("error", err), logger.Value("id", exec.ID))
				cancel()
				return
			}
		}(executor)
	}
	close(startChan)
	wg.Wait()

	if err := atomicErr.Load(); err != nil {
		ctr.Logger.Error(ctx, "failed to find error",
			logger.Value("error", err.(error)), logger.Value("on", "MassExecuteBatch"))
		return err.(error)
	}

	return nil
}

type MassiveExecThreadExecutor struct {
	ID              int
	outputFile      *os.File
	RequestExecutor httpexec.MassRequestExecutor
	TermChan        chan execbatch.TerminateType
	successBreak    []string
}

func NewMassiveRequestThreadExecutor(
	id int,
	outputFile *os.File,
	req httpexec.MassRequestExecutor,
) *MassiveExecThreadExecutor {
	return &MassiveExecThreadExecutor{
		ID:              id,
		outputFile:      outputFile,
		RequestExecutor: req,
	}
}

func (e *MassiveExecThreadExecutor) Execute(
	ctx context.Context,
	ctr *container.Container,
	startChan <-chan struct{},
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	<-startChan

	ctr.Logger.Info(ctx, "Execute Start",
		logger.Value("ExecutorID", e.ID))
	err := e.RequestExecutor.MassRequestExecute(ctx, ctr)
	if err != nil {
		return err
	}

	termType := <-e.TermChan
	ctr.Logger.Info(ctx, "Execute End For Break",
		logger.Value("ExecuteID", e.ID))
	for _, breakType := range e.successBreak {
		if termType == execbatch.NewTerminateTypeFromString(breakType) {
			ctr.Logger.Info(ctx, "Execute End For Success Break", logger.Value("ExecuteID", e.ID))
			return nil
		}
	}
	return fmt.Errorf("execute End For Fail Break: %s", termType.String())
}

func (e *MassiveExecThreadExecutor) Close(ctx context.Context) error {
	e.outputFile.Close()
	return nil
}
