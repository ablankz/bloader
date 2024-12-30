package slave

import (
	"fmt"
	"net"

	rpc "buf.build/gen/go/cresplanex/bloader/grpc/go/cresplanex/bloader/v1/bloaderv1grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/ablankz/bloader/internal/container"
	"github.com/ablankz/bloader/internal/logger"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
)

func SlaveRun(ctr *container.Container) error {
	creds, err := credentials.NewServerTLSFromFile(
		ctr.Config.SlaveSetting.Certificate.SlaveCert,
		ctr.Config.SlaveSetting.Certificate.SlaveKey,
	)
	if err != nil {
		return fmt.Errorf("failed to load certificate: %v", err)
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))

	slCtr := slcontainer.NewSlaveContainer(ctr)

	rpc.RegisterBloaderSlaveServiceServer(grpcServer, NewServer(ctr, slCtr))
	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", ctr.Config.SlaveSetting.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	ctr.Logger.Info(ctr.Ctx, "Starting the worker node",
		logger.Value("port", ctr.Config.SlaveSetting.Port))

	if err := grpcServer.Serve(lister); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}
