package master

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"

	rpc "buf.build/gen/go/cresplanex/bloader/grpc/go/cresplanex/bloader/v1/bloaderv1grpc"
	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/encrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectionMapData is a struct that holds the connection information.
type ConnectionMapData struct {
	connectionID string
	conn         *grpc.ClientConn
	cli          rpc.BloaderSlaveServiceClient
}

// ConnectionContainer is a struct that holds the connection information.
type ConnectionContainer struct {
	mu     *sync.RWMutex
	conMap map[string]*ConnectionMapData // Key: slaveID
}

// NewConnectMap creates a new ConnectMap.
func NewConnectMap() *ConnectionContainer {
	return &ConnectionContainer{
		mu:     &sync.RWMutex{},
		conMap: make(map[string]*ConnectionMapData),
	}
}

// Connect adds a connection to the map.
func (c *ConnectionContainer) Connect(
	ctx context.Context,
	env string,
	encryptCtr encrypt.EncrypterContainer,
	slaveID string,
	conInfo SlaveConnect,
) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, slave := range conInfo.Slaves {
		grpcDialOptions := []grpc.DialOption{}
		if slave.Certificate.Enabled {
			b, err := os.ReadFile(slave.Certificate.CACert)
			if err != nil {
				return fmt.Errorf("credentials: failed to read CA certificate: %v", err)
			}
			cp := x509.NewCertPool()
			if !cp.AppendCertsFromPEM(b) {
				return fmt.Errorf("credentials: failed to append certificates")
			}
			creds := credentials.NewTLS(&tls.Config{
				ServerName:         slave.Certificate.ServerNameOverride,
				InsecureSkipVerify: slave.Certificate.InsecureSkipVerify,
				RootCAs:            cp,
			})
			grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(creds))
		} else {
			grpcDialOptions = append(grpcDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		if slave.Encrypt.Enabled {
			encrypter, ok := encryptCtr[slave.Encrypt.EncryptID]
			if !ok {
				return fmt.Errorf("encrypter not found: %s", slave.Encrypt.EncryptID)
			}
			grpcDialOptions = append(
				grpcDialOptions,
				grpc.WithUnaryInterceptor(UnaryClientEncryptInterceptor(encrypter)),
				grpc.WithStreamInterceptor(StreamClientInterceptor(encrypter)),
			)
		}

		conn, err := grpc.NewClient(slave.URI, grpcDialOptions...)
		if err != nil {
			return fmt.Errorf("failed to connect to slave: %v", err)
		}

		cli := rpc.NewBloaderSlaveServiceClient(conn)

		conReq := &pb.ConnectRequest{
			Environment: env,
		}

		res, err := cli.Connect(ctx, conReq)
		if err != nil {
			return fmt.Errorf("failed to connect to slave: %v", err)
		}

		c.conMap[slave.ID] = &ConnectionMapData{
			connectionID: res.ConnectionId,
			conn:         conn,
			cli:          cli,
		}
	}

	return nil
}

// Disconnect removes a connection from the map.
func (c *ConnectionContainer) Disconnect(slaveID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, ok := c.conMap[slaveID]
	if !ok {
		return fmt.Errorf("connection not found: %s", slaveID)
	}

	disReq := &pb.DisconnectRequest{
		ConnectionId: conn.connectionID,
	}

	_, err := conn.cli.Disconnect(context.Background(), disReq)
	if err != nil {
		return fmt.Errorf("failed to disconnect from slave: %v", err)
	}
	if err := conn.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %v", err)
	}
	delete(c.conMap, slaveID)

	return nil
}
