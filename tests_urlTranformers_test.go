package main

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestUrlTransformer_Transform(t *testing.T) {
	transformer, _ := NewRequestUrlTransformerFromParams([]string{}, map[string]any{"oldPrefix": "/foo", "newPrefix": "/bar"})
	req := http.Request{}
	req.URL, _ = url.Parse("https://example.com/foo")
	wrapper := APIWrapper{Request: &req}
	_, _ = transformer.Transform(&wrapper)
	if wrapper.Request.URL.String() != "https://example.com/bar" {
		t.Error("Prefix transformation failed")
	}
}
