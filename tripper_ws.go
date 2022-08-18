package main

import (
	"bytes"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"io"
	"net/http"
	"time"
)

func WSTripper(request *http.Request, _ *Rule) (*http.Response, error) {
	socket := websocketproxy.NewProxy(request.URL)
	timeout, _ := time.ParseDuration(config.Network.Upstream.Timeout)
	socket.Dialer = &websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: timeout,
	}
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
