package main

import (
	"bytes"
	"github.com/koding/websocketproxy"
	"io/ioutil"
	"net/http"
)

func WSTripper(request *http.Request, _ *Rule) (*http.Response, error) {
	socket := websocketproxy.NewProxy(request.URL)
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
	socket.ServeHTTP(wrapper.ResponseWriter, request)
	response := http.Response{StatusCode: 200, Request: request, Body: ioutil.NopCloser(bytes.NewReader([]byte{}))}
	return &response, nil
}
