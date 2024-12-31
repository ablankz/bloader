package slave

import (
	"context"
	"fmt"
	"io"
	"sync"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/encrypt"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/runner"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
	"github.com/ablankz/bloader/internal/utils"
	common "github.com/ablankz/common-proto/lib/go"
	"google.golang.org/grpc"
)

// Server represents the server for the worker node
type Server struct {
	mu         *sync.RWMutex
	encryptCtr encrypt.EncrypterContainer
	env        string
	log        logger.Logger
	slCtrMap   map[string]*slcontainer.SlaveContainer
	reqConMap  *slcontainer.RequestConnectionMapper
}

// NewServer creates a new server for the worker node
func NewServer(ctr *container.Container) *Server {
	return &Server{
		mu:         &sync.RWMutex{},
		encryptCtr: ctr.EncypterContainer,
		env:        ctr.Config.Env,
		log:        ctr.Logger,
		slCtrMap:   make(map[string]*slcontainer.SlaveContainer),
		reqConMap:  slcontainer.NewRequestConnectionMapper(),
	}
}

// Connect handles the connection request from the master node
func (s *Server) Connect(ctx context.Context, req *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	response := &pb.ConnectResponse{}
	if req.Environment != s.env {
		return nil, ErrInvalidEnvironment
	}
	uid := utils.GenerateUniqueID()
	s.slCtrMap[uid] = slcontainer.NewSlaveContainer()
	response.ConnectionId = uid
	return response, nil
}

// Disconnect handles the disconnection request from the master node
func (s *Server) Disconnect(ctx context.Context, req *pb.DisconnectRequest) (*pb.DisconnectResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.slCtrMap, req.ConnectionId)
	s.reqConMap.DeleteRequestConnection(req.ConnectionId)
	return &pb.DisconnectResponse{}, nil
}

// SlaveCommand handles the command request from the master node
func (s *Server) SlaveCommand(ctx context.Context, req *pb.SlaveCommandRequest) (*pb.SlaveCommandResponse, error) {
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[req.ConnectionId]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrInvalidConnectionID
	}

	uid := utils.GenerateUniqueID()
	term := slCtr.ReceiveChanelRequestContainer.SendLoaderResourceRequests(
		ctx,
		req.ConnectionId,
		s.reqConMap,
		slcontainer.LoaderResourceRequest{
			LoaderID: req.LoaderId,
		},
	)
	if term == nil {
		return nil, ErrFailedToSendLoaderResourceRequest
	}
	select {
	case <-ctx.Done():
		return nil, ErrFailedToSendLoaderResourceRequest
	case <-term:
	}
	s.log.Info(ctx, "Initial Loader Received",
		logger.Value("ConnectionID", req.ConnectionId), logger.Value("LoaderID", req.LoaderId))
	if !ok {
		return nil, ErrLoaderNotFound
	}
	strMap := sync.Map{}
	threadOnlyStrMap := sync.Map{}

	fMap, err := common.FromFlexMap(req.DefaultStore)
	if err != nil {
		return nil, fmt.Errorf("failed to convert default store: %v", err)
	}
	for k, v := range fMap {
		strMap.Store(k, v)
	}
	fMap, err = common.FromFlexMap(req.DefaultThreadOnlyStore)
	if err != nil {
		return nil, fmt.Errorf("failed to convert default thread only store: %v", err)
	}
	for k, v := range fMap {
		threadOnlyStrMap.Store(k, v)
	}

	cmdMapData := slcontainer.CommandMapData{
		LoaderID:         req.LoaderId,
		OutputRoot:       req.OutputRoot,
		StrMap:           &strMap,
		ThreadOnlyStrMap: &threadOnlyStrMap,
	}
	slCtr.AddCommandMap(uid, cmdMapData)

	return &pb.SlaveCommandResponse{
		CommandId: uid,
	}, nil
}

// CallExec handles the exec request from the master node
func (s *Server) CallExec(req *pb.CallExecRequest, stream grpc.ServerStreamingServer[pb.CallExecResponse]) error {
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[req.ConnectionId]
	s.mu.RUnlock()
	if !ok {
		return ErrInvalidConnectionID
	}
	data, ok := slCtr.GetCommandMap(req.CommandId)
	if !ok {
		return ErrCommandNotFound
	}
	tmplFactor := &SlaveTmplFactor{
		loader:                        slCtr.Loader,
		connectionID:                  req.ConnectionId,
		receiveChanelRequestContainer: slCtr.ReceiveChanelRequestContainer,
		mapper:                        s.reqConMap,
	}
	targetFactor := &SlaveTargetFactor{
		target:                        slCtr.Target,
		connectionID:                  req.ConnectionId,
		receiveChanelRequestContainer: slCtr.ReceiveChanelRequestContainer,
		mapper:                        s.reqConMap,
	}
	authFactor := &SlaveAuthenticatorFactor{
		auth:                          slCtr.Auth,
		connectionID:                  req.ConnectionId,
		receiveChanelRequestContainer: slCtr.ReceiveChanelRequestContainer,
		mapper:                        s.reqConMap,
	}
	store := &SlaveStore{
		store:                         slCtr.Store,
		connectionID:                  req.ConnectionId,
		receiveChanelRequestContainer: slCtr.ReceiveChanelRequestContainer,
		mapper:                        s.reqConMap,
	}

	outputChan := make(chan *pb.CallExecResponse)
	outputFactor := &SlaveOutputFactor{
		outputChan: outputChan,
	}

	go func(st grpc.ServerStreamingServer[pb.CallExecResponse]) {
		for {
			select {
			case <-stream.Context().Done():
				return
			case res := <-outputChan:
				if err := st.Send(res); err != nil {
					s.log.Error(stream.Context(), "failed to send a response",
						logger.Value("Error", err))
					return
				}
			}
		}
	}(stream)

	exec := runner.BaseExecutor{
		Logger:       s.log,
		EncryptCtr:   s.encryptCtr,
		TmplFactor:   tmplFactor,
		TargetFactor: targetFactor,
		AuthFactor:   authFactor,
		Store:        store,
		OutputFactor: outputFactor,
	}
	if err := exec.Execute(
		stream.Context(),
		data.LoaderID,
		data.StrMap,
		data.ThreadOnlyStrMap,
		data.OutputRoot,
		0,
		0,
	); err != nil {
		return fmt.Errorf("failed to execute: %v", err)
	}

	return nil
}

