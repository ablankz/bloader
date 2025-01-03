package runner

import (
	"context"
	"fmt"

	"github.com/ablankz/bloader/internal/utils"
)

// RunnerEvent represents the flow step flow depends on event
type RunnerEvent string

const (
	// RunnerEventStart represents the event start
	RunnerEventStart RunnerEvent = "sys:start"
	// RunnerEventStoreImporting represents the event store importing
	RunnerEventStoreImporting RunnerEvent = "sys:store:importing"
	// RunnerEventStoreImported represents the event store imported
	RunnerEventStoreImported RunnerEvent = "sys:store:imported"
	// RunnerEventValidating represents the event validating
	RunnerEventValidating RunnerEvent = "sys:validating"
	// RunnerEventValidated represents the event validated
	RunnerEventValidated RunnerEvent = "sys:validated"
	// RunnerEventTerminated represents the event terminated
	RunnerEventTerminated RunnerEvent = "sys:terminated"
)

// EventCaster is an interface for casting event
type EventCaster interface {
	// CastEvent casts the event
	CastEvent(ctx context.Context, event RunnerEvent) error
	// CastEventWithWait casts the event with wait
	CastEventWithWait(ctx context.Context, event RunnerEvent) error
	// Subscribe subscribes to the event
	Subscribe(ctx context.Context) error
	// Unsubscribe unsubscribes to the event
	Unsubscribe(ctx context.Context, ch chan RunnerEvent) error
	// Close closes the event caster
	Close(ctx context.Context) error
}

// DefaultEventCaster is a struct that holds the event caster information
type DefaultEventCaster struct {
	// Caster is an event caster
	Caster *utils.Broadcaster[RunnerEvent]
}

// NewDefaultEventCaster creates a new DefaultEventCaster
func NewDefaultEventCaster() *DefaultEventCaster {
	return &DefaultEventCaster{
		Caster: utils.NewBroadcaster[RunnerEvent](),
	}
}

// NewDefaultEventCasterWithBroadcaster creates a new DefaultEventCaster with broadcaster
func NewDefaultEventCasterWithBroadcaster(broadcaster *utils.Broadcaster[RunnerEvent]) *DefaultEventCaster {
	return &DefaultEventCaster{
		Caster: broadcaster,
	}
}

// CastEvent casts the event
func (ec *DefaultEventCaster) CastEvent(ctx context.Context, event RunnerEvent) error {
	ec.Caster.Broadcast(event)
	return nil
}

// CastEventWithWait casts the event with wait
func (ec *DefaultEventCaster) CastEventWithWait(ctx context.Context, event RunnerEvent) error {
	waitChan := ec.Caster.Broadcast(event)
	select {
	case <-waitChan:
	case <-ctx.Done():
		return fmt.Errorf("failed to wait for the event: %w", ctx.Err())
	}
	return nil
}

// Subscribe subscribes to the event
func (ec *DefaultEventCaster) Subscribe(ctx context.Context) error {
	ec.Caster.Subscribe()
	return nil
}

// Unsubscribe unsubscribes to the event
func (ec *DefaultEventCaster) Unsubscribe(ctx context.Context, ch chan RunnerEvent) error {
	ec.Caster.Unsubscribe(ch)
	return nil
}

// Close closes the event caster
func (ec *DefaultEventCaster) Close(ctx context.Context) error {
	ec.Caster.Close()
	return nil
}
