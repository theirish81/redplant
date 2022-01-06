package main

import (
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
)

func stringInArray(search string, array []string) bool {
	for _, sx := range array {
		if search == sx {
			return true
		}
	}
	return false
}

func getFieldName(val reflect.Value, index int) string {
	structField := reflect.Indirect(val).Type().Field(index)
	return structField.Name
}

// getEnvs converts environment variables to a map
func getEnvs() *map[string]string {
	envs := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}
	return &envs
}

func hasPrefixes(data string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(data, prefix) {
			return true
		}
	}
	return false
}

func isString(data interface{}) bool {
	return reflect.ValueOf(data).Type().String() == "string"
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func convertMaps(intf interface{}) interface{} {
	switch obj := intf.(type) {
	case map[string]interface{}:
		for k, v := range obj {
			obj[k] = convertMaps(v)
		}
	case map[interface{}]interface{}:
		nuMap := map[string]interface{}{}
		for k, v := range obj {
			nuMap[k.(string)] = convertMaps(v)
		}
		return nuMap
	case []interface{}:
		for index, object := range obj {
			obj[index] = convertMaps(object)
		}
	}
	return intf
}

type IPAddresser struct {
	cidrs []*net.IPNet
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
