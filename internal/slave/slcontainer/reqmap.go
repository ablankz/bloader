package slcontainer

import "sync"

// RequestConnectionMapper is a struct that represents a request connection mapper
type RequestConnectionMapper struct {
	mu                     *sync.RWMutex
	requestKeyConnectionID *sync.Map
	connectionKeyRequestID *sync.Map
}

// NewRequestConnectionMapper creates a new request connection mapper
func NewRequestConnectionMapper() *RequestConnectionMapper {
	return &RequestConnectionMapper{
		mu:                     &sync.RWMutex{},
		requestKeyConnectionID: &sync.Map{},
		connectionKeyRequestID: &sync.Map{},
	}
}

// RegisterRequestConnection registers a new request connection to the mapper
func (r *RequestConnectionMapper) RegisterRequestConnection(reqID, connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.requestKeyConnectionID.Store(reqID, connectionID)
	currentReqIDs, ok := r.connectionKeyRequestID.Load(connectionID)
	if ok {
		r.connectionKeyRequestID.Store(connectionID, append(currentReqIDs.([]string), reqID))
	} else {
		r.connectionKeyRequestID.Store(connectionID, []string{reqID})
	}
}

// GetConnectionID returns the connection ID for the request
func (r *RequestConnectionMapper) GetConnectionID(reqID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if connectionID, ok := r.requestKeyConnectionID.Load(reqID); ok {
		return connectionID.(string), true
	}
	return "", false
}

// GetRequestID returns the request ID for the connection
func (r *RequestConnectionMapper) GetRequestID(connectionID string) ([]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if reqIDs, ok := r.connectionKeyRequestID.Load(connectionID); ok {
		return reqIDs.([]string), true
	}
	return nil, false
}

// DeleteRequestConnection deletes the request connection from the mapper
func (r *RequestConnectionMapper) DeleteRequestConnection(connectionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if reqIDs, ok := r.connectionKeyRequestID.Load(connectionID); ok {
		for _, reqID := range reqIDs.([]string) {
			r.requestKeyConnectionID.Delete(reqID)
		}
	}
	r.connectionKeyRequestID.Delete(connectionID)
}

// DeleteRequest deletes the request from the mapper
func (r *RequestConnectionMapper) DeleteRequest(reqID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if connectionID, ok := r.requestKeyConnectionID.Load(reqID); ok {
		if reqIDs, ok := r.connectionKeyRequestID.Load(connectionID); ok {
			var newReqIDs []string
			for _, id := range reqIDs.([]string) {
				if id != reqID {
					newReqIDs = append(newReqIDs, id)
				}
			}
			r.connectionKeyRequestID.Store(connectionID, newReqIDs)
		}
	}
	r.requestKeyConnectionID.Delete(reqID)
}