// ReceiveChanelConnect handles the channel connection request from the master node
func (s *Server) ReceiveChanelConnect(req *pb.ReceiveChanelConnectRequest, stream grpc.ServerStreamingServer[pb.ReceiveChanelConnectResponse]) error {
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[req.ConnectionId]
	s.mu.RUnlock()
	if !ok {
		return ErrInvalidConnectionID
	}

	for {
		select {
		case res := <-slCtr.ReceiveChanelRequestContainer.ReqChan:
			if err := stream.Send(res); err != nil {
				return fmt.Errorf("failed to send a response: %v", err)
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

// SendLoader handles the loader request from the master node
func (s *Server) SendLoader(stream grpc.ClientStreamingServer[pb.SendLoaderRequest, pb.SendLoaderResponse]) error {
	for {
		chunk, err := stream.Recv()
		conId, ok := s.reqConMap.GetConnectionID(chunk.RequestId)
		if !ok {
			return ErrRequestNotFound
		}
		s.mu.RLock()
		slCtr, ok := s.slCtrMap[conId]
		s.mu.RUnlock()
		if !ok {
			return ErrRequestNotFound
		}
		if err == io.EOF {
			// Stream is done
			slCtr.Loader.Build(chunk.LoaderId)
			slCtr.ReceiveChanelRequestContainer.Cast(chunk.RequestId)
			s.reqConMap.DeleteRequest(chunk.RequestId)
			return stream.SendAndClose(&pb.SendLoaderResponse{})
		}
		if err != nil {
			s.mu.Unlock()
			return fmt.Errorf("failed to receive a chunk: %v", err)
		}
		slCtr.Loader.WriteString(chunk.LoaderId, string(chunk.Content))
	}
}

// SendAuth handles the auth request from the master node
func (s *Server) SendAuth(ctx context.Context, req *pb.SendAuthRequest) (*pb.SendAuthResponse, error) {
	conID, ok := s.reqConMap.GetConnectionID(req.RequestId)
	if !ok {
		return nil, ErrRequestNotFound
	}
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[conID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrRequestNotFound
	}
	if err := slCtr.Auth.AddFromProto(req.AuthId, req.Auth); err != nil {
		return nil, err
	}
	if req.IsDefault {
		slCtr.Auth.DefaultAuthenticator = req.AuthId
	}
	slCtr.ReceiveChanelRequestContainer.Cast(req.RequestId)
	s.reqConMap.DeleteRequest(req.RequestId)

	return &pb.SendAuthResponse{}, nil
}

// SendStoreData handles the store data request from the master node
func (s *Server) SendStoreData(ctx context.Context, req *pb.SendStoreDataRequest) (*pb.SendStoreDataResponse, error) {
	conID, ok := s.reqConMap.GetConnectionID(req.RequestId)
	if !ok {
		return nil, ErrRequestNotFound
	}
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[conID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrRequestNotFound
	}

	slCtr.Store.AddData(req.BucketId, req.StoreKey, req.Data)
	slCtr.ReceiveChanelRequestContainer.Cast(req.RequestId)
	s.reqConMap.DeleteRequest(req.RequestId)
	return &pb.SendStoreDataResponse{}, nil
}

// SendStoreOk handles the store ok request from the master node
func (s *Server) SendStoreOk(ctx context.Context, req *pb.SendStoreOkRequest) (*pb.SendStoreOkResponse, error) {
	conID, ok := s.reqConMap.GetConnectionID(req.RequestId)
	if !ok {
		return nil, ErrRequestNotFound
	}
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[conID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrRequestNotFound
	}
	slCtr.ReceiveChanelRequestContainer.Cast(req.RequestId)
	s.reqConMap.DeleteRequest(req.RequestId)

	return &pb.SendStoreOkResponse{}, nil
}

// SendTarget handles the target request from the master node
func (s *Server) SendTarget(ctx context.Context, req *pb.SendTargetRequest) (*pb.SendTargetResponse, error) {
	conID, ok := s.reqConMap.GetConnectionID(req.RequestId)
	if !ok {
		return nil, ErrRequestNotFound
	}
	s.mu.RLock()
	slCtr, ok := s.slCtrMap[conID]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrRequestNotFound
	}
	if err := slCtr.Target.AddFromProto(req.TargetId, req.Target); err != nil {
		return nil, err
	}
	slCtr.ReceiveChanelRequestContainer.Cast(req.RequestId)
	s.reqConMap.DeleteRequest(req.RequestId)

	return &pb.SendTargetResponse{}, nil
}
