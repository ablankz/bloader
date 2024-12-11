package oneexecbatch

import (
	"context"
	"fmt"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
)

func OneExecuteBatch(
	ctx context.Context,
	ctr *app.Container,
	conf OneExecuteConfig,
	store *sync.Map,
) (map[string]string, error) {
	newStore := sync.Map{}
	err := executeRequest(ctx, ctr, 0, conf.Request, &newStore)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to find error",
			logger.Value("error", err))
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}

	newMap := make(map[string]string)

	newStore.Range(func(key, value interface{}) bool {
		store.Store(key, value)
		newMap[key.(string)] = value.(string)
		return true
	})

	return newMap, nil
}
