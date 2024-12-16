package runner

import (
	"fmt"
	"sync"
	"time"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/output"
	"github.com/ablankz/bloader/internal/prompt"
)

func Run(ctr *container.Container, filename string, data map[string]any) error {
	var err error
	if filename == "" {
		filename, err = prompt.PromptText(
			"Enter the file to run the load test",
			false,
		)
		if err != nil {
			return fmt.Errorf("failed to get the file to run the load test: %w", err)
		}
	}

	globalStore := sync.Map{}
	threadOnlyStore := sync.Map{}
	outputCtr := output.NewOutputContainer(ctr.Config.Env, ctr.Config.Outputs)
	ctx := ctr.Ctx

	outputRoot := time.Now().Format("20060102_150405")

	for k, v := range data {
		globalStore.Store(k, v)
	}

	if err := baseExecute(
		ctx,
		ctr,
		filename,
		&globalStore,
		&threadOnlyStore,
		outputRoot,
		outputCtr,
		0,
		0,
	); err != nil {
		return fmt.Errorf("failed to execute the load test: %w", err)
	}

	return nil
}
