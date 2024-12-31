package runner

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/ablankz/bloader/internal/encrypt"
	"github.com/ablankz/bloader/internal/logger"
)

// Flow represents the flow runner
type Flow struct {
	Step FlowStep `yaml:"step"`
}

// ValidFlow represents a valid flow runner
type ValidFlow struct {
	Step ValidFlowStep
}

// Validate validates a flow runner
func (r Flow) Validate() (ValidFlow, error) {
	validFlowStep, err := r.Step.Validate()
	if err != nil {
		return ValidFlow{}, err
	}
	return ValidFlow{Step: validFlowStep}, nil
}

// FlowStep represents a flow step
type FlowStep struct {
	Concurrency *int           `yaml:"concurrency"`
	Flows       []FlowStepFlow `yaml:"flows"`
}

// ValidFlowStep represents a valid flow step
type ValidFlowStep struct {
	Concurrency int
	Flows       []ValidFlowStepFlow
}

// Validate validates a flow step
func (r FlowStep) Validate() (ValidFlowStep, error) {
	var validFlowStep ValidFlowStep
	if r.Concurrency == nil {
		validFlowStep.Concurrency = 0
	} else {
		validFlowStep.Concurrency = *r.Concurrency
	}
	idSet := make(map[string]struct{})
	for i, flow := range r.Flows {
		var validFlowStepFlow ValidFlowStepFlow
		if flow.ID == nil {
			return ValidFlowStep{}, fmt.Errorf("id is required")
		}
		if _, ok := idSet[*flow.ID]; ok {
			return ValidFlowStep{}, fmt.Errorf("id %s is duplicated", *flow.ID)
		}
		idSet[*flow.ID] = struct{}{}
		validFlowStepFlow.ID = *flow.ID
		err := flow.Validate(&validFlowStepFlow)
		if err != nil {
			return ValidFlowStep{}, fmt.Errorf("failed to validate flow[%d]: %v", i, err)
		}
		validFlowStep.Flows = append(validFlowStep.Flows, validFlowStepFlow)
	}
	return validFlowStep, nil
}

// FlowStepFlowType represents the flow step flow type
type FlowStepFlowType string

const (
	// FlowStepFlowTypeFile represents the file flow step flow type
	FlowStepFlowTypeFile FlowStepFlowType = "file"
	// FlowStepFlowTypeFlow represents the flow flow step flow type
	FlowStepFlowTypeFlow FlowStepFlowType = "flow"
)

// FlowStepFlow represents a flow step flow
type FlowStepFlow struct {
	ID               *string             `yaml:"id"`
	Type             *string             `yaml:"type"`
	File             *string             `yaml:"file"`
	Mkdir            bool                `yaml:"mkdir"`
	Count            *int                `yaml:"count"`
	Values           []FlowStepFlowValue `yaml:"values"`
	ThreadOnlyValues []FlowStepFlowValue `yaml:"thread_only_values"`
	Flows            []FlowStepFlow      `yaml:"flows"`
	Concurrency      *int                `yaml:"concurrency"`
}

// ValidFlowStepFlow represents a valid flow step flow
type ValidFlowStepFlow struct {
	ID               string
	Type             FlowStepFlowType
	File             string
	Mkdir            bool
	Count            int
	Values           []ValidFlowStepFlowValue
	ThreadOnlyValues []ValidFlowStepFlowValue
	Flows            []ValidFlowStepFlow
	Concurrency      int
}

// FlowStepFlowValue represents a flow step flow value
type FlowStepFlowValue struct {
	Key   *string `yaml:"key"`
	Value *any    `yaml:"value"`
}

// ValidFlowStepFlowValue represents a valid flow step flow value
type ValidFlowStepFlowValue struct {
	Key   string
	Value any
}

// Validate validates a flow step flow value
func (r FlowStepFlowValue) Validate() (ValidFlowStepFlowValue, error) {
	var validFlowStepFlowValue ValidFlowStepFlowValue
	if r.Key == nil {
		return ValidFlowStepFlowValue{}, fmt.Errorf("key is required")
	}
	validFlowStepFlowValue.Key = *r.Key
	if r.Value == nil {
		return ValidFlowStepFlowValue{}, fmt.Errorf("value is required")
	}
	validFlowStepFlowValue.Value = *r.Value
	return validFlowStepFlowValue, nil
}

