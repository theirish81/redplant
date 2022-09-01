package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestRequestPayloadTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	template = NewRPTemplate()
	tx, _ := NewRequestPayloadTransformer([]string{}, nil, map[string]any{"template": "etc/templates/main.templ"})
	ux, _ := url.Parse("http://example.com")
	req := NewAPIRequest(&http.Request{URL: ux, Method: "GET"})
	wrapper := APIWrapper{Request: req, Context: context.TODO()}
	_, _ = tx.Transform(&wrapper)
	data, _ := io.ReadAll(wrapper.Request.Body)
	if string(data) != "{\n  \"foo\":\"bar\",\n  \"method\": \"GET\"\n}" {
		t.Error("request payload transformer not working")
	}
}
