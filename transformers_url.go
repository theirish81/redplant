package main

import (
	"net/http"
	"strings"
)

// RequestUrlTransformer will transform the URL based on certain configuration keys
// OldPrefix is the path prefix we want to get rid of
// NewPrefix si the path prefix we want instead of OldPrefix
// Query
type RequestUrlTransformer struct {
	OldPrefix      string `yaml:"oldPrefix"`
	NewPrefix      string `yaml:"newPrefix"`
	Query          Query
	ActivateOnTags []string
}

type Query struct {
	Set    map[string]string
	Remove []string
}

func (t *RequestUrlTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	path := wrapper.Request.URL.Path
	if len(t.OldPrefix) > 0 {
		if strings.HasPrefix(path, t.OldPrefix) {
			wrapper.Request.URL.Path = strings.Replace(path, t.OldPrefix, t.NewPrefix, 1)
		}
	}
	query := wrapper.Request.URL.Query()
	for k, v := range t.Query.Set {
		query.Set(k, v)
	}
	for _, remove := range t.Query.Remove {
		query.Del(remove)
	}
	wrapper.Request.URL.RawQuery = query.Encode()
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

func (t *RequestUrlTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

func NewRequestUrlTransformerFromParams(activateOnTags []string, params map[string]interface{}) (*RequestUrlTransformer, error) {
	transformer := RequestUrlTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &transformer, nil, []string{"Query"})
	return &transformer, err
}
