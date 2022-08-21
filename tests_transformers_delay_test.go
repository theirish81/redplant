package main

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestDelayTransformer_Transform(t *testing.T) {
	transformer, _ := NewDelayTransformer([]string{}, nil, map[string]any{"min": "1s", "max": "3s"})
	ux, _ := url.Parse("http://www.example.com")
	wrapper := APIWrapper{Request: &http.Request{URL: ux}}
	before := time.Now()
	_, _ = transformer.Transform(&wrapper)
	after := time.Now()
	if after.Sub(before).Seconds() > 3 || after.Sub(before).Seconds() < 1 {
		t.Error("Delay is not working as expected")
	}
}
