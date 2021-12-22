package main

import (
	"net/http"
	"strings"
)

type RequestUrlTransformer struct {
	OldPrefix string `yaml:"oldPrefix" mapstructure:"oldPrefix"`

	NewPrefix string `yaml:"newPrefix" mapstructure:"newPrefix"`
}

func (t *RequestUrlTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	path := wrapper.Request.URL.Path
	if strings.HasPrefix(path, t.OldPrefix) {
		wrapper.Request.URL.Path = strings.Replace(path, t.OldPrefix, t.NewPrefix, 1)
	}
	return wrapper, nil
}

func (t *RequestUrlTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *RequestUrlTransformer) HandleError(_ *http.ResponseWriter) {
}

func (t *RequestUrlTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *RequestUrlTransformer) ShouldExpandResponse() bool {
	return false
}

func NewRequestUrlTransformerFromParams(params map[string]interface{}) (*RequestUrlTransformer, error) {
	var transformer RequestUrlTransformer
	err := DecodeAndTempl(params, &transformer, nil, []string{})
	return &transformer, err
}
