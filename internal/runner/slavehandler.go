package runner

import (
	"context"
	"encoding/json"
	"fmt"

	rpc "buf.build/gen/go/cresplanex/bloader/grpc/go/cresplanex/bloader/v1/bloaderv1grpc"
	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/logger"
)

// SlaveRequestHandler is a struct that holds the response handler information.
type SlaveRequestHandler struct {
	// resChan is a channel.
	resChan <-chan *pb.ReceiveChanelConnectResponse
	// cli is a client.
	cli rpc.BloaderSlaveServiceClient
	// chunkSize is an integer.
	chunkSize int
	// receiveTermChan is a channel for receiving term.
	receiveTermChan <-chan ReceiveTermType
}

const defaultChunkSize = 1024

// NewSlaveRequestHandler creates a new ResponseHandler.
func NewSlaveRequestHandler(
	resChan <-chan *pb.ReceiveChanelConnectResponse,
	cli rpc.BloaderSlaveServiceClient,
	termChan <-chan ReceiveTermType,
) *SlaveRequestHandler {
	return &SlaveRequestHandler{
		resChan:         resChan,
		cli:             cli,
		chunkSize:       defaultChunkSize,
		receiveTermChan: termChan,
	}
}

// HandleResponse handles the response.
func (rh *SlaveRequestHandler) HandleResponse(
	ctx context.Context,
	log logger.Logger,
	tmplFactor TmplFactor,
	authFactor AuthenticatorFactor,
	targetFactor TargetFactor,
	store Store,
) error {

	for {
		select {
		case termType := <-rh.receiveTermChan:
			switch termType {
			case ReceiveTermTypeReceiveTermTypeEOF:
				return nil
			case ReceiveTermTypeReceiveTermTypeResponseReceiveError:
				return fmt.Errorf("response receive error")
			case ReceiveTermTypeReceiveTermTypeContextDone:
				return nil
			case ReceiveTermTypeReceiveTermTypeStreamContextDone:
				return nil
			case ReceiveTermTypeReceiveTermTypeDisconnected:
				return nil
			default:
				return fmt.Errorf("unknown term type: %v", termType)
			}
		case res := <-rh.resChan:
			log.Debug(ctx, "Received response: %v",
				logger.Value("response", res))
			if res == nil {
				return fmt.Errorf("response is nil")
			}
			switch res.RequestType {
			case pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_LOADER:
				loaderResourceReq := res.GetLoaderResourceRequest()
				stream, err := rh.cli.SendLoader(ctx)
				if err != nil {
					return fmt.Errorf("failed to send loader: %v", err)
				}
				tmplStr, err := tmplFactor.TmplFactorize(ctx, loaderResourceReq.LoaderId)
				if err != nil {
					return fmt.Errorf("failed to factorize template: %v", err)
				}
				buffer := []byte(tmplStr)
				for i := 0; i < len(buffer); i += rh.chunkSize {
					end := i + rh.chunkSize
					if end > len(buffer) {
						end = len(buffer)
					}
					if err := stream.Send(&pb.SendLoaderRequest{
						RequestId:   res.RequestId,
						LoaderId:    loaderResourceReq.LoaderId,
						Content:     buffer[i:end],
						IsLastChunk: end == len(buffer),
					}); err != nil {
						return fmt.Errorf("failed to send loader request: %v", err)
					}
				}
				_, err = stream.CloseAndRecv()
				if err != nil {
					return fmt.Errorf("failed to receive loader response: %v", err)
				}
				log.Info(ctx, "Sent loader: %v",
					logger.Value("loader_id", loaderResourceReq.LoaderId))
			case pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_AUTH:
				authResourceReq := res.GetAuthResourceRequest()
				auth, err := authFactor.Factorize(ctx, authResourceReq.AuthId, authResourceReq.IsDefault)
				if err != nil {
					return fmt.Errorf("failed to factorize auth: %v", err)
				}
				_, err = rh.cli.SendAuth(ctx, &pb.SendAuthRequest{
					RequestId: res.RequestId,
					AuthId:    authResourceReq.AuthId,
					Auth:      auth.GetAuthValue(),
					IsDefault: authFactor.IsDefault(authResourceReq.AuthId),
				})
				if err != nil {
					return fmt.Errorf("failed to send auth: %v", err)
				}
				log.Info(ctx, "Sent auth: %v",
					logger.Value("auth_id", authResourceReq.AuthId))
			case pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_STORE:
				storeResourceReq := res.GetStoreResourceRequest()
				validStore := make([]ValidStoreImportData, len(storeResourceReq.StoreImportRequest))
				for i, storeReq := range storeResourceReq.StoreImportRequest {
					validStore[i] = ValidStoreImportData{
						BucketID: storeReq.BucketId,
						StoreKey: storeReq.StoreKey,
						Encrypt: ValidCredentialEncryptConfig{
							Enabled:   storeReq.Encryption.Enabled,
							EncryptID: storeReq.Encryption.EncryptId,
						},
					}
				}
				strData := make([]*pb.StoreExportData, 0, len(storeResourceReq.StoreImportRequest))
				store.Import(
					ctx,
					validStore,
					func(ctx context.Context, data ValidStoreImportData, val any, valBytes []byte) error {
						var err error
						if valBytes == nil {
							valBytes, err = json.Marshal(val)
							if err != nil {
								return fmt.Errorf("failed to marshal store data: %v", err)
							}
						}
						strData = append(strData, &pb.StoreExportData{
							BucketId: data.BucketID,
							StoreKey: data.StoreKey,
							Data:     valBytes,
						})
						return nil
					},
				)
				_, err := rh.cli.SendStoreData(ctx, &pb.SendStoreDataRequest{
					RequestId: res.RequestId,
					StoreData: strData,
				})
				if err != nil {
					return fmt.Errorf("failed to send store data: %v", err)
				}
				log.Info(ctx, "Sent store: %v",
					logger.Value("store_data", strData))
			case pb.RequestType_REQUEST_TYPE_STORE:
				storeReq := res.GetStore()
				storeData := make([]ValidStoreValueData, len(storeReq.StoreData))
				for i, data := range storeReq.StoreData {
					storeData[i] = ValidStoreValueData{
						BucketID: data.BucketId,
						Key:      data.StoreKey,
						Value:    data.Data,
						Encrypt: ValidCredentialEncryptConfig{
							Enabled:   data.Encryption.Enabled,
							EncryptID: data.Encryption.EncryptId,
						},
					}
				}
				if err := store.Store(ctx, storeData, nil); err != nil {
					return fmt.Errorf("failed to store data: %v", err)
				}
				if _, err := rh.cli.SendStoreOk(ctx, &pb.SendStoreOkRequest{
					RequestId: res.RequestId,
				}); err != nil {
					return fmt.Errorf("failed to send store ok: %v", err)
				}
				log.Info(ctx, "Stored data: %v",
					logger.Value("store_data", storeData))
			case pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_TARGET:
				targetResourceReq := res.GetTargetResourceRequest()
				target, err := targetFactor.Factorize(ctx, targetResourceReq.TargetId)
				if err != nil {
					return fmt.Errorf("failed to factorize target: %v", err)
				}
				_, err = rh.cli.SendTarget(ctx, &pb.SendTargetRequest{
					RequestId: res.RequestId,
					TargetId:  targetResourceReq.TargetId,
					Target:    target.GetTarget(),
				})
				if err != nil {
					return fmt.Errorf("failed to send target: %v", err)
				}
				log.Info(ctx, "Sent target: %v",
					logger.Value("target_id", targetResourceReq.TargetId))
			default:
				return fmt.Errorf("unknown request type: %v", res.RequestType)
			}
		}
	}
}
