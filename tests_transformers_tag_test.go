package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"testing"
)

func TestTagTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	params := map[string]any{
		"tags": []string{"foo", "bar"},
	}
	transformer, _ := NewTagTransformer(nil, params)
	ux, _ := url.Parse("http://example.com")
	wrapper := APIWrapper{Request: &APIRequest{Request: &http.Request{URL: ux}}}
	_, _ = transformer.Transform(&wrapper)
	if wrapper.Tags[0] != "foo" || wrapper.Tags[1] != "bar" {
		t.Error("wrapper not properly tagged")
	}
}
