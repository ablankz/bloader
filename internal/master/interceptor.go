package master

import (
	"context"

	"github.com/ablankz/bloader/internal/encrypt"
	"google.golang.org/grpc"
)

// UnaryClientEncryptInterceptor is a client-side interceptor that encrypts the request and decrypts the response.
func UnaryClientEncryptInterceptor(encrypter encrypt.Encrypter) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if plainReq, ok := req.([]byte); ok {
			encryptedReq, err := encrypter.Encrypt(plainReq)
			if err != nil {
				return err
			}
			req = encryptedReq
		}

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			return err
		}

		if encryptedResp, ok := reply.(string); ok {
			plainResp, err := encrypter.Decrypt(encryptedResp)
			if err != nil {
				return err
			}
			*reply.(*[]byte) = plainResp
		}

		return nil
	}
}

// StreamClientInterceptor is a client-side interceptor that encrypts the request and decrypts the response.
func StreamClientInterceptor(encrypter encrypt.Encrypter) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		stream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			return nil, err
		}
		return &wrappedClientStream{ClientStream: stream, encrypter: encrypter}, nil
	}
}

type wrappedClientStream struct {
	grpc.ClientStream
	encrypter encrypt.Encrypter
}

// SendMsg encrypts the message before sending it.
func (w *wrappedClientStream) SendMsg(m any) error {
	if plainMsg, ok := m.([]byte); ok {
		encryptedMsg, err := w.encrypter.Encrypt(plainMsg)
		if err != nil {
			return err
		}
		return w.ClientStream.SendMsg(encryptedMsg)
	}
	return w.ClientStream.SendMsg(m)
}

// RecvMsg decrypts the message after receiving it.
func (w *wrappedClientStream) RecvMsg(m any) error {
	var encryptedMsg []byte
	if err := w.ClientStream.RecvMsg(&encryptedMsg); err != nil {
		return err
	}
	decryptedMsg, err := w.encrypter.Decrypt(string(encryptedMsg))
	if err != nil {
		return err
	}
	*m.(*[]byte) = decryptedMsg
	return nil
}