// Validate validates a flow step flow
func (f FlowStepFlow) Validate(valid *ValidFlowStepFlow) error {
	valid.Mkdir = f.Mkdir
	for i, value := range f.Values {
		valValue, err := value.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate flow value[%d]: %v", i, err)
		}
		valid.Values = append(valid.Values, valValue)
	}
	for i, value := range f.ThreadOnlyValues {
		valValue, err := value.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate flow thread only value[%d]: %v", i, err)
		}
		valid.ThreadOnlyValues = append(valid.ThreadOnlyValues, valValue)
	}
	if f.Type == nil {
		return fmt.Errorf("type is required")
	}
	switch FlowStepFlowType(*f.Type) {
	case FlowStepFlowTypeFile:
		valid.Type = FlowStepFlowType(*f.Type)
		if f.File == nil {
			return fmt.Errorf("file is required")
		}
		valid.File = *f.File
		if f.Count == nil {
			valid.Count = 1
		} else {
			if *f.Count < 0 {
				return fmt.Errorf("count must be greater than or equal to 0")
			}
			valid.Count = *f.Count
		}
	case FlowStepFlowTypeFlow:
		valid.Type = FlowStepFlowType(*f.Type)
		if f.Concurrency == nil {
			valid.Concurrency = 0
		} else {
			valid.Concurrency = *f.Concurrency
		}
		subIdSet := make(map[string]struct{})
		for i, f := range f.Flows {
			var subValid ValidFlowStepFlow
			if f.ID == nil {
				return fmt.Errorf("id is required")
			}
			if _, ok := subIdSet[*f.ID]; ok {
				return fmt.Errorf("id %s is duplicated", *f.ID)
			}
			subIdSet[*f.ID] = struct{}{}
			subValid.ID = *f.ID
			err := f.Validate(&subValid)
			if err != nil {
				return fmt.Errorf("failed to validate flow[%d]: %v", i, err)
			}
			valid.Flows = append(valid.Flows, subValid)
		}
	default:
		return fmt.Errorf("invalid type value: %s", *f.Type)
	}

	return nil
}

type flowExecutor struct {
	flowType        FlowStepFlowType
	filename        string
	rootDir         string
	threadOnlyStore *sync.Map
	concurrency     int
	flows           []ValidFlowStepFlow
	loopCount       int
}

// Run runs a flow step flow
func (f ValidFlow) Run(
	ctx context.Context,
	log logger.Logger,
	encryptCtr encrypt.EncrypterContainer,
	tmplFactor TmplFactor,
	store Store,
	authFactor AuthenticatorFactor,
	outFactor OutputFactor,
	targetFactor TargetFactor,
	str *sync.Map,
	outputRoot string,
	callCount int,
) error {
	return run(
		ctx,
		log,
		encryptCtr,
		tmplFactor,
		store,
		authFactor,
		outFactor,
		targetFactor,
		str,
		outputRoot,
		callCount,
		f.Step.Flows,
		f.Step.Concurrency,
	)
}

