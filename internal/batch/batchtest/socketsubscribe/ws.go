package socketsubscribe

import (
	"context"
	"fmt"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"github.com/LabGroupware/go-measure-tui/internal/ws"
)

type Socket struct {
	ws *ws.WebSocket
}

func NewSocket(ctx context.Context, ctr *app.Container, msgHandler ws.EventResponseMessageHandleFunc) (*Socket, error) {
	sock := ws.NewWebSocket()

	sock.CannotParseMsgHandler = func(ws *ws.WebSocket, msg *ws.CannotParseResponseMessage) error {
		ctr.Logger.Error(ctx, "cannot parse response",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.SubscribeMsgHandler = func(ws *ws.WebSocket, msg *ws.SubscribeResponseMessage) error {
		ctr.Logger.Debug(ctx, "subscribed",
			logger.Value("subscription_id", msg.SubscriptionID))

		return nil
	}
	sock.UnsubscribeMsgHandler = func(ws *ws.WebSocket, msg *ws.UnsubscribeResponseMessage) error {
		ctr.Logger.Debug(ctx, "unsubscribed",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.UnsupportedMsgHandler = func(ws *ws.WebSocket, msg *ws.UnsupportedResponseMessage) error {
		ctr.Logger.Error(ctx, "unsupported response",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.EventMsgHandler = msgHandler

	return &Socket{
		ws: sock,
	}, nil
}

func NewSocketWithSubscribeHandler(
	ctx context.Context,
	ctr *app.Container,
	msgHandler ws.EventResponseMessageHandleFunc,
	subscribeHandler ws.SubscribeResponseMessageHandleFunc,
) (*Socket, error) {
	sock := ws.NewWebSocket()

	sock.CannotParseMsgHandler = func(ws *ws.WebSocket, msg *ws.CannotParseResponseMessage) error {
		ctr.Logger.Error(ctx, "cannot parse response",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.SubscribeMsgHandler = subscribeHandler
	sock.UnsubscribeMsgHandler = func(ws *ws.WebSocket, msg *ws.UnsubscribeResponseMessage) error {
		ctr.Logger.Debug(ctx, "unsubscribed",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.UnsupportedMsgHandler = func(ws *ws.WebSocket, msg *ws.UnsupportedResponseMessage) error {
		ctr.Logger.Error(ctx, "unsupported response",
			logger.Value("message_id", msg.MessageID))

		return nil
	}
	sock.EventMsgHandler = msgHandler

	return &Socket{
		ws: sock,
	}, nil
}

func (s *Socket) Subscribe(
	ctx context.Context,
	consumerID string,
	aggregateType ws.AggregateType,
	aggregateIDs []string,
	eventTypes []ws.EventType,
) error {
	if err := s.ws.SendSubscribeMessage(consumerID, aggregateType, aggregateIDs, eventTypes); err != nil {
		return fmt.Errorf("failed to send subscribe message: %v", err)
	}
	return nil
}

func (s *Socket) Close() {
	s.ws.AllUnsubscribe()
	s.ws.Close()
}

func (s *Socket) Connect(ctx context.Context, ctr *app.Container, config ws.ConnectConfig) (<-chan ws.TerminateType, error) {
	done, err := s.ws.Connect(ctx, ctr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to websocket: %v", err)
	}

	return done, nil
}
