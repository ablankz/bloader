package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// DelayedTransport is an http.RoundTripper that introduces a delay before
type DelayedTransport struct {
	Transport http.RoundTripper
	Delay     time.Duration
}

// RoundTrip executes a single HTTP transaction
func (d *DelayedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	time.Sleep(d.Delay)
	resMap := map[string]interface{}{
		"status":  "ok",
		"message": "delayed response",
	}
	res, _ := json.Marshal(resMap)
	reader := io.NopCloser(bytes.NewReader(res))
	return &http.Response{
		StatusCode: 200,
		Body:       reader,
	}, nil
	// return d.Transport.RoundTrip(req)
}
