package socketsubscribe

type SocketConnectConfig struct {
	Type        string                             `yaml:"type"`
	Output      BatchTestOutput                    `yaml:"output"`
	ID          string                             `yaml:"id"`
	SuccessTerm []string                           `yaml:"successTerm"`
	Term        SocketSubscribeTermConditionConfig `yaml:"termCondition"`
}

type SocketSubscribeConfig struct {
	Type                        string                               `yaml:"type"`
	Output                      BatchTestOutput                      `yaml:"output"`
	ID                          string                               `yaml:"id"`
	ConnectID                   string                               `yaml:"connectId"`
	SelfEventFilter             SocketSubscribeSelfEventFilterConfig `yaml:"selfEventFilter"`
	Subscribe                   SocketSubscribeSubscribeConfig       `yaml:"subscribe"`
	SuccessUnsubscribeActionIDs []string                             `yaml:"successUnsubscribeActionIds"`
	Actions                     []SocketSubscribeActionConfig        `yaml:"actions"`
}

type SocketSubscribeSelfEventFilterConfig struct {
	JMESPath string `yaml:"jmesPath"`
}

type SocketConnectAndSubscribeConfig struct {
	Type        string                             `yaml:"type"`
	Output      BatchTestOutput                    `yaml:"output"`
	Subscribes  []SocketSubscribeSubscribeConfig   `yaml:"subscribes"`
	Actions     []SocketSubscribeActionConfig      `yaml:"actions"`
	SuccessTerm []string                           `yaml:"successTerm"`
	Term        SocketSubscribeTermConditionConfig `yaml:"termCondition"`
}

type BatchTestOutput struct {
	Enabled bool `yaml:"enabled"`
}

type SocketSubscribeSubscribeConfig struct {
	AggregateType string   `yaml:"aggregateType"`
	AggregateId   []string `yaml:"aggregateId"`
	EventTypes    []string `yaml:"eventTypes"`
}

type SocketSubscribeActionConfig struct {
	ID         string                            `yaml:"id"`
	Types      []string                          `yaml:"types"`
	EventTypes []string                          `yaml:"eventTypes"`
	Data       []SocketSubscribeActionDataConfig `yaml:"data"`
}

type SocketSubscribeActionDataConfig struct {
	Key      string `yaml:"key"`
	JMESPath string `yaml:"jmesPath"`
	OnNil    string `yaml:"onNil"`
	OnError  string `yaml:"onError"`
}

type SocketSubscribeTermConditionConfig struct {
	Time  *string                                   `yaml:"time"`
	Error []string                                  `yaml:"error"`
	Event []SocketSubscribeTermConditionEventConfig `yaml:"event"`
}

type SocketSubscribeTermConditionEventConfig struct {
	Types    []string `yaml:"types"`
	Success  bool     `yaml:"success"`
	JMESPath string   `yaml:"jmesPath"`
}

type ErrorTypeForTerm string

const (
	ErrorTypeForTermParseError     ErrorTypeForTerm = "parse_error"
	ErrorTypeForTermUnmarshalError ErrorTypeForTerm = "unmarshal_error"
	ErrorTypeForTermReadError      ErrorTypeForTerm = "read_error"
	ErrorTypeForTermSendError      ErrorTypeForTerm = "send_error"
)

func ContainsTermError(conditions []string, termError ErrorTypeForTerm) bool {
	for _, condition := range conditions {
		if condition == string(termError) {
			return true
		}
	}
	return false
}

type SuccessTerm string

const (
	SuccessTermClose SuccessTerm = "close"
	SuccessTermTime  SuccessTerm = "time"
	SuccessTermError SuccessTerm = "error"
	SuccessTermEvent SuccessTerm = "event"
	SuccessTermData  SuccessTerm = "data"
)

func ContainsSuccessTerm(terms []string, term SuccessTerm) bool {
	for _, t := range terms {
		if t == string(term) {
			return true
		}
	}
	return false
}

type SocketActionType string

const (
	SocketActionTypeStore       SocketActionType = "store"
	SocketActionTypeOutput      SocketActionType = "output"
	SocketActionTypeUnsubscribe SocketActionType = "unsubscribe"
)

func ContainsSocketActionType(types []string, actionType ...SocketActionType) bool {
	for _, t := range types {
		for _, at := range actionType {
			if t == string(at) {
				return true
			}
		}
	}
	return false
}
