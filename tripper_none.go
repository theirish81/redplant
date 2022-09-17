package main

import (
	"bytes"
	"io"
	"net/http"
)

// NoneTrip is a tripper which does... nothing. It will return an empty response with status code 200
func NoneTrip(request *http.Request) (*http.Response, error) {
	response := http.Response{StatusCode: 200, Request: request, Uncompressed: true}
	response.Header = http.Header{}
	response.Body = io.NopCloser(bytes.NewReader([]byte{}))
	return &response, nil
}
