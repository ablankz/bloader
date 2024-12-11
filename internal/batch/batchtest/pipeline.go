package batchtest

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
)

type PipelineConfig struct {
	Type        string               `yaml:"type"`
	Concurrency int                  `yaml:"concurrency"`
	Files       []PipelineFileConfig `yaml:"files"`
}

type PipelineFileConfig struct {
	ID               string                        `yaml:"id"`
	File             string                        `yaml:"file"`
	Count            int                           `yaml:"count"`
	NoLoopOverride   bool                          `yaml:"noLoopOverride"`
	ThreadOnlyValues []PipelineFileThreadOnlyValue `yaml:"threadOnlyValues"`
}

type PipelineFileThreadOnlyValue struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

type executeRequest struct {
	id              string
	filename        string
	testRootDir     string
	metricsRootDir  string
	threadOnlyStore *sync.Map
}

func pipelineBatch(
	ctx context.Context,
	ctr *app.Container,
	conf PipelineConfig,
	store *sync.Map,
	rootLoopCount string,
	testOutput,
	metricsOutput string,
) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sumCount := 0
	for _, f := range conf.Files {
		if f.Count <= 0 {
			f.Count = 1
		}
		sumCount += f.Count
	}

	requests := make([]executeRequest, sumCount)

	var count int
	for _, f := range conf.Files {
		if f.Count > 1 {
			for j := 0; j < f.Count; j++ {

				testDirPath := fmt.Sprintf("%s/%s_%d", testOutput, f.ID, j)
				metricsDirPath := fmt.Sprintf("%s/%s_%d", metricsOutput, f.ID, j)

				requests[count] = executeRequest{
					id:              fmt.Sprintf("%s_%d", f.ID, j),
					filename:        f.File,
					testRootDir:     testDirPath,
					metricsRootDir:  metricsDirPath,
					threadOnlyStore: &sync.Map{},
				}
				if f.NoLoopOverride {
					requests[count].threadOnlyStore.Store("loopCount", rootLoopCount)

					if f.ThreadOnlyValues != nil {
						for _, v := range f.ThreadOnlyValues {
							requests[count].threadOnlyStore.Store(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
						}
					}
				} else {
					requests[count].threadOnlyStore.Store("loopCount", fmt.Sprintf("%d", j))

					if f.ThreadOnlyValues != nil {
						for _, v := range f.ThreadOnlyValues {
							requests[count].threadOnlyStore.Store(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
						}
					}

					count++
				}
			}
		} else {
			testDirPath := fmt.Sprintf("%s/%s", testOutput, f.ID)
			metricsDirPath := fmt.Sprintf("%s/%s", metricsOutput, f.ID)
			requests[count] = executeRequest{
				id:              f.ID,
				filename:        f.File,
				testRootDir:     testDirPath,
				metricsRootDir:  metricsDirPath,
				threadOnlyStore: &sync.Map{},
			}

			if f.NoLoopOverride {
				requests[count].threadOnlyStore.Store("loopCount", rootLoopCount)

				if f.ThreadOnlyValues != nil {
					for _, v := range f.ThreadOnlyValues {
						requests[count].threadOnlyStore.Store(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
					}
				}
			} else {
				requests[count].threadOnlyStore.Store("loopCount", fmt.Sprintf("%d", 0))

				if f.ThreadOnlyValues != nil {
					for _, v := range f.ThreadOnlyValues {
						requests[count].threadOnlyStore.Store(fmt.Sprintf("%v", v.Key), fmt.Sprintf("%v", v.Value))
					}
				}
			}

			count++
		}
	}

	var sequential bool
	concurrency := conf.Concurrency
	if concurrency < 0 {
		concurrency = len(requests)
	}
	if concurrency == 0 {
		concurrency = 1
		sequential = true
	}

	if sequential {
		for _, req := range requests {

			err := baseExecute(ctx, ctr, req.filename, store, req.threadOnlyStore, req.testRootDir, req.metricsRootDir)

			if err != nil {
				ctr.Logger.Error(ctr.Ctx, "failed to execute request",
					logger.Value("error", err), logger.Value("on", "PipelineBatch"))
				return fmt.Errorf("failed to execute request: %v", err)
			}

			ctr.Logger.Debug(ctr.Ctx, "request finished",
				logger.Value("on", "PipelineBatch"))
		}

		return nil
	} else {
		atomicErr := atomic.Value{}
		var wg sync.WaitGroup
		sem := make(chan struct{}, concurrency)
		for _, req := range requests {
			wg.Add(1)

			go func(preReq executeRequest) {
				defer wg.Done()

				sem <- struct{}{}

				err := baseExecute(ctx, ctr, preReq.filename, store, preReq.threadOnlyStore, preReq.testRootDir, preReq.metricsRootDir)

				if err != nil {
					atomicErr.Store(err)
					ctr.Logger.Error(ctr.Ctx, "failed to execute request",
						logger.Value("error", err), logger.Value("on", "PipelineBatch"))
					cancel()
					return
				}
				ctr.Logger.Debug(ctr.Ctx, "request finished",
					logger.Value("on", "PipelineBatch"))

				<-sem
			}(req)
		}

		wg.Wait()

		close(sem)

		if err := atomicErr.Load(); err != nil {
			ctr.Logger.Error(ctr.Ctx, "failed to find error",
				logger.Value("error", err.(error)), logger.Value("on", "PipelineBatch"))
			return err.(error)
		}

		return nil
	}
}
