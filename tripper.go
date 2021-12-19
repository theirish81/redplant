package main

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// configTransport configures the transport
func configTransport() http.RoundTripper {
	timeout,err := time.ParseDuration(config.Network.Upstream.Timeout)
	if err !=nil {
		log.Fatal("Timeout is not in the right format",err,nil)
	}
	keepAlive,err := time.ParseDuration(config.Network.Upstream.KeepAlive)
	if err !=nil {
		log.Fatal("KeepAlive is not in the right format",err,nil)
	}
	idleConnTimeout,err := time.ParseDuration(config.Network.Upstream.IdleConnectionTimeout)
	if err !=nil {
		log.Fatal("IdleConnectionTimeout not in the right format",err,nil)
	}
	expectContinueTimeout,err := time.ParseDuration(config.Network.Upstream.ExpectContinueTimeout)
	if err !=nil {
		log.Fatal("ExpectContinueTimeout not in the right format",err,nil)
	}
	return &RoundTripperFilter{&http.Transport {
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer {
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
		return nil,errors.New("no_mapping")
	}
	if wrapper.Err != nil {
		return nil,wrapper.Err
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

// handleURL transforms the URL based on the rules
func handleURL(rule *Rule, req *http.Request) {
	newUrl, _ := url.Parse(rule.Origin)
	newUrl.Path = newUrl.Path+req.URL.Path
	req.URL = newUrl
	req.Host = req.URL.Host
}

type IPAddresser struct {
	cidrs	[]*net.IPNet
}
func newIPAddresser() *IPAddresser {
	maxCidrBlocks := []string{
		"127.0.0.1/8",    // localhost
		"10.0.0.0/8",     // 24-bit block
		"172.16.0.0/12",  // 20-bit block
		"192.168.0.0/16", // 16-bit block
		"169.254.0.0/16", // link local address
		"::1/128",        // localhost IPv6
		"fc00::/7",       // unique local address IPv6
		"fe80::/10",      // link local address IPv6
	}

	addresser := IPAddresser{}

	addresser.cidrs = make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		addresser.cidrs[i] = cidr
	}
	return &addresser
}

func (a *IPAddresser)  isPrivateAddress(address string) (bool, error) {
	ipAddress := net.ParseIP(address)
	if ipAddress == nil {
		return false, errors.New("address is not valid")
	}

	for i := range a.cidrs {
		if a.cidrs[i].Contains(ipAddress) {
			return true, nil
		}
	}

	return false, nil
}

func (a *IPAddresser) FromRequest(r *http.Request) string {
	xRealIP := r.Header.Get("X-Real-Ip")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	if xRealIP == "" && xForwardedFor == "" {
		var remoteIP string

		if strings.ContainsRune(r.RemoteAddr, ':') {
			remoteIP, _, _ = net.SplitHostPort(r.RemoteAddr)
		} else {
			remoteIP = r.RemoteAddr
		}

		return remoteIP
	}

	for _, address := range strings.Split(xForwardedFor, ",") {
		address = strings.TrimSpace(address)
		isPrivate, err := a.isPrivateAddress(address)
		if !isPrivate && err == nil {
			return address
		}
	}

	return xRealIP
}

func (a *IPAddresser) RealIP(r *http.Request) string {
	return a.FromRequest(r)
}