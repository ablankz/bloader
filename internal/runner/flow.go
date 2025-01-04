package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	common "github.com/ablankz/go-cmproto"

	"github.com/ablankz/bloader/internal/encrypt"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/utils"
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
		err := flow.Validate(&validFlowStepFlow, idSet)
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
	// FlowStepFlowTypeSlaveCmd represents the slave command flow step flow type
	FlowStepFlowTypeSlaveCmd FlowStepFlowType = "slaveCmd"
)

// FlowStepFlowDependsOn represents the flow step flow depends on
type FlowStepFlowDependsOn struct {
	Flow  *string `yaml:"flow"`
	Event *string `yaml:"event"`
}

// ValidFlowStepFlowDependsOn represents a valid flow step flow depends on
type ValidFlowStepFlowDependsOn struct {
	Flow  string
	Event RunnerEvent
}

// Validate validates a flow step flow depends on
func (r FlowStepFlowDependsOn) Validate() (ValidFlowStepFlowDependsOn, error) {
	var validFlowStepFlowDependsOn ValidFlowStepFlowDependsOn
	if r.Flow == nil {
		return ValidFlowStepFlowDependsOn{}, fmt.Errorf("flow is required")
	}
	validFlowStepFlowDependsOn.Flow = *r.Flow
	if r.Event == nil {
		return ValidFlowStepFlowDependsOn{}, fmt.Errorf("event is required")
	}
	validFlowStepFlowDependsOn.Event = RunnerEvent(*r.Event)
	return validFlowStepFlowDependsOn, nil
}

// FlowStepFlow represents a flow step flow
type FlowStepFlow struct {
	ID               *string                 `yaml:"id"`
	DependsOn        []FlowStepFlowDependsOn `yaml:"depends_on"`
	Type             *string                 `yaml:"type"`
	File             *string                 `yaml:"file"`
	Mkdir            bool                    `yaml:"mkdir"`
	Count            *int                    `yaml:"count"`
	Values           []FlowStepFlowValue     `yaml:"values"`
	ThreadOnlyValues []FlowStepFlowValue     `yaml:"thread_only_values"`
	Flows            []FlowStepFlow          `yaml:"flows"`
	Concurrency      *int                    `yaml:"concurrency"`
	Executors        []FlowStepFlowExecutor  `yaml:"executors"`
}

// ValidFlowStepFlow represents a valid flow step flow
type ValidFlowStepFlow struct {
	ID               string
	DependsOn        []ValidFlowStepFlowDependsOn
	Type             FlowStepFlowType
	File             string
	Mkdir            bool
	Count            int
	Values           []ValidFlowStepFlowValue
	ThreadOnlyValues []ValidFlowStepFlowValue
	Flows            []ValidFlowStepFlow
	Concurrency      int
	Executors        []ValidFlowStepFlowExecutor
	waitFunc         func(ctx context.Context) error
}

// FlowStepFlowExecutorOutput represents a flow step flow executor output
type FlowStepFlowExecutorOutput struct {
	Enabled  bool    `yaml:"enabled"`
	RootPath *string `yaml:"root_path"`
}

// ValidFlowStepFlowExecutorOutput represents a valid flow step flow executor output
type ValidFlowStepFlowExecutorOutput struct {
	Enabled  bool
	RootPath string
}

// Validate validates a flow step flow executor output
func (r FlowStepFlowExecutorOutput) Validate() (ValidFlowStepFlowExecutorOutput, error) {
	if !r.Enabled {
		return ValidFlowStepFlowExecutorOutput{}, nil
	}
	var validFlowStepFlowExecutorOutput ValidFlowStepFlowExecutorOutput
	validFlowStepFlowExecutorOutput.Enabled = r.Enabled
	if r.RootPath == nil {
		return ValidFlowStepFlowExecutorOutput{}, fmt.Errorf("root_path is required")
	}
	validFlowStepFlowExecutorOutput.RootPath = *r.RootPath
	return validFlowStepFlowExecutorOutput, nil
}

