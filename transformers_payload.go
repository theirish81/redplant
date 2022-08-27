package main

import (
	"bytes"
	"context"
	"github.com/theirish81/gowalker"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type RequestPayloadTransformer struct {
	ActivateOnTags []string
	log            *STLogHelper
	Template       string
	subTemplates   gowalker.SubTemplates
}

func NewRequestPayloadTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*RequestPayloadTransformer, error) {
	t := RequestPayloadTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(context.Background(), params, &t, nil, []string{})
	if err != nil {
		return &t, err
	}
	data, err := os.ReadFile(t.Template)
	if err != nil {
		return &t, err
	}
	templDir := path.Dir(t.Template)
	t.subTemplates = gowalker.NewSubTemplates()
	files, err := os.ReadDir(templDir)
	if err != nil {
		return &t, err
	}
	rootTemplateName := filepath.Base(t.Template)
	t.Template = string(data)
	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && file.Name() != rootTemplateName {
			data, _ = os.ReadFile(path.Join(templDir, file.Name()))
			subTemplateName := file.Name()[0:strings.LastIndex(file.Name(), ".")]
			t.subTemplates[subTemplateName] = string(data)
		}
	}
	return &t, nil
}

func (t *RequestPayloadTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	data, err := template.TemplWithSub(wrapper.Context, t.Template, t.subTemplates, wrapper)
	if err != nil {
		return wrapper, err
	}
	wrapper.Request.Body = io.NopCloser(bytes.NewReader([]byte(data)))
	return wrapper, nil
}

func (t *RequestPayloadTransformer) ShouldExpandRequest() bool {
	return true
}

func (t *RequestPayloadTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *RequestPayloadTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *RequestPayloadTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *RequestPayloadTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

type ResponsePayloadTransformer struct {
	*RequestPayloadTransformer
}

func NewResponsePayloadTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*ResponsePayloadTransformer, error) {
	t, err := NewRequestPayloadTransformer(activateOnTags, logCfg, params)
	return &ResponsePayloadTransformer{t}, err
}

func (t *ResponsePayloadTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	data, err := template.TemplWithSub(wrapper.Context, t.Template, t.subTemplates, wrapper)
	if err != nil {
		return wrapper, err
	}
	wrapper.Response.Header.Del("content-length")
	wrapper.Response.Header.Del("transfer-encoding")
	wrapper.Response.TransferEncoding = make([]string, 0)

	wrapper.Response.Body = io.NopCloser(bytes.NewReader([]byte(data)))
	return wrapper, nil
}
