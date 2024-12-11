package prefetchbatch

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"github.com/LabGroupware/go-measure-tui/internal/utils"
)

type prefetchExecuteRequest struct {
	id              string
	request         *PrefetchRequest
	mustWaitChan    *[]<-chan struct{}
	selfBroadCaster *utils.Broadcaster[struct{}]
}

func PrefetchBatch(ctx context.Context, ctr *app.Container, conf PrefetchConfig, store *sync.Map) (map[string]string, error) {
	newStore := sync.Map{}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	requests := make([]prefetchExecuteRequest, len(conf.Requests))
	for i, req := range conf.Requests {
		requests[i] = prefetchExecuteRequest{
			id:              req.ID,
			request:         req,
			mustWaitChan:    &[]<-chan struct{}{},
			selfBroadCaster: utils.NewBroadcaster[struct{}](),
		}
		defer requests[i].selfBroadCaster.Close()
	}

	for _, req := range requests {
		if len(req.request.DependsOn) == 0 {
			continue
		}
		for _, depID := range req.request.DependsOn {
			var exists bool
			for _, depReq := range requests {
				if depID == depReq.id {
					*req.mustWaitChan = append(*req.mustWaitChan, depReq.selfBroadCaster.Subscribe())
					exists = true
				}
			}
			if !exists {
				return nil, fmt.Errorf("request %s depends on non-existent request(%s)", req.id, depID)
			}
		}
		ctr.Logger.Debug(ctx, "request depends on",
			logger.Value("id", req.id), logger.Value("depends_on", req.request.DependsOn), logger.Value("on", "PrefetchBatch"))
	}

	atomicErr := atomic.Value{}
	wg := sync.WaitGroup{}
	for i, req := range requests {
		wg.Add(1)
		go func(preReq prefetchExecuteRequest) {
			defer func() {
				broadDone := preReq.selfBroadCaster.Broadcast(struct{}{})
				<-broadDone
				wg.Done()
			}()

			unprocessed := len(*preReq.mustWaitChan)
			for _, waitChan := range *preReq.mustWaitChan {
			loop:
				select {
				case <-ctx.Done():
					<-waitChan
					return
				case <-waitChan:
					unprocessed--
					if unprocessed == 0 {
						break loop
					}
				}
			}

			err := executeRequest(ctx, ctr, i, preReq.request, &newStore, len(*preReq.mustWaitChan) > 0)

			if err != nil {
				atomicErr.Store(err)
				ctr.Logger.Error(ctx, "failed to execute request",
					logger.Value("request_id", preReq.id), logger.Value("error", err), logger.Value("on", "PrefetchBatch"))
				ctr.Logger.Info(ctx, "terminating all requests: failed to execute prefetch request",
					logger.Value("request_id", preReq.id), logger.Value("on", "PrefetchBatch"))
				cancel()
				return
			}
			ctr.Logger.Debug(ctx, "request finished",
				logger.Value("id", preReq.id), logger.Value("on", "PrefetchBatch"))
		}(req)
	}

	ctr.Logger.Debug(ctx, "waiting for all requests to finish",
		logger.Value("on", "PrefetchBatch"))
	wg.Wait()
	ctr.Logger.Debug(ctx, "all requests finished",
		logger.Value("on", "PrefetchBatch"))

	if err := atomicErr.Load(); err != nil {
		ctr.Logger.Error(ctx, "failed to find error",
			logger.Value("error", err.(error)), logger.Value("on", "PrefetchBatch"))
		return nil, err.(error)
	}

	result := make(map[string]string)
	newStore.Range(func(key, value interface{}) bool {
		store.Store(key.(string), value.(string))
		result[key.(string)] = value.(string)
		return true
	})

	return result, nil
}
