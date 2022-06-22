package main

import (
	"errors"
	"net"
	"net/http"
	"strings"
)

// IPAddresser will determine which IP address the request is coming from
type IPAddresser struct {
	cidrs []*net.IPNet
}

// NewIPAddresser is the constructor of IPAddresser
func NewIPAddresser() *IPAddresser {
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

// isPrivateAddress will return true if the IP address is private. Note that it may error out if the address
// is not valid at all
func (a *IPAddresser) isPrivateAddress(address string) (bool, error) {
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

// FromRequest will try to extract the IP address of the requesting agent from a request
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

// RealIP will return the real IP address of the requesting agent
func (a *IPAddresser) RealIP(r *http.Request) string {
	return a.FromRequest(r)
}
