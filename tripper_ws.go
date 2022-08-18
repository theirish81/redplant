package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"io"
	"net/http"
	"time"
)

// WSTripper is the tripper for websocket requests
func WSTripper(request *http.Request, _ *Rule) (*http.Response, error) {
	// create a new websocket proxy for the provided URL
	socket := websocketproxy.NewProxy(request.URL)
	// set a timeout to the connection. It is the same as the Upstream.Timeout
	timeout, _ := time.ParseDuration(config.Network.Upstream.Timeout)
	socket.Dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: timeout,
	}
	// During upgrades, we need to make sure a certain set of incoming headers are not overwritten
	socket.Director = func(incoming *http.Request, out http.Header) {
		for k, v := range incoming.Header {
			switch k {
			case "Sec-Websocket-Key", "Connection", "Sec-Websocket-Version", "Sec-Websocket-Extensions", "Upgrade":
			default:
				out.Set(k, v[0])
			}
		}
	}
	wrapper := GetWrapper(request)

	// setting the connection as "hijacked". No further writes are possible in this response
	wrapper.Hijacked = true

	socket.ServeHTTP(wrapper.ResponseWriter, request)
	response := http.Response{StatusCode: 200, Request: request, Body: io.NopCloser(bytes.NewReader([]byte{}))}
	return &response, nil
}
