package socketsubscribe

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"sync"

	"github.com/LabGroupware/go-measure-tui/internal/app"
	"github.com/LabGroupware/go-measure-tui/internal/logger"
	"github.com/LabGroupware/go-measure-tui/internal/ws"
	"github.com/google/uuid"
	"github.com/jmespath/go-jmespath"
)

func SocketSubscribe(
	ctx context.Context,
	ctr *app.Container,
	conf SocketSubscribeConfig,
	store *sync.Map,
	selfLoopCount string,
	outputRoot string,
) error {
	defer fmt.Println("SocketSubscribe done", "----", conf.ID)

	s, err := GlobalSock.FindSocket(conf.ConnectID)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to find socket",
			logger.Value("error", err))
		return fmt.Errorf("failed to find socket: %v", err)
	}

	if conf.Output.Enabled {
		for _, action := range conf.Actions {
			if ContainsSocketActionType(action.Types, SocketActionTypeOutput) {
				logFilePath := fmt.Sprintf("%s/socket_subscribe_%s.csv", outputRoot, action.ID)
				file, err := os.Create(logFilePath)
				if err != nil {
					return fmt.Errorf("failed to create file: %v", err)
				}
				defer file.Close()
				s.AddActionsFileMap(action.ID, file)

				header := []string{"EventType", "StartTime", "ReceivedDatetime", "TotalTime"}
				for _, data := range action.Data {
					header = append(header, data.Key)
				}
				writer := csv.NewWriter(file)
				if err := writer.Write(header); err != nil {
					return fmt.Errorf("failed to write header: %v", err)
				}
				writer.Flush()
			}
		}
	}

	consumerID := uuid.New().String()

	var path *jmespath.JMESPath

	if conf.SelfEventFilter.JMESPath != "" {
		if path, err = jmespath.Compile(conf.SelfEventFilter.JMESPath); err != nil {
			ctr.Logger.Error(ctx, "failed to compile jmespath",
				logger.Value("error", err))
			return fmt.Errorf("failed to compile jmespath: %v", err)
		}
	}

	term, err := s.Subscribe(
		ctx,
		consumerID,
		ws.AggregateType(conf.Subscribe.AggregateType),
		conf.Subscribe.AggregateId,
		ws.NewFromEventTypesString(conf.Subscribe.EventTypes),
		conf.Actions,
		&SelfEventFilter{
			JMESPath: path,
		},
	)
	if err != nil {
		ctr.Logger.Error(ctx, "failed to subscribe",
			logger.Value("error", err))
		return fmt.Errorf("failed to subscribe: %v", err)
	}
	// defer s.Unsubscribe(ctx, consumerID)

	select {
	case <-ctx.Done():
		ctr.Logger.Warn(ctx, "context cancelled")
		return fmt.Errorf("context cancelled")
	case actionID := <-term:
		fmt.Println("--------------actionID: ", actionID)
		for _, action := range conf.SuccessUnsubscribeActionIDs {
			if action == actionID {
				ctr.Logger.Debug(ctx, "success unsubscribe",
					logger.Value("action_id", actionID))
				return nil
			}
		}
		ctr.Logger.Error(ctx, "failed to unsubscribe",
			logger.Value("action_id", actionID))
		return fmt.Errorf("failed to unsubscribe: %s", actionID)
	}
}
