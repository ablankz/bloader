package prefetchbatch

type PrefetchConfig struct {
	Enabled  bool               `yaml:"enabled"`
	Requests []*PrefetchRequest `yaml:"requests"`
}

type PrefetchRequest struct {
	ID            string              `yaml:"id"`
	EndpointType  string              `yaml:"endpointType"`
	QueryParam    map[string][]string `yaml:"queryParam"`
	PathVariables map[string]string   `yaml:"pathVariables"`
	Body          any                 `yaml:"body"`
	Vars          []PrefetchVariable  `yaml:"vars"`
	DependsOn     []string            `yaml:"dependsOn"`
}

type PrefetchVariable struct {
	ID       string `yaml:"id"`
	JMESPath string `yaml:"jmesPath"`
	OnError  string `yaml:"onError"`
}