func run(
	ctx context.Context,
	log logger.Logger,
	encryptCtr encrypt.EncrypterContainer,
	tmplFactor TmplFactor,
	store Store,
	authFactor AuthenticatorFactor,
	outFactor OutputFactor,
	targetFactor TargetFactor,
	str *sync.Map,
	outputRoot string,
	callCount int,
	flows []ValidFlowStepFlow,
	concurrency int,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sumCount := 0
	for _, flow := range flows {
		if flow.Count <= 0 {
			flow.Count = 1
		}
		sumCount += flow.Count
	}

	executors := make([]flowExecutor, sumCount)

	var count int
	for _, flow := range flows {
		for _, v := range flow.Values {
			str.Store(v.Key, v.Value)
		}
		threadOnlyStore := &sync.Map{}
		for _, v := range flow.ThreadOnlyValues {
			threadOnlyStore.Store(v.Key, v.Value)
		}
		if flow.Count > 1 {
			for j := 0; j < flow.Count; j++ {
				var rootDir string
				if flow.Mkdir {
					rootDir = fmt.Sprintf("%s/%s_%d", outputRoot, flow.ID, j)
				} else {
					rootDir = outputRoot
				}

				executors[count] = flowExecutor{
					flowType:        flow.Type,
					filename:        flow.File,
					rootDir:         rootDir,
					threadOnlyStore: threadOnlyStore,
					concurrency:     flow.Concurrency,
					flows:           flow.Flows,
					loopCount:       j,
				}
				count++
			}
		} else {
			var rootDir string
			if flow.Mkdir {
				rootDir = fmt.Sprintf("%s/%s", outputRoot, flow.ID)
			} else {
				rootDir = outputRoot
			}

			executors[count] = flowExecutor{
				flowType:        flow.Type,
				filename:        flow.File,
				rootDir:         rootDir,
				threadOnlyStore: threadOnlyStore,
				concurrency:     flow.Concurrency,
				flows:           flow.Flows,
				loopCount:       0,
			}
			count++
		}
	}

	var sequential bool
	if concurrency < 0 {
		concurrency = len(executors)
	}
	if concurrency == 0 {
		concurrency = 1
		sequential = true
	}

	if sequential {
		for i, executor := range executors {
			switch executor.flowType {
			case FlowStepFlowTypeFile:
				baseExecutor := BaseExecutor{
					EncryptCtr:   encryptCtr,
					Logger:       log,
					TmplFactor:   tmplFactor,
					Store:        store,
					AuthFactor:   authFactor,
					OutputFactor: outFactor,
					TargetFactor: targetFactor,
				}
				err := baseExecutor.Execute(
					ctx,
					executor.filename,
					str,
					executor.threadOnlyStore,
					executor.rootDir,
					executor.loopCount,
					callCount+1,
				)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
						logger.Value("error", err), logger.Value("on", "Flow"))
					return fmt.Errorf("failed to execute flow: %v", err)
				}
			case FlowStepFlowTypeFlow:
				err := run(
					ctx,
					log,
					encryptCtr,
					tmplFactor,
					store,
					authFactor,
					outFactor,
					targetFactor,
					str,
					executor.rootDir,
					callCount+1,
					executor.flows,
					executor.concurrency,
				)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
						logger.Value("error", err), logger.Value("on", "Flow"))
					return fmt.Errorf("failed to execute flow: %v", err)
				}
				log.Debug(ctx, "flow finished",
					logger.Value("on", "Flow"))
			}
		}
	} else {
		atomicErr := atomic.Value{}
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		for i, executor := range executors {
			wg.Add(1)

			go func(preExecutor flowExecutor) {
				defer wg.Done()

				sem <- struct{}{}

				switch preExecutor.flowType {
				case FlowStepFlowTypeFile:
					baseExecutor := BaseExecutor{
						Logger:       log,
						EncryptCtr:   encryptCtr,
						TmplFactor:   tmplFactor,
						Store:        store,
						AuthFactor:   authFactor,
						OutputFactor: outFactor,
						TargetFactor: targetFactor,
					}
					err := baseExecutor.Execute(
						ctx,
						preExecutor.filename,
						str,
						preExecutor.threadOnlyStore,
						preExecutor.rootDir,
						preExecutor.loopCount,
						callCount+1,
					)
					if err != nil {
						atomicErr.Store(err)
						log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
							logger.Value("error", err), logger.Value("on", "Flow"))
						cancel()
						return
					}
				case FlowStepFlowTypeFlow:
					err := run(
						ctx,
						log,
						encryptCtr,
						tmplFactor,
						store,
						authFactor,
						outFactor,
						targetFactor,
						str,
						preExecutor.rootDir,
						callCount+1,
						preExecutor.flows,
						preExecutor.concurrency,
					)
					if err != nil {
						atomicErr.Store(err)
						log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
							logger.Value("error", err), logger.Value("on", "Flow"))
						cancel()
						return
					}
				}
				log.Debug(ctx, "flow finished",
					logger.Value("on", "Flow"))

				<-sem
			}(executor)
		}

		wg.Wait()

		close(sem)

		if err := atomicErr.Load(); err != nil {
			log.Error(ctx, "failed to find error",
				logger.Value("error", err.(error)), logger.Value("on", "Flow"))
			return err.(error)
		}

		return nil
	}

	return nil
}
