package slave

import (
	"context"

	"github.com/ablankz/bloader/internal/encrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryClientEncryptInterceptor is a client-side interceptor that encrypts the request and decrypts the response.
func UnaryServerEncryptInterceptor(encrypter encrypt.Encrypter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if encryptedReq, ok := req.(string); ok {
			plainReq, err := encrypter.Decrypt(encryptedReq)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to decrypt request: %v", err)
			}
			req = plainReq
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		if plainResp, ok := resp.([]byte); ok {
			encryptedResp, err := encrypter.Encrypt(plainResp)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to encrypt response: %v", err)
			}
			resp = encryptedResp
		}

		return resp, nil
	}
}

// StreamClientInterceptor is a client-side interceptor that encrypts the request and decrypts the response.
func StreamServerInterceptor(encrypter encrypt.Encrypter) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			encrypter:    encrypter,
		}
		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	encrypter encrypt.Encrypter
}

// RecvMsg implements the grpc.ServerStream interface.
func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	var encryptedMsg []byte
	if err := w.ServerStream.RecvMsg(&encryptedMsg); err != nil {
		return err
	}
	descryptedMsg, err := w.encrypter.Decrypt(string(encryptedMsg))
	if err != nil {
		return status.Errorf(codes.Internal, "failed to decrypt request: %v", err)
	}
	*m.(*[]byte) = descryptedMsg
	return nil
}

// SendMsg implements the grpc.ServerStream interface.
func (w *wrappedServerStream) SendMsg(m interface{}) error {
	if plainMsg, ok := m.([]byte); ok {
		encryptedMsg, err := w.encrypter.Encrypt(plainMsg)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to encrypt response: %v", err)
		}
		m = encryptedMsg
	}
	return w.ServerStream.SendMsg(m)
}
