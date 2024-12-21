package runner

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ablankz/bloader/internal/auth"
	"github.com/ablankz/bloader/internal/config"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/executor/httpexec"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/runner/matcher"
	"github.com/ablankz/bloader/internal/utils"
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
	for i, req := range r.Requests {
		validRequest, err := req.Validate(ctr, massExecType)
		if err != nil {
			return ValidMassExec{}, fmt.Errorf("failed to validate request[%d]: %v", i, err)
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

// MassExecRequestBreak represents the break configuration for the MassExec runner
type MassExecRequestBreak struct {
	Time         *string                      `yaml:"time"`
	Count        *int                         `yaml:"count"`
	SysError     bool                         `yaml:"sys_error"`
	ParseError   bool                         `yaml:"parse_error"`
	WriteError   bool                         `yaml:"write_error"`
	StatusCode   matcher.StatusCodeConditions `yaml:"status_code"`
	ResponseBody matcher.BodyConditions       `yaml:"response_body"`
}

// ValidMassExecRequestBreak represents the valid break configuration for the MassExec runner
type ValidMassExecRequestBreak struct {
	Time struct {
		Enabled bool
		Time    time.Duration
	}
	Count               httpexec.RequestCountLimit
	SysError            bool
	ParseError          bool
	WriteError          bool
	StatusCodeMatcher   matcher.StatusCodeConditionsMatcher
	ResponseBodyMatcher matcher.BodyConditionsMatcher
}

// Validate validates the MassExecRequestBreak
func (b MassExecRequestBreak) Validate(ctx context.Context, ctr *container.Container) (ValidMassExecRequestBreak, error) {
	var valid ValidMassExecRequestBreak
	var err error
	if b.Time != nil {
		duration, err := time.ParseDuration(*b.Time)
		if err != nil {
			return ValidMassExecRequestBreak{}, fmt.Errorf("failed to parse time: %v", err)
		}
		valid.Time.Enabled = true
		valid.Time.Time = duration
	}
	if b.Count != nil {
		valid.Count.Enabled = true
		valid.Count.Count = *b.Count
	}
	valid.SysError = b.SysError
	valid.ParseError = b.ParseError
	valid.WriteError = b.WriteError
	if valid.StatusCodeMatcher, err = b.StatusCode.MatcherGenerate(ctx, ctr); err != nil {
		return ValidMassExecRequestBreak{}, fmt.Errorf("failed to generate status code matcher: %v", err)
	}
	if valid.ResponseBodyMatcher, err = b.ResponseBody.MatcherGenerate(ctx, ctr); err != nil {
		return ValidMassExecRequestBreak{}, fmt.Errorf("failed to generate response body matcher: %v", err)
	}
	return valid, nil
}

// MassExecRequestRecordExcludeFilter represents the record exclude filter configuration for the MassExec runner
type MassExecRequestRecordExcludeFilter struct {
	Count        matcher.CountConditions      `yaml:"count"`
	StatusCode   matcher.StatusCodeConditions `yaml:"status_code"`
	ResponseBody matcher.BodyConditions       `yaml:"response_body"`
}

// ValidMassExecRequestRecordExcludeFilter represents the valid record exclude filter configuration for the MassExec runner
type ValidMassExecRequestRecordExcludeFilter struct {
	CountFilter        matcher.CountConditionsMatcher
	StatusCodeFilter   matcher.StatusCodeConditionsMatcher
	ResponseBodyFilter matcher.BodyConditionsMatcher
}

// Validate validates the MassExecRequestRecordExcludeFilter
func (f MassExecRequestRecordExcludeFilter) Validate(ctx context.Context, ctr *container.Container) (ValidMassExecRequestRecordExcludeFilter, error) {
	var valid ValidMassExecRequestRecordExcludeFilter
	var err error
	if valid.CountFilter, err = f.Count.MatcherGenerate(ctx, ctr); err != nil {
		return ValidMassExecRequestRecordExcludeFilter{}, fmt.Errorf("failed to generate count filter: %v", err)
	}
	if valid.StatusCodeFilter, err = f.StatusCode.MatcherGenerate(ctx, ctr); err != nil {
		return ValidMassExecRequestRecordExcludeFilter{}, fmt.Errorf("failed to generate status code filter: %v", err)
	}
	if valid.ResponseBodyFilter, err = f.ResponseBody.MatcherGenerate(ctx, ctr); err != nil {
		return ValidMassExecRequestRecordExcludeFilter{}, fmt.Errorf("failed to generate response body filter: %v", err)
	}
	return valid, nil
}

// MassExecRequest represents the request configuration for the MassExec runner
type MassExecRequest struct {
	TargetID            *string                            `yaml:"target_id"`
	Endpoint            *string                            `yaml:"endpoint"`
	Method              *string                            `yaml:"method"`
	QueryParam          map[string]any                     `yaml:"query_param"`
	PathVariables       map[string]string                  `yaml:"path_variables"`
	Headers             map[string]any                     `yaml:"headers"`
	BodyType            *string                            `yaml:"body_type"`
	Body                any                                `yaml:"body"`
	ResponseType        *string                            `yaml:"response_type"`
	Data                []ExecRequestData                  `yaml:"data"`
	Interval            *string                            `yaml:"interval"`
	AwaitPrevResp       bool                               `yaml:"await_prev_response"`
	SuccessBreak        []string                           `yaml:"success_break"`
	Break               MassExecRequestBreak               `yaml:"break"`
	RecordExcludeFilter MassExecRequestRecordExcludeFilter `yaml:"record_exclude_filter"`
}

// ValidMassExecRequest represents the valid request configuration for the MassExec runner
type ValidMassExecRequest struct {
	URL                 string
	Method              string
	QueryParam          map[string]any
	PathVariables       map[string]string
	Headers             map[string]any
	BodyType            httpexec.HTTPRequestBodyType
	Body                any
	ResponseType        string
	Data                ValidExecRequestDataSlice
	Interval            time.Duration
	AwaitPrevResp       bool
	SuccessBreak        matcher.TerminateTypeAndParamsSlice
	Break               ValidMassExecRequestBreak
	RecordExcludeFilter ValidMassExecRequestRecordExcludeFilter
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
	for i, d := range r.Data {
		validData, err := d.Validate()
		if err != nil {
			return ValidMassExecRequest{}, fmt.Errorf("failed to validate data[%d]: %v", i, err)
		}
		valid.Data = append(valid.Data, validData)
	}
	if r.Interval == nil {
		return ValidMassExecRequest{}, fmt.Errorf("interval is required")
	}
	if valid.Interval, err = time.ParseDuration(*r.Interval); err != nil {
		return ValidMassExecRequest{}, fmt.Errorf("failed to parse interval: %v", err)
	}
	valid.AwaitPrevResp = r.AwaitPrevResp
	if valid.SuccessBreak, err = matcher.NewTerminateTypeAndParamsSliceFromStringSlice(r.SuccessBreak); err != nil {
		return ValidMassExecRequest{}, fmt.Errorf("failed to parse success break: %v", err)
	}
	if valid.Break, err = r.Break.Validate(ctr.Ctx, ctr); err != nil {
		return ValidMassExecRequest{}, fmt.Errorf("failed to validate break: %v", err)
	}
	if valid.RecordExcludeFilter, err = r.RecordExcludeFilter.Validate(ctr.Ctx, ctr); err != nil {
		return ValidMassExecRequest{}, fmt.Errorf("failed to validate record exclude filter: %v", err)
	}
	return valid, nil
}

// Run runs the MassExec runner
func (r ValidMassExec) Run(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
) error {
	switch r.Type {
	case MassExecTypeHTTP:
		return r.runHTTP(ctx, ctr, outputRoot)
	}
	return nil
}

func (r ValidMassExec) runHTTP(
	ctx context.Context,
	ctr *container.Container,
	outputRoot string,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	concurrentCount := len(r.Requests)

	threadExecutors := make([]*MassiveExecThreadExecutor, concurrentCount)
	uniqueName := fmt.Sprintf("%s/%s", outputRoot, utils.GenerateUniqueID())

	for i := 0; i < concurrentCount; i++ {
		request := r.Requests[i]
		threadExecutors[i] = &MassiveExecThreadExecutor{
			ID: i + 1,
		}

		req := httpexec.HTTPRequest{
			Method:        request.Method,
			URL:           request.URL,
			Headers:       request.Headers,
			QueryParams:   request.QueryParam,
			PathVariables: request.PathVariables,
			BodyType:      request.BodyType,
			Body:          request.Body,
			AttachRequestInfo: func(ctx context.Context, ctr *container.Container, req *http.Request) error {
				(*r.Auth).SetOnRequest(ctx, ctr.Store, req)
				return nil
			},
		}
		resChan := make(chan httpexec.ResponseContent)
		exe := httpexec.MassRequestContent[httpexec.HTTPRequest]{
			Req:          req,
			Interval:     request.Interval,
			ResponseWait: request.AwaitPrevResp,
			ResChan:      resChan,
			CountLimit:   request.Break.Count,
			ResponseType: httpexec.ResponseType(request.ResponseType),
		}

		writers := make([]output.HTTPDataWrite, 0)
		uName := fmt.Sprintf("%s_%d", uniqueName, i+1)
		var writeCloser []output.Close
		for _, o := range r.Output {
			writer, close, err := o.HTTPDataWriteFactory(
				ctx,
				ctr,
				true,
				uName,
				append(
					[]string{
						"Success",
						"SendDatetime",
						"ReceivedDatetime",
						"Count",
						"ResponseTime",
						"StatusCode",
					},
					request.Data.ExtractHeader()...,
				),
			)
			if err != nil {
				return fmt.Errorf("failed to create writer: %v", err)
			}
			writeCloser = append(writeCloser, close)
			writers = append(writers, writer)
		}

		closer := func() error {
			for _, c := range writeCloser {
				c()
			}
			close(resChan)
			return nil
		}

		termChan := make(chan termChanType)

		threadExecutors[i].closer = closer
		threadExecutors[i].RequestExecutor = exe
		threadExecutors[i].TermChan = termChan
		threadExecutors[i].successBreak = request.SuccessBreak

		consumer := func(
			ctx context.Context,
			ctr *container.Container,
			id int,
			data WriteData,
		) error {
			var additionalData []string
			for _, d := range request.Data {
				result, err := d.Extractor.Extract(data.RawData)
				if err != nil {
					return fmt.Errorf("failed to extract data: %v", err)
				}
				additionalData = append(additionalData, fmt.Sprint(result))
			}

			for _, w := range writers {
				if err := w(ctx, ctr, append(data.ToSlice(), additionalData...)); err != nil {
					return fmt.Errorf("failed to write data: %v", err)
				}
			}

			return nil
		}

		RunAsyncProcessing(
			ctx,
			ctr,
			threadExecutors[i].ID,
			r.Requests[i],
			termChan,
			resChan,
			consumer,
		)
	}

	var wg sync.WaitGroup
	atomicErr := atomic.Value{}
	var startChan = make(chan struct{})
	for _, executor := range threadExecutors {
		wg.Add(1)
		go func(exec *MassiveExecThreadExecutor) {
			defer func() {
				if err := exec.Close(ctx); err != nil {
					ctr.Logger.Error(ctx, "failed to close",
						logger.Value("error", err), logger.Value("id", exec.ID))
				}
			}()
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

// termChanType represents the type of termChan
type termChanType struct {
	termType matcher.TerminateType
	param    string
}

// NewTermChanType creates a new termChanType
func NewTermChanType(termType matcher.TerminateType, param string) termChanType {
	return termChanType{
		termType: termType,
		param:    param,
	}
}

// MassiveExecThreadExecutor represents the thread executor for the MassExec runner
type MassiveExecThreadExecutor struct {
	ID              int
	RequestExecutor httpexec.MassRequestExecutor
	TermChan        chan termChanType
	successBreak    matcher.TerminateTypeAndParamsSlice
	closer          func() error
}

// NewMassiveRequestThreadExecutor creates a new MassiveExecThreadExecutor
func NewMassiveRequestThreadExecutor(
	id int,
	closer func() error,
	req httpexec.MassRequestExecutor,
) *MassiveExecThreadExecutor {
	return &MassiveExecThreadExecutor{
		ID:              id,
		closer:          closer,
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

	fmt.Println("Execute Start", e.ID)

	ctr.Logger.Info(ctx, "Execute Start",
		logger.Value("ExecutorID", e.ID))
	err := e.RequestExecutor.MassRequestExecute(ctx, ctr)
	if err != nil {
		return err
	}

	termType := <-e.TermChan
	ctr.Logger.Info(ctx, "Execute End For Break",
		logger.Value("ExecuteID", e.ID))
	if e.successBreak.Match(termType.termType, termType.param) {
		fmt.Println("Execute End For Success Break", e.ID)
		ctr.Logger.Info(ctx, "Execute End For Success Break", logger.Value("ExecuteID", e.ID))
		return nil
	}
	return fmt.Errorf("execute End For Fail Break: %v(%v)", termType.termType, termType.param)
}

func (e *MassiveExecThreadExecutor) Close(ctx context.Context) error {
	if e.closer != nil {
		return e.closer()
	}
	return nil
}
