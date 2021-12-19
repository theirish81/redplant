package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestBarrageRequestTransformer_Transform(t *testing.T) {
	transformer, _ := NewBarrageRequestTransformer(map[string]interface{}{"headerValueRegexp": "log4j.*"})
	req := http.Request{Header: http.Header{}}
	wrapper := APIWrapper{Request: &req}
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
	transformer, _ = NewBarrageRequestTransformer(map[string]interface{}{"headerNameRegexp": "Log4j.*"})
	req.Header.Set("log4jFoo", "123")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for header name")
	}
	req.Header.Del("log4jFoo")
	transformer, _ = NewBarrageRequestTransformer(map[string]interface{}{"headerRegexp": "Foo:log4j.*"})
	req.Header.Set("foo", "log4jBananas")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for header pair")
	}
	req.Body = ioutil.NopCloser(bytes.NewReader([]byte("foo bar foo")))
	wrapper.ExpandRequest()
	transformer, _ = NewBarrageRequestTransformer(map[string]interface{}{"bodyRegexp": ".*bar.*"})
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Barrage did not trigger for body")
	}
}
