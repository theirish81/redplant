package main

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"
)

// configTransport configures the transport
func configTransport() http.RoundTripper {
	timeout, err := time.ParseDuration(config.Network.Upstream.Timeout)
	if err != nil {
		log.Fatal("Timeout is not in the right format", err, nil)
	}
	keepAlive, err := time.ParseDuration(config.Network.Upstream.KeepAlive)
	if err != nil {
		log.Fatal("KeepAlive is not in the right format", err, nil)
	}
	idleConnTimeout, err := time.ParseDuration(config.Network.Upstream.IdleConnectionTimeout)
	if err != nil {
		log.Fatal("IdleConnectionTimeout not in the right format", err, nil)
	}
	expectContinueTimeout, err := time.ParseDuration(config.Network.Upstream.ExpectContinueTimeout)
	if err != nil {
		log.Fatal("ExpectContinueTimeout not in the right format", err, nil)
	}
	return &RoundTripperFilter{&http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).DialContext,
		MaxIdleConns:          config.Network.Upstream.MaxIdleConnections,
		IdleConnTimeout:       idleConnTimeout,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: expectContinueTimeout,
		TLSClientConfig:       &tls.Config{},
	}}
}

// RoundTripperFilter is a wrapper around Transport. We need to do this to handle errors or unusual upstreams
type RoundTripperFilter struct {
	parent *http.Transport
}

// RoundTrip wraps the default RoundTrip to handle errors and unusual upstreams
func (rtf *RoundTripperFilter) RoundTrip(r *http.Request) (*http.Response, error) {
	wrapper := GetWrapper(r)
	if wrapper == nil {
		return nil, errors.New("no_mapping")
	}
	if wrapper.Err != nil {
		return nil, wrapper.Err
	}
	scheme := wrapper.Request.URL.Scheme
	switch scheme {
	case "file":
		return FileTrip(r)
	case "postgres":
		return DBTrip(r, wrapper.Rule)
	case "mysql":
		return DBTrip(r, wrapper.Rule)
	default:
		return rtf.parent.RoundTrip(r)
	}
}
