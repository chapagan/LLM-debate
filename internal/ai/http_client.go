package ai

import (
	"net"
	"net/http"
	"time"
)

const defaultResponseHeaderTimeout = 30 * time.Second

func defaultHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = defaultResponseHeaderTimeout
	transport.DialContext = (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext
	return &http.Client{
		Timeout:   0, // streaming responses can outlive a fixed total timeout
		Transport: transport,
	}
}
