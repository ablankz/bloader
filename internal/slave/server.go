package slave

import (
	"context"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
	"github.com/ablankz/bloader/internal/utils"
	"google.golang.org/grpc"
)

// Server represents the server for the worker node
type Server struct {
	ctr   *container.Container
	slCtr *slcontainer.SlaveContainer
}

// NewServer creates a new server for the worker node
func NewServer(ctr *container.Container, slCtr *slcontainer.SlaveContainer) *Server {
	return &Server{
		ctr:   ctr,
		slCtr: slCtr,
	}
}

// Connect handles the connection request from the master node
func (s *Server) Connect(ctx context.Context, req *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	response := &pb.ConnectResponse{}
	if req.Environment != s.ctr.Config.Env {
		return nil, ErrInvalidEnvironment
	}
	response.ConnectionId = utils.GenerateUniqueID()
	return response, nil
}

// SlaveCommand handles the command request from the master node
func (s *Server) SlaveCommand(ctx context.Context, req *pb.SlaveCommandRequest) (*pb.SlaveCommandResponse, error) {
	return &pb.SlaveCommandResponse{}, nil
}

// CallExec handles the exec request from the master node
func (s *Server) CallExec(req *pb.CallExecRequest, stream grpc.ServerStreamingServer[pb.CallExecResponse]) error {
	return nil
}

// ReceiveChanelConnect handles the channel connection request from the master node
func (s *Server) ReceiveChanelConnect(req *pb.ReceiveChanelConnectRequest, stream grpc.ServerStreamingServer[pb.ReceiveChanelConnectResponse]) error {
	return nil
}

// SendLoader handles the loader request from the master node
func (s *Server) SendLoader(stream grpc.ClientStreamingServer[pb.SendLoaderRequest, pb.SendLoaderResponse]) error {
	return nil
}

// SendAuth handles the auth request from the master node
func (s *Server) SendAuth(ctx context.Context, req *pb.SendAuthRequest) (*pb.SendAuthResponse, error) {
	return &pb.SendAuthResponse{}, nil
}

// SendStoreData handles the store data request from the master node
func (s *Server) SendStoreData(ctx context.Context, req *pb.SendStoreDataRequest) (*pb.SendStoreDataResponse, error) {
	return &pb.SendStoreDataResponse{}, nil
}

// SendTarget handles the target request from the master node
func (s *Server) SendTarget(ctx context.Context, req *pb.SendTargetRequest) (*pb.SendTargetResponse, error) {
	return &pb.SendTargetResponse{}, nil
}
