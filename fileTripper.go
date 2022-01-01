package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// loadFile loads a file (to be used as an upstream server instead of HTTP)
func loadFile(path *url.URL) ([]byte, error) {
	return os.ReadFile(path.Host + path.Path)
}

// getCT brutal, minimalistic content type detection
func getCT(body []byte, fn string) string {
	ct := http.DetectContentType(body)
	if strings.HasPrefix(ct, "text/plain") {
		if strings.HasSuffix(fn, ".json") {
			return strings.Replace(ct, "text/plain", "application/json", 1)
		}
		if strings.HasSuffix(fn, ".xml") {
			return strings.Replace(ct, "text/plain", "text/xml", 1)
		}
	}
	return ct
}

// FileTrip will receive a request, load a local file based on the information and produce a response
func FileTrip(request *http.Request) (*http.Response, error) {
	response := http.Response{StatusCode: 200, Request: request}
	response.Header = http.Header{}

	body, err := loadFile(request.URL)
	if err != nil {
		return nil, err
	}
	response.Header.Set("content-type", getCT(body, request.URL.Path))
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	return &response, nil
}
