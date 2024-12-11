package oneexecbatch

type OneExecuteConfig struct {
	Type    string          `yaml:"type"`
	Request *OneExecRequest `yaml:"request"`
}

type OneExecRequest struct {
	EndpointType  string               `yaml:"endpointType"`
	Body          any                  `yaml:"body"`
	QueryParam    map[string][]string  `yaml:"queryParam"`
	PathVariables map[string]string    `yaml:"pathVariables"`
	Outputs       []OneExecuteVariable `yaml:"outputs"`
}

type OneExecuteVariable struct {
	ID       string `yaml:"id"`
	JMESPath string `yaml:"jmesPath"`
	OnError  string `yaml:"onError"`
}
