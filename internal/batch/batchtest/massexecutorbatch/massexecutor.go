package massexecutorbatch

import "github.com/LabGroupware/go-measure-tui/internal/batch/batchtest/execbatch"

type MassExecuteData struct {
	Requests []MassExecuteOneData `yaml:"requests"`
}

type MassExecuteOneData struct {
	execbatch.ExecRequest `yaml:",inline"`
	SuccessBreak          []string `yaml:"successBreak"`
}

type MassExecute struct {
	Type   string          `yaml:"type"`
	Data   MassExecuteData `yaml:"data"`
	Output BatchTestOutput `yaml:"output"`
}

type BatchTestOutput struct {
	Enabled bool `yaml:"enabled"`
}
