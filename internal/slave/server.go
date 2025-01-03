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
	common "github.com/ablankz/go-cmproto"
	"google.golang.org/grpc"
)

// commandTermData represents the command term data
type commandTermData struct {
	Success bool
}

// Server represents the server for the worker node
type Server struct {
	mu          *sync.RWMutex
	encryptCtr  encrypt.EncrypterContainer
	env         string
	log         logger.Logger
	slaveConCtr *runner.ConnectionContainer
	slCtrMap    map[string]*slcontainer.SlaveContainer
	reqConMap   *slcontainer.RequestConnectionMapper
	cmdTermMap  map[string]chan commandTermData
}

// NewServer creates a new server for the worker node
func NewServer(ctr *container.Container, slaveConCtr *runner.ConnectionContainer) *Server {
	return &Server{
		mu:          &sync.RWMutex{},
		encryptCtr:  ctr.EncypterContainer,
		env:         ctr.Config.Env,
		log:         ctr.Logger,
		slaveConCtr: slaveConCtr,
		slCtrMap:    make(map[string]*slcontainer.SlaveContainer),
		reqConMap:   slcontainer.NewRequestConnectionMapper(),
		cmdTermMap:  make(map[string]chan commandTermData),
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
	slaveValues, err := common.FromFlexMap(req.SlaveValues)
	if err != nil {
		return nil, fmt.Errorf("failed to convert slave values: %v", err)
	}

	cmdMapData := slcontainer.CommandMapData{
		LoaderID:         req.LoaderId,
		OutputRoot:       req.OutputRoot,
		StrMap:           &strMap,
		ThreadOnlyStrMap: &threadOnlyStrMap,
		SlaveValues:      slaveValues,
	}
	slCtr.AddCommandMap(uid, cmdMapData)
	s.cmdTermMap[uid] = make(chan commandTermData)

	return &pb.SlaveCommandResponse{
		CommandId: uid,
	}, nil
}

// CallExec handles the exec request from the master node
func (s *Server) CallExec(req *pb.CallExecRequest, stream grpc.ServerStreamingServer[pb.CallExecResponse]) error {
	s.mu.Lock()
	slCtr, ok := s.slCtrMap[req.ConnectionId]
	if !ok {
		return ErrInvalidConnectionID
	}
	fmt.Println("req.CommandId", req.CommandId, "CommandMap", slCtr.CommandMap)
	data, ok := slCtr.GetCommandMap(req.CommandId)
	if !ok {
		return ErrCommandNotFound
	}
	s.mu.Unlock()
	var err error
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if err != nil {
			s.cmdTermMap[req.CommandId] <- commandTermData{
				Success: false,
			}
			return
		}
		s.cmdTermMap[req.CommandId] <- commandTermData{
			Success: true,
		}
	}()
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
				s.log.Warn(st.Context(), "stream context done",
					logger.Value("ConnectionID", req.ConnectionId), logger.Value("Error", stream.Context().Err()))
				return
			case res := <-outputChan:
				fmt.Println("out res", res)
				if err := st.Send(res); err != nil {
					s.log.Error(stream.Context(), "failed to send a response",
						logger.Value("Error", err))
				}
			}
		}
	}(stream)

	exec := runner.BaseExecutor{
		Logger:                s.log,
		Env:                   s.env,
		SlaveConnectContainer: s.slaveConCtr,
		EncryptCtr:            s.encryptCtr,
		TmplFactor:            tmplFactor,
		TargetFactor:          targetFactor,
		AuthFactor:            authFactor,
		Store:                 store,
		OutputFactor:          outputFactor,
	}
	if err = exec.Execute(
		stream.Context(),
		data.LoaderID,
		data.StrMap,
		data.ThreadOnlyStrMap,
		data.OutputRoot,
		0,
		0,
		data.SlaveValues,
		runner.NewDefaultEventCaster(),
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
			s.log.Warn(stream.Context(), "stream context done",
				logger.Value("ConnectionID", req.ConnectionId), logger.Value("Error", stream.Context().Err()))
			return fmt.Errorf("context done: %v", stream.Context().Err())
		}
	}
}

// SendLoader handles the loader request from the master node
func (s *Server) SendLoader(stream grpc.ClientStreamingServer[pb.SendLoaderRequest, pb.SendLoaderResponse]) error {
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return fmt.Errorf("failed to receive a chunk: %v", err)
		}
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
		if chunk.IsLastChunk {
			// Stream is done
			slCtr.Loader.WriteString(chunk.LoaderId, string(chunk.Content))
			slCtr.Loader.Build(chunk.LoaderId)
			slCtr.ReceiveChanelRequestContainer.Cast(chunk.RequestId)
			s.reqConMap.DeleteRequest(chunk.RequestId)
			return stream.SendAndClose(&pb.SendLoaderResponse{})
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

	for _, data := range req.StoreData {
		slCtr.Store.AddData(data.BucketId, data.StoreKey, data.Data)
	}
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

// ReceiveLoadTermChannel handles the load term channel request from the master node
func (s *Server) ReceiveLoadTermChannel(ctx context.Context, req *pb.ReceiveLoadTermChannelRequest) (*pb.ReceiveLoadTermChannelResponse, error) {
	s.mu.RLock()
	cmdTermChan, ok := s.cmdTermMap[req.CommandId]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrCommandNotFound
	}
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		close(s.cmdTermMap[req.CommandId])
		delete(s.cmdTermMap, req.CommandId)
	}()

	select {
	case data := <-cmdTermChan:
		return &pb.ReceiveLoadTermChannelResponse{
			Success: data.Success,
		}, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("context done")
	}
}
