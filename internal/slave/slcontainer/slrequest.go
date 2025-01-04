package slcontainer

import (
	"context"
	"sync"

	pb "buf.build/gen/go/cresplanex/bloader/protocolbuffers/go/cresplanex/bloader/v1"
	"github.com/ablankz/bloader/internal/utils"
)

// RequestTermCaster is an interface that represents a request term caster
type RequestTermCaster struct {
	mu  *sync.RWMutex
	req map[string]chan<- struct{}
}

// NewRequestTermCaster creates a new request term caster
func NewRequestTermCaster() *RequestTermCaster {
	return &RequestTermCaster{
		mu:  &sync.RWMutex{},
		req: make(map[string]chan<- struct{}),
	}
}

// RegisterRequest registers a new request to the caster
func (r *RequestTermCaster) RegisterRequest(reqID string) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan struct{})
	r.req[reqID] = ch
	return ch
}

// Cast casts a term to the request
func (r *RequestTermCaster) Cast(reqID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if ch, ok := r.req[reqID]; ok {
		close(ch)
	}
}

// ReceiveChanelRequestContainer is a struct that represents a container for the receive chanel requests
type ReceiveChanelRequestContainer struct {
	mu         *sync.RWMutex
	termCaster *RequestTermCaster
	ReqChan    chan *pb.ReceiveChanelConnectResponse
}

// NewReceiveChanelRequestContainer creates a new request container
func NewReceiveChanelRequestContainer() *ReceiveChanelRequestContainer {
	return &ReceiveChanelRequestContainer{
		mu:         &sync.RWMutex{},
		termCaster: NewRequestTermCaster(),
		ReqChan:    make(chan *pb.ReceiveChanelConnectResponse),
	}
}

// SendLoaderResourceRequests sets the loader requests channel
func (r *ReceiveChanelRequestContainer) SendLoaderResourceRequests(
	ctx context.Context,
	connectionID string,
	mapper *RequestConnectionMapper,
	req LoaderResourceRequest,
) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := utils.GenerateUniqueID()

	pbReq := &pb.ReceiveChanelConnectResponse{
		RequestId:   requestID,
		RequestType: pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_LOADER,
		Request: &pb.ReceiveChanelConnectResponse_LoaderResourceRequest{
			LoaderResourceRequest: &pb.ReceiveChanelConnectLoaderResourceRequest{
				LoaderId: req.LoaderID,
			},
		},
	}

	select {
	case <-ctx.Done():
		return nil
	case r.ReqChan <- pbReq: // nothing
	}

	mapper.RegisterRequestConnection(requestID, connectionID)

	return r.termCaster.RegisterRequest(requestID)
}

// SendAuthResourceRequests sets the auth requests channel
func (r *ReceiveChanelRequestContainer) SendAuthResourceRequests(
	ctx context.Context,
	connectionID string,
	mapper *RequestConnectionMapper,
	req AuthResourceRequest,
) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := utils.GenerateUniqueID()

	pbReq := &pb.ReceiveChanelConnectResponse{
		RequestId:   requestID,
		RequestType: pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_AUTH,
		Request: &pb.ReceiveChanelConnectResponse_AuthResourceRequest{
			AuthResourceRequest: &pb.ReceiveChanelConnectAuthResourceRequest{
				AuthId:    req.AuthID,
				IsDefault: req.IsDefault,
			},
		},
	}

	select {
	case <-ctx.Done():
		return nil
	case r.ReqChan <- pbReq: // nothing
	}

	mapper.RegisterRequestConnection(requestID, connectionID)

	return r.termCaster.RegisterRequest(requestID)
}

// SendStore send store requests
func (r *ReceiveChanelRequestContainer) SendStore(
	ctx context.Context,
	connectionID string,
	mapper *RequestConnectionMapper,
	req StoreDataRequest,
) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := utils.GenerateUniqueID()

	strData := make([]*pb.StoreData, 0, len(req.StoreData))

	for _, storeData := range req.StoreData {
		strData = append(strData, &pb.StoreData{
			BucketId: storeData.BucketID,
			StoreKey: storeData.StoreKey,
			Data:     storeData.Data,
			Encryption: &pb.Encryption{
				Enabled:   storeData.Encryption.Enabled,
				EncryptId: storeData.Encryption.EncryptID,
			},
		})
	}

	pbReq := &pb.ReceiveChanelConnectResponse{
		RequestId:   requestID,
		RequestType: pb.RequestType_REQUEST_TYPE_STORE,
		Request: &pb.ReceiveChanelConnectResponse_Store{
			Store: &pb.ReceiveChanelConnectStore{
				StoreData: strData,
			},
		},
	}

	select {
	case <-ctx.Done():
	case r.ReqChan <- pbReq: // nothing
	}

	mapper.RegisterRequestConnection(requestID, connectionID)

	return r.termCaster.RegisterRequest(requestID)
}

// SendStoreResourceRequests sets the store requests channel
func (r *ReceiveChanelRequestContainer) SendStoreResourceRequests(
	ctx context.Context,
	connectionID string,
	mapper *RequestConnectionMapper,
	req StoreResourceRequest,
) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := utils.GenerateUniqueID()

	importReqs := make([]*pb.StoreImportRequest, 0, len(req.Requests))

	for _, importReq := range req.Requests {
		importReqs = append(importReqs, &pb.StoreImportRequest{
			BucketId: importReq.BucketID,
			StoreKey: importReq.StoreKey,
		})
	}

	pbReq := &pb.ReceiveChanelConnectResponse{
		RequestId:   requestID,
		RequestType: pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_STORE,
		Request: &pb.ReceiveChanelConnectResponse_StoreResourceRequest{
			StoreResourceRequest: &pb.ReceiveChanelConnectStoreResourceRequest{
				StoreImportRequest: importReqs,
			},
		},
	}

	select {
	case <-ctx.Done():
		return nil
	case r.ReqChan <- pbReq: // nothing
	}

	mapper.RegisterRequestConnection(requestID, connectionID)

	return r.termCaster.RegisterRequest(requestID)
}

// SendTargetResourceRequests sets the target requests channel
func (r *ReceiveChanelRequestContainer) SendTargetResourceRequests(
	ctx context.Context,
	connectionID string,
	mapper *RequestConnectionMapper,
	req TargetResourceRequest,
) <-chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()

	requestID := utils.GenerateUniqueID()

	pbReq := &pb.ReceiveChanelConnectResponse{
		RequestId:   requestID,
		RequestType: pb.RequestType_REQUEST_TYPE_REQUEST_RESOURCE_TARGET,
		Request: &pb.ReceiveChanelConnectResponse_TargetResourceRequest{
			TargetResourceRequest: &pb.ReceiveChanelConnectTargetResourceRequest{
				TargetId: req.TargetID,
			},
		},
	}

	select {
	case <-ctx.Done():
		return nil
	case r.ReqChan <- pbReq: // nothing
	}

	mapper.RegisterRequestConnection(requestID, connectionID)

	return r.termCaster.RegisterRequest(requestID)
}

// Cast casts a term to the request
func (r *ReceiveChanelRequestContainer) Cast(reqID string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.termCaster.Cast(reqID)
}
