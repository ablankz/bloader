package utils

import (
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
	return d.Transport.RoundTrip(req)
}
