package socketsubscribe

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/LabGroupware/go-measure-tui/internal/ws"
	"github.com/jmespath/go-jmespath"
)

type sock struct {
	mu                         *sync.Mutex
	s                          *Socket
	actionsFileMap             *sync.Map
	actionIndexToConsumerID    *sync.Map
	actionsToConsumerIDMap     *sync.Map
	consumerSelfEventFilterMap *sync.Map
	consumerTermChanMap        *sync.Map
	consumerStarttimeMap       *sync.Map
}

type SelfEventFilter struct {
	JMESPath *jmespath.JMESPath
}

type globalSock struct {
	mu      *sync.Mutex
	sockMap map[string]*sock
}

var GlobalSock = &globalSock{
	mu:      &sync.Mutex{},
	sockMap: make(map[string]*sock),
}

func NewSock(s *Socket, outputEnabled bool) *sock {
	return &sock{
		mu:                         &sync.Mutex{},
		s:                          s,
		actionsFileMap:             &sync.Map{},
		actionIndexToConsumerID:    &sync.Map{},
		actionsToConsumerIDMap:     &sync.Map{},
		consumerStarttimeMap:       &sync.Map{},
		consumerSelfEventFilterMap: &sync.Map{},
		consumerTermChanMap:        &sync.Map{},
	}
}

func (s *sock) Subscribe(
	ctx context.Context,
	consumerId string,
	aggregateType ws.AggregateType,
	aggregateIDs []string,
	eventTypes []ws.EventType,
	actions []SocketSubscribeActionConfig,
	selfEventFilter *SelfEventFilter,
) (<-chan string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.s.Subscribe(ctx, consumerId, aggregateType, aggregateIDs, eventTypes); err != nil {
		return nil, fmt.Errorf("failed to send subscribe message: %v", err)
	}
	notifyChan := make(chan string)

	for _, action := range actions {
		s.actionIndexToConsumerID.Store(action.ID, consumerId)
	}
	s.actionsToConsumerIDMap.Store(consumerId, actions)
	s.consumerTermChanMap.Store(consumerId, notifyChan)
	s.consumerSelfEventFilterMap.Store(consumerId, selfEventFilter)
	s.consumerStarttimeMap.Store(consumerId, time.Now())

	return notifyChan, nil
}

func (s *sock) GetStartTimeByConsumerID(consumerID string) (time.Time, bool) {

	s.mu.Lock()
	defer s.mu.Unlock()

	if startTime, ok := s.consumerStarttimeMap.Load(consumerID); ok {
		return startTime.(time.Time), true
	}

	return time.Time{}, false
}

func (s *sock) GetOwnerFromData(data any) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var selfConsumer string

	s.consumerSelfEventFilterMap.Range(func(k, v any) bool {
		if v.(*SelfEventFilter).JMESPath != nil {
			if result, err := v.(*SelfEventFilter).JMESPath.Search(data); err == nil && result != nil {
				res, ok := result.(bool)
				if ok && res {
					selfConsumer = k.(string)
					return false
				}
			}
		}
		return true
	})

	return selfConsumer
}

func (s *sock) UnsubscribeNotifyByAction(
	ctx context.Context,
	actionId string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var consumerId string
	var ok bool
	if consumerId, ok = s.getConsumerIDByActionID(actionId); ok {
		if v, ok := s.consumerTermChanMap.Load(consumerId); ok {
			if v != nil {
				if v, ok := v.(chan string); ok {
					select {
					case <-ctx.Done():
						return fmt.Errorf("context cancelled")
					case v <- actionId: // Notify
					}
				}
			}
			return nil
		}
	}
	return nil
}

func (s *sock) getConsumerIDByActionID(actionID string) (string, bool) {

	if consumerID, ok := s.actionIndexToConsumerID.Load(actionID); ok {
		return consumerID.(string), true
	}

	return "", false
}

func (s *sock) Unsubscribe(
	ctx context.Context,
	consumerId string,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.s.ws.UnsubscribeByConsumerID(consumerId); err != nil {
		return fmt.Errorf("failed to send unsubscribe message: %v", err)
	}

	s.removeActions(s.getActionsByConsumerID(consumerId))

	// s.DisplayDebug()

	return nil
}

func (s *sock) getActionsByConsumerID(consumerID string) []SocketSubscribeActionConfig {

	if actions, ok := s.actionsToConsumerIDMap.Load(consumerID); ok {
		return actions.([]SocketSubscribeActionConfig)
	}

	return []SocketSubscribeActionConfig{}
}

func (s *sock) removeActions(actions []SocketSubscribeActionConfig) {
	actionIDs := []string{}
	for _, action := range actions {
		actionIDs = append(actionIDs, action.ID)
		s.actionIndexToConsumerID.Delete(action.ID)
	}

	s.removePluralActionsFileMap(actionIDs)
}

func (s *sock) AddActionsFileMap(actionId string, file *os.File) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.actionsFileMap.Store(actionId, file)
}

func (s *sock) GetActionsFileMap(actionId string) (*os.File, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if file, ok := s.actionsFileMap.Load(actionId); ok {
		return file.(*os.File), true
	}

	return nil, false
}

func (s *sock) removePluralActionsFileMap(actionIds []string) {
	for _, actionId := range actionIds {
		s.actionsFileMap.Delete(actionId)
	}
}

func (s *sock) removeAllActionsFileMap() {
	s.actionsFileMap = &sync.Map{}
}

func (s *sock) removeAllActions() {
	s.actionIndexToConsumerID = &sync.Map{}
}

func (s *globalSock) FindSocket(id string) (*sock, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sock, ok := s.sockMap[id]; ok {
		return sock, nil
	}

	return nil, fmt.Errorf("socket not found: %s", id)
}

func (s *globalSock) AddSocket(id string, sock *sock) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sockMap[id] = sock
}

func (s *globalSock) CloseSocket(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sock, ok := s.sockMap[id]; ok {
		sock.removeAllActions()
		sock.removeAllActionsFileMap()
		sock.s.Close()
		delete(s.sockMap, id)
	}
}