// FlowStepFlowExecutor represents a flow step flow executor
type FlowStepFlowExecutor struct {
	SlaveID                    *string                    `yaml:"slave_id"`
	Output                     FlowStepFlowExecutorOutput `yaml:"output"`
	InheritValues              bool                       `yaml:"inherit_values"`
	AdditionalValues           []FlowStepFlowValue        `yaml:"additional_values"`
	AdditionalThreadOnlyValues []FlowStepFlowValue        `yaml:"additional_thread_only_values"`
}

// ValidFlowStepFlowExecutor represents a valid flow step flow executor
type ValidFlowStepFlowExecutor struct {
	SlaveID                    string
	Output                     ValidFlowStepFlowExecutorOutput
	InheritValues              bool
	AdditionalValues           []ValidFlowStepFlowValue
	AdditionalThreadOnlyValues []ValidFlowStepFlowValue
}

// Validate validates a flow step flow executor
func (r FlowStepFlowExecutor) Validate() (ValidFlowStepFlowExecutor, error) {
	var validFlowStepFlowExecutor ValidFlowStepFlowExecutor
	if r.SlaveID == nil {
		return ValidFlowStepFlowExecutor{}, fmt.Errorf("slave_id is required")
	}
	validFlowStepFlowExecutor.SlaveID = *r.SlaveID
	validOutput, err := r.Output.Validate()
	if err != nil {
		return ValidFlowStepFlowExecutor{}, fmt.Errorf("failed to validate output: %v", err)
	}
	validFlowStepFlowExecutor.Output = validOutput
	validFlowStepFlowExecutor.InheritValues = r.InheritValues
	for i, value := range r.AdditionalValues {
		valValue, err := value.Validate()
		if err != nil {
			return ValidFlowStepFlowExecutor{}, fmt.Errorf("failed to validate additional value[%d]: %v", i, err)
		}
		validFlowStepFlowExecutor.AdditionalValues = append(validFlowStepFlowExecutor.AdditionalValues, valValue)
	}
	for i, value := range r.AdditionalThreadOnlyValues {
		valValue, err := value.Validate()
		if err != nil {
			return ValidFlowStepFlowExecutor{}, fmt.Errorf("failed to validate additional thread only value[%d]: %v", i, err)
		}
		validFlowStepFlowExecutor.AdditionalThreadOnlyValues = append(validFlowStepFlowExecutor.AdditionalThreadOnlyValues, valValue)
	}
	return validFlowStepFlowExecutor, nil
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
func (f FlowStepFlow) Validate(valid *ValidFlowStepFlow, idSet map[string]struct{}) error {
	valid.Mkdir = f.Mkdir
	for i, dep := range f.DependsOn {
		validDep, err := dep.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate depends_on[%d]: %v", i, err)
		}
		valid.DependsOn = append(valid.DependsOn, validDep)
	}
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
	case FlowStepFlowTypeSlaveCmd:
		valid.Type = FlowStepFlowType(*f.Type)
		if f.File == nil {
			return fmt.Errorf("file is required")
		}
		valid.File = *f.File
		for i, e := range f.Executors {
			validExecutor, err := e.Validate()
			if err != nil {
				return fmt.Errorf("failed to validate executor[%d]: %v", i, err)
			}
			valid.Executors = append(valid.Executors, validExecutor)
		}
	case FlowStepFlowTypeFlow:
		valid.Type = FlowStepFlowType(*f.Type)
		if f.Concurrency == nil {
			valid.Concurrency = 0
		} else {
			valid.Concurrency = *f.Concurrency
		}
		for i, f := range f.Flows {
			var subValid ValidFlowStepFlow
			if f.ID == nil {
				return fmt.Errorf("id is required")
			}
			if _, ok := idSet[*f.ID]; ok {
				return fmt.Errorf("id %s is duplicated", *f.ID)
			}
			idSet[*f.ID] = struct{}{}
			subValid.ID = *f.ID
			err := f.Validate(&subValid, idSet)
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
	rootStr         *sync.Map
	threadOnlyStore *sync.Map
	concurrency     int
	flows           []ValidFlowStepFlow
	executors       []ValidFlowStepFlowExecutor
	loopCount       int
	waitFunc        func(ctx context.Context) error
	castFunc        func(ctx context.Context) error
	eventCaster     *utils.Broadcaster[RunnerEvent]
}

type closer func() error

func createBroadCastMap(
	flows []ValidFlowStepFlow,
	broadCastMap map[string]*utils.Broadcaster[RunnerEvent],
) ([]closer, error) {
	closeFuncs := make([]closer, 0)
	for _, flow := range flows {
		if _, ok := broadCastMap[flow.ID]; ok {
			return nil, fmt.Errorf("id %s is duplicated", flow.ID)
		}
		broadCastMap[flow.ID] = utils.NewBroadcaster[RunnerEvent]()
		if len(flow.Flows) > 0 {
			cl, err := createBroadCastMap(flow.Flows, broadCastMap)
			if err != nil {
				return nil, err
			}
			closeFuncs = append(closeFuncs, cl...)
		}
	}
	return closeFuncs, nil
}

func attachWaitChan(
	flows []ValidFlowStepFlow,
	broadCastMap map[string]*utils.Broadcaster[RunnerEvent],
) error {
	for i, flow := range flows {
		if len(flow.Flows) > 0 {
			if err := attachWaitChan(flow.Flows, broadCastMap); err != nil {
				return err
			}
		}
		flowEventMap := make(map[string][]RunnerEvent)
		for _, dep := range flow.DependsOn {
			if _, ok := flowEventMap[dep.Flow]; !ok {
				flowEventMap[dep.Flow] = make([]RunnerEvent, 0)
			}
			flowEventMap[dep.Flow] = append(flowEventMap[dep.Flow], dep.Event)
		}
		flowWaitFuncMap := make(map[string]func(ctx context.Context) error)
		for k, v := range flowEventMap {
			caster, ok := broadCastMap[k]
			if !ok {
				return fmt.Errorf("failed to find depends_on %s", k)
			}
			waitChan := caster.Subscribe()
			flowWaitFuncMap[k] = func(ctx context.Context) error {
				mustEvents := v
				for len(mustEvents) > 0 {
					select {
					case event := <-waitChan:
						mustEvents = utils.RemoveElement(mustEvents, event)
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				return nil
			}
		}
		flow.waitFunc = func(ctx context.Context) error {
			for k, f := range flowWaitFuncMap {
				if err := f(ctx); err != nil {
					return fmt.Errorf("failed to wait for %s: %v", k, err)
				}
			}
			return nil
		}

		flows[i] = flow
	}

	return nil
}

// Run runs a flow step flow
func (f *ValidFlow) Run(
	ctx context.Context,
	env string,
	log logger.Logger,
	slaveConCtr *ConnectionContainer,
	encryptCtr encrypt.EncrypterContainer,
	tmplFactor TmplFactor,
	store Store,
	authFactor AuthenticatorFactor,
	outFactor OutputFactor,
	targetFactor TargetFactor,
	str *sync.Map,
	outputRoot string,
	callCount int,
	slaveValues map[string]any,
) error {
	broadCastMap := make(map[string]*utils.Broadcaster[RunnerEvent])
	cl, err := createBroadCastMap(f.Step.Flows, broadCastMap)
	if err != nil {
		return err
	}
	defer func() {
		for _, c := range cl {
			if err := c(); err != nil {
				log.Error(ctx, "failed to close",
					logger.Value("error", err), logger.Value("on", "Flow"))
			}
		}
	}()
	if err := attachWaitChan(f.Step.Flows, broadCastMap); err != nil {
		return err
	}
	return run(
		ctx,
		env,
		log,
		slaveConCtr,
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
		slaveValues,
		broadCastMap,
	)
}

func run(
	ctx context.Context,
	env string,
	log logger.Logger,
	slaveConCtr *ConnectionContainer,
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
	slaveValues map[string]any,
	broadCastMap map[string]*utils.Broadcaster[RunnerEvent],
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
		caster, ok := broadCastMap[flow.ID]
		if !ok {
			return fmt.Errorf("failed to find depends_on %s", flow.ID)
		}
		castFunc := func(ctx context.Context) error {
			caster.Broadcast(RunnerEventTerminated)
			return nil
		}
		rootStr := &sync.Map{}
		for _, v := range flow.Values {
			str.Store(v.Key, v.Value)
			rootStr.Store(v.Key, v.Value)
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
					rootStr:         rootStr,
					threadOnlyStore: threadOnlyStore,
					concurrency:     flow.Concurrency,
					flows:           flow.Flows,
					executors:       flow.Executors,
					loopCount:       j,
					waitFunc:        flow.waitFunc,
					castFunc:        castFunc,
					eventCaster:     caster,
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
				rootStr:         rootStr,
				threadOnlyStore: threadOnlyStore,
				concurrency:     flow.Concurrency,
				flows:           flow.Flows,
				executors:       flow.Executors,
				loopCount:       0,
				waitFunc:        flow.waitFunc,
				castFunc:        castFunc,
				eventCaster:     caster,
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
			if err := executor.waitFunc(ctx); err != nil {
				log.Error(ctx, fmt.Sprintf("failed to wait[%d]", i),
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to wait: %v", err)
			}
			switch executor.flowType {
			case FlowStepFlowTypeFile:
				baseExecutor := BaseExecutor{
					Env:                   env,
					EncryptCtr:            encryptCtr,
					Logger:                log,
					SlaveConnectContainer: slaveConCtr,
					TmplFactor:            tmplFactor,
					Store:                 store,
					AuthFactor:            authFactor,
					OutputFactor:          outFactor,
					TargetFactor:          targetFactor,
				}
				err := baseExecutor.Execute(
					ctx,
					executor.filename,
					str,
					executor.threadOnlyStore,
					executor.rootDir,
					executor.loopCount,
					callCount+1,
					slaveValues,
					NewDefaultEventCasterWithBroadcaster(executor.eventCaster),
				)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
						logger.Value("error", err), logger.Value("on", "Flow"))
					return fmt.Errorf("failed to execute flow: %v", err)
				}
			case FlowStepFlowTypeSlaveCmd:
				err := slaveCmdRun(
					ctx,
					log,
					slaveConCtr,
					outFactor,
					str,
					executor.rootDir,
					flows[i],
				)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
						logger.Value("error", err), logger.Value("on", "Flow"))
					return fmt.Errorf("failed to execute flow: %v", err)
				}
				log.Debug(ctx, "flow finished",
					logger.Value("on", "Flow"))
			case FlowStepFlowTypeFlow:
				err := run(
					ctx,
					env,
					log,
					slaveConCtr,
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
					slaveValues,
					broadCastMap,
				)
				if err != nil {
					log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
						logger.Value("error", err), logger.Value("on", "Flow"))
					return fmt.Errorf("failed to execute flow: %v", err)
				}
				log.Debug(ctx, "flow finished",
					logger.Value("on", "Flow"))
			}

			if err := executor.castFunc(ctx); err != nil {
				log.Error(ctx, fmt.Sprintf("failed to cast[%d]", i),
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to cast: %v", err)
			}
		}
	} else {
		atomicErr := atomic.Value{}
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		for i, executor := range executors {
			if err := executor.waitFunc(ctx); err != nil {
				log.Error(ctx, fmt.Sprintf("failed to wait[%d]", i),
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to wait: %v", err)
			}

			wg.Add(1)

			go func(preExecutor flowExecutor) {
				defer wg.Done()

				defer func() {
					if err := preExecutor.castFunc(ctx); err != nil {
						log.Error(ctx, fmt.Sprintf("failed to cast[%d]", i),
							logger.Value("error", err), logger.Value("on", "Flow"))
						atomicErr.Store(err)
						cancel()
					}
				}()

				sem <- struct{}{}

				switch preExecutor.flowType {
				case FlowStepFlowTypeFile:
					baseExecutor := BaseExecutor{
						Env:                   env,
						Logger:                log,
						SlaveConnectContainer: slaveConCtr,
						EncryptCtr:            encryptCtr,
						TmplFactor:            tmplFactor,
						Store:                 store,
						AuthFactor:            authFactor,
						OutputFactor:          outFactor,
						TargetFactor:          targetFactor,
					}
					err := baseExecutor.Execute(
						ctx,
						preExecutor.filename,
						str,
						preExecutor.threadOnlyStore,
						preExecutor.rootDir,
						preExecutor.loopCount,
						callCount+1,
						slaveValues,
						NewDefaultEventCasterWithBroadcaster(preExecutor.eventCaster),
					)
					if err != nil {
						atomicErr.Store(err)
						log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
							logger.Value("error", err), logger.Value("on", "Flow"))
						cancel()
						return
					}
				case FlowStepFlowTypeSlaveCmd:
					err := slaveCmdRun(
						ctx,
						log,
						slaveConCtr,
						outFactor,
						str,
						preExecutor.rootDir,
						flows[i],
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
						env,
						log,
						slaveConCtr,
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
						slaveValues,
						broadCastMap,
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

func slaveCmdRun(
	ctx context.Context,
	log logger.Logger,
	slaveConCtr *ConnectionContainer,
	outFactor OutputFactor,
	str *sync.Map,
	outputRoot string,
	f ValidFlowStepFlow,
) error {
	globalStr := make(map[string]any)
	threadOnlyStr := make(map[string]any)
	str.Range(func(key, value any) bool {
		globalStr[key.(string)] = value
		return true
	})
	f.ThreadOnlyValues = append(f.ThreadOnlyValues, f.Values...)
	for _, v := range f.ThreadOnlyValues {
		threadOnlyStr[v.Key] = v.Value
	}
	slaveExecutors := make([]slaveExecutor, len(f.Executors))
	for i, exec := range f.Executors {
		slaveID := exec.SlaveID
		mapData, ok := slaveConCtr.Find(slaveID)
		if !ok {
			log.Error(ctx, fmt.Sprintf("failed to find slave: %s", slaveID),
				logger.Value("on", "Flow"))
			return fmt.Errorf("failed to find slave: %s", slaveID)
		}
		if exec.InheritValues {
			str.Range(func(key, value any) bool {
				globalStr[key.(string)] = value
				return true
			})
		}
		for _, v := range exec.AdditionalValues {
			globalStr[v.Key] = v.Value
		}
		for _, v := range exec.AdditionalThreadOnlyValues {
			threadOnlyStr[v.Key] = v.Value
		}
		defaultStr, err := common.NewFlexMap(globalStr)
		if err != nil {
			log.Error(ctx, "failed to create default store",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to create default store: %v", err)
		}
		defaultThreadOnlyStr, err := common.NewFlexMap(threadOnlyStr)
		if err != nil {
			log.Error(ctx, "failed to create default thread only store",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to create default thread only store: %v", err)
		}
		oRoot := outputRoot
		if exec.Output.Enabled {
			oRoot = oRoot + "/" + exec.Output.RootPath
		}
		slaveValuesMap := map[string]any{
			"SlaveID": slaveID,
			"Index":   i,
		}
		slaveValues, err := common.NewFlexMap(slaveValuesMap)
		if err != nil {
			log.Error(ctx, "failed to create",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to create slave values: %v", err)
		}
		res, err := mapData.Cli.SlaveCommand(ctx, &pb.SlaveCommandRequest{
			ConnectionId: mapData.ConnectionID,
			LoaderId:     f.File,
			OutputRoot:   oRoot,
		})
		if err != nil {
			log.Error(ctx, "failed to execute",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to execute slave command: %v", err)
		}

		stream, err := mapData.Cli.SlaveCommandDefaultStore(ctx)
		if err != nil {
			log.Error(ctx, "failed to execute",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to execute slave command default store: %v", err)
		}
		defaultStrBytes, err := json.Marshal(defaultStr)
		if err != nil {
			log.Error(ctx, "failed to marshal",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to marshal default store: %v", err)
		}
		for i := 0; i < len(defaultStrBytes); i += defaultChunkSize {
			end := i + defaultChunkSize
			if end > len(defaultStrBytes) {
				end = len(defaultStrBytes)
			}
			if err := stream.Send(&pb.SlaveCommandDefaultStoreRequest{
				ConnectionId: mapData.ConnectionID,
				CommandId:    res.CommandId,
				StoreType:    pb.SlaveCommandDefaultStoreType_SLAVE_COMMAND_DEFAULT_STORE_TYPE_STORE,
				DefaultStore: defaultStrBytes[i:end],
				IsLastChunk:  end == len(defaultStrBytes),
			}); err != nil {
				log.Error(ctx, "failed to send",
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to send slave command default store request: %v", err)
			}
		}
		defaultThreadOnlyStrBytes, err := json.Marshal(defaultThreadOnlyStr)
		if err != nil {
			log.Error(ctx, "failed to marshal",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to marshal default thread only store: %v", err)
		}
		for i := 0; i < len(defaultThreadOnlyStrBytes); i += defaultChunkSize {
			end := i + defaultChunkSize
			if end > len(defaultThreadOnlyStrBytes) {
				end = len(defaultThreadOnlyStrBytes)
			}
			if err := stream.Send(&pb.SlaveCommandDefaultStoreRequest{
				ConnectionId: mapData.ConnectionID,
				CommandId:    res.CommandId,
				StoreType:    pb.SlaveCommandDefaultStoreType_SLAVE_COMMAND_DEFAULT_STORE_TYPE_THREAD_ONLY_STORE,
				DefaultStore: defaultThreadOnlyStrBytes[i:end],
				IsLastChunk:  end == len(defaultThreadOnlyStrBytes),
			}); err != nil {
				log.Error(ctx, "failed to send",
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to send slave command default store request: %v", err)
			}
		}
		defaultSlaveValuesStrBytes, err := json.Marshal(slaveValues)
		if err != nil {
			log.Error(ctx, "failed to marshal",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to marshal slave values: %v", err)
		}
		for i := 0; i < len(defaultSlaveValuesStrBytes); i += defaultChunkSize {
			end := i + defaultChunkSize
			if end > len(defaultSlaveValuesStrBytes) {
				end = len(defaultSlaveValuesStrBytes)
			}
			if err := stream.Send(&pb.SlaveCommandDefaultStoreRequest{
				ConnectionId: mapData.ConnectionID,
				CommandId:    res.CommandId,
				StoreType:    pb.SlaveCommandDefaultStoreType_SLAVE_COMMAND_DEFAULT_STORE_TYPE_SLAVE_VALUES,
				DefaultStore: defaultSlaveValuesStrBytes[i:end],
				IsLastChunk:  end == len(defaultSlaveValuesStrBytes),
			}); err != nil {
				log.Error(ctx, "failed to send",
					logger.Value("error", err), logger.Value("on", "Flow"))
				return fmt.Errorf("failed to send slave command default store request: %v", err)
			}
		}
		if _, err := stream.CloseAndRecv(); err != nil {
			log.Error(ctx, "failed to receive",
				logger.Value("error", err), logger.Value("on", "Flow"))
			return fmt.Errorf("failed to receive slave command default store response: %v", err)
		}

		slaveExecutors[i] = slaveExecutor{
			slaveID:       slaveID,
			cmdID:         res.CommandId,
			mapData:       mapData,
			outputEnabled: exec.Output.Enabled,
			outFactor:     outFactor,
		}
	}

	atomicErr := atomic.Value{}
	var wg sync.WaitGroup
	for i, executor := range slaveExecutors {
		wg.Add(1)

		go func(preExecutor slaveExecutor) {
			defer wg.Done()

			err := preExecutor.exec(ctx, log)
			if err != nil {
				atomicErr.Store(err)
				log.Error(ctx, fmt.Sprintf("failed to execute flow[%d]", i),
					logger.Value("error", err), logger.Value("on", "Flow"))
				return
			}
		}(executor)
	}

	wg.Wait()

	if err := atomicErr.Load(); err != nil {
		log.Error(ctx, "failed to find error",
			logger.Value("error", err.(error)), logger.Value("on", "Flow"))
		return err.(error)
	}

	return nil
}

type slaveExecutor struct {
	slaveID       string
	cmdID         string
	mapData       *ConnectionMapData
	outputEnabled bool
	outFactor     OutputFactor
}

// exec executes a slave command
func (e slaveExecutor) exec(
	ctx context.Context,
	log logger.Logger,
) error {
	stream, err := e.mapData.Cli.CallExec(ctx, &pb.CallExecRequest{
		ConnectionId: e.mapData.ConnectionID,
		CommandId:    e.cmdID,
	})
	if err != nil {
		log.Error(ctx, "failed to call exec",
			logger.Value("error", err), logger.Value("on", "Flow"))
		return fmt.Errorf("failed to call exec: %v", err)
	}

	go func() {
		first := true
		var httpDataWriter output.HTTPDataWrite
		var closer output.Close
		defer func() {
			if closer != nil {
				if err := closer(); err != nil {
					log.Error(ctx, "failed to close",
						logger.Value("error", err), logger.Value("on", "Flow"))
				}
			}
		}()
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				if err := stream.CloseSend(); err != nil {
					log.Error(ctx, "failed to close send",
						logger.Value("error", err), logger.Value("on", "Flow"))
				}

				break
			}
			if err != nil {
				log.Error(ctx, "failed to receive exec",
					logger.Value("error", err), logger.Value("on", "Flow"))
				if err := stream.CloseSend(); err != nil {
					log.Error(ctx, "failed to close send",
						logger.Value("error", err), logger.Value("on", "Flow"))
				}
				return
			}
			if !e.outputEnabled {
				continue
			}
			output, err := e.outFactor.Factorize(ctx, res.OutputId)
			if err != nil {
				log.Error(ctx, "failed to factorize output",
					logger.Value("error", err), logger.Value("on", "Flow"))
				if err := stream.CloseSend(); err != nil {
					log.Error(ctx, "failed to close send",
						logger.Value("error", err), logger.Value("on", "Flow"))
				}
				return
			}
			switch res.OutputType {
			case pb.CallExecOutputType_CALL_EXEC_OUTPUT_TYPE_HTTP:
				httpOut := res.GetOutputHttp()
				if first {
					if httpDataWriter, closer, err = output.HTTPDataWriteFactory(
						ctx,
						log,
						true,
						res.OutputRoot,
						httpOut.Data,
					); err != nil {
						log.Error(ctx, "failed to create http data writer",
							logger.Value("error", err), logger.Value("on", "Flow"))
						if err := stream.CloseSend(); err != nil {
							log.Error(ctx, "failed to close send",
								logger.Value("error", err), logger.Value("on", "Flow"))
						}
						return
					}
					first = false
					continue
				}
				if err := httpDataWriter(
					ctx,
					log,
					httpOut.Data,
				); err != nil {
					log.Error(ctx, "failed to write http data",
						logger.Value("error", err), logger.Value("on", "Flow"))
					if err := stream.CloseSend(); err != nil {
						log.Error(ctx, "failed to close send",
							logger.Value("error", err), logger.Value("on", "Flow"))
					}
					return
				}
			default:
				log.Error(ctx, "invalid output type",
					logger.Value("on", "Flow"))
				if err := stream.CloseSend(); err != nil {
					log.Error(ctx, "failed to close send",
						logger.Value("error", err), logger.Value("on", "Flow"))
				}
				return
			}
		}
	}()
	termRes, err := e.mapData.Cli.ReceiveLoadTermChannel(ctx, &pb.ReceiveLoadTermChannelRequest{
		ConnectionId: e.mapData.ConnectionID,
		CommandId:    e.cmdID,
	})
	if err != nil {
		log.Error(ctx, "failed to receive term channel",
			logger.Value("error", err), logger.Value("on", "Flow"))
		return fmt.Errorf("failed to receive term channel: %v", err)
	}
	if !termRes.Success {
		log.Error(ctx, "failed to receive term channel",
			logger.Value("on", "Flow"))
		return fmt.Errorf("failed to receive term channel: %s", e.slaveID)
	}

	return nil
}
