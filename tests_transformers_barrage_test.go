package main

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestBarrageRequestTransformer_Transform(t *testing.T) {
	transformer, _ := NewBarrageRequestTransformer([]string{}, nil, map[string]any{"headerValueRegexp": "log4j.*"})
	ux, _ := url.Parse("http://www.example.com")
	req := http.Request{Header: http.Header{}, URL: ux}
	wrapper := APIWrapper{Request: NewAPIRequest(&req)}
	_, err := transformer.Transform(&wrapper)
	if err != nil {
		t.Error("Something went wrong during no-barrage")
	}

	req.Header.Set("foo", "log4j123")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for header value")
	}
	req.Header.Del("foo")
	transformer, _ = NewBarrageRequestTransformer([]string{}, nil, map[string]any{"headerNameRegexp": "Log4j.*"})
	req.Header.Set("log4jFoo", "123")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for header name")
	}
	req.Header.Del("log4jFoo")
	transformer, _ = NewBarrageRequestTransformer([]string{}, nil, map[string]any{"headerRegexp": "Foo:log4j.*"})
	req.Header.Set("foo", "log4jBananas")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for header pair")
	}
	req.Body = io.NopCloser(bytes.NewReader([]byte("foo bar foo")))
	wrapper.ExpandRequest()
	transformer, _ = NewBarrageRequestTransformer([]string{}, nil, map[string]any{"bodyRegexp": ".*bar.*"})
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for body")
	}
}
