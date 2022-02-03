package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRoundTripperFilter_RoundTrip(t *testing.T) {
	config = Config{}
	config.Network.Upstream.Timeout = "10s"
	config.Network.Upstream.KeepAlive = "5s"
	config.Network.Upstream.IdleConnectionTimeout = "2s"
	config.Network.Upstream.ExpectContinueTimeout = "10s"
	transport := configTransport()
	request := &http.Request{Header: http.Header{}}
	request.URL, _ = url.Parse("https://www.google.com")
	request = ReqWithContext(request, nil, nil)
	GetWrapper(request).Request = request
	response, _ := transport.RoundTrip(request)
	if response == nil {
		t.Error("Could not roundtrip to response")
	}
	if response.Header.Get("Content-Type") == "" {
		t.Error("Response seems invalid")
	}
}
