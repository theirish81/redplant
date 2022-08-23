package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"testing"
)

func TestRequestParserTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	transformer, _ := NewRequestParserTransformer([]string{}, nil)
	ux, _ := url.Parse("http://example.com")
	req := NewAPIRequest(&http.Request{URL: ux})
	req.ExpandedBody = []byte(`{"foo":"bar"}`)
	wrapper := APIWrapper{Request: req}
	_, _ = transformer.Transform(&wrapper)
	if wrapper.Request.ParsedBody.(map[string]interface{})["foo"] != "bar" {
		t.Error("request parser did not parse")
	}
}

func TestResponseParserTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	transformer, _ := NewResponseParserTransformer([]string{}, nil)
	ux, _ := url.Parse("http://example.com")
	req := NewAPIRequest(&http.Request{URL: ux})
	res := NewAPIResponse(&http.Response{})
	res.ExpandedBody = []byte(`{"foo":"bar"}`)
	wrapper := APIWrapper{Request: req, Response: res}
	_, _ = transformer.Transform(&wrapper)
	if wrapper.Response.ParsedBody.(map[string]interface{})["foo"] != "bar" {
		t.Error("request parser did not parse")
	}
}
