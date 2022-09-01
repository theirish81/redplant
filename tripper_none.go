package main

import (
	"bytes"
	"io"
	"net/http"
)

func NoneTrip(request *http.Request) (*http.Response, error) {
	response := http.Response{StatusCode: 200, Request: request}
	response.Header = http.Header{}
	response.Body = io.NopCloser(bytes.NewReader([]byte{}))
	return &response, nil
}
