package batchtest

import (
	"fmt"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/testprompt"
)

func BatchTest(ctr *app.Container) error {

	filename, err := testprompt.PromptInput("File name")
	if err != nil {
		return fmt.Errorf("failed to get file name: %v", err)
	}

	globalStore := sync.Map{}
	threadOnlyStore := sync.Map{}
	ctx := ctr.Ctx

	testOutput := ctr.Config.Batch.Test.Output
	metricsOutput := ctr.Config.Batch.Metrics.Output

	timestamp := time.Now().Format("20060102_150405")
	testDirPath := fmt.Sprintf("%s/test_%s", testOutput, timestamp)
	metricsDirPath := fmt.Sprintf("%s/test_%s", metricsOutput, timestamp)

	if err := baseExecute(ctx, ctr, filename, &globalStore, &threadOnlyStore, testDirPath, metricsDirPath); err != nil {
		return fmt.Errorf("failed to execute batch test: %v", err)
	}

	return nil
}
