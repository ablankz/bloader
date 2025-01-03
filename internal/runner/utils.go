package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/ablankz/bloader/internal/logger"
)

func wait(
	ctx context.Context,
	log logger.Logger,
	conf ValidRunner,
	after RunnerSleepValueAfter,
) error {
	if v, wait := conf.RetrieveSleepValue(after); wait {
		log.Debug(ctx, "sleeping after execute",
			logger.Value("duration", v))
		fmt.Println("sleeping for", v, "...")
		select {
		case <-time.After(v):
		case <-ctx.Done():
			return fmt.Errorf("context canceled")
		}
		fmt.Println("sleeping complete")
	}

	return nil
}

func validate(
	ctx context.Context,
	eventCaster EventCaster,
	validateFunc func() error,
) error {
	if err := eventCaster.CastEvent(ctx, RunnerEventValidating); err != nil {
		return fmt.Errorf("failed to cast event: %v", err)
	}
	if err := validateFunc(); err != nil {
		return err
	}
	if err := eventCaster.CastEvent(ctx, RunnerEventValidated); err != nil {
		return fmt.Errorf("failed to cast event: %v", err)
	}

	return nil
}
