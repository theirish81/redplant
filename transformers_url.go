package main

import (
	"context"
	"net/http"
	"strings"
)

// RequestUrlTransformer will transform the URL based on certain configuration keys
// OldPrefix is the path prefix we want to get rid of
// NewPrefix si the path prefix we want instead of OldPrefix
// Query is a set of instructions on operations to perform against the query
type RequestUrlTransformer struct {
	OldPrefix      string `yaml:"oldPrefix"`
	NewPrefix      string `yaml:"newPrefix"`
	Query          Query
	ActivateOnTags []string
	log            *STLogHelper
}

// Query is a collection of operations to apply to the query
// Set is a map of query params to set
// Remove is an array of query params to remove
type Query struct {
	Set    StringMap
	Remove []string
}

func (t *RequestUrlTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("request url transformer triggered", wrapper, t.log.Debug)
	path := wrapper.Request.URL.Path
	// if we have an OldPrefix...
	if len(t.OldPrefix) > 0 {
		// ... then if the path has that OldPrefix
		if strings.HasPrefix(path, t.OldPrefix) {
			t.log.Log("url transformed", wrapper, t.log.Debug)
			// we replace the OldPrefix with the NewPrefix
			wrapper.Request.URL.Path = strings.Replace(path, t.OldPrefix, t.NewPrefix, 1)
		}
	}
	query := wrapper.Request.URL.Query()
	// for every set instruction for the query, we set a param
	for k, v := range t.Query.Set {
		query.Set(k, v)
	}
	// for every remove instruction for the query, we remove a pram
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

// NewRequestUrlTransformerFromParams is the constructor for RequestUrlTransformer
func NewRequestUrlTransformerFromParams(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*RequestUrlTransformer, error) {
	transformer := RequestUrlTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(context.Background(), params, &transformer, nil, []string{"Query"})
	return &transformer, err
}
