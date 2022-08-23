package main

import (
	"github.com/bitly/go-simplejson"
	"net/http"
)

type RequestParserTransformer struct {
	ActivateOnTags []string
	log            *STLogHelper
}

func NewRequestParserTransformer(activateOnTags []string, logCfg *STLogConfig) (*RequestParserTransformer, error) {
	return &RequestParserTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}, nil
}

func (t *RequestParserTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering parse request", wrapper, t.log.Debug)
	parsedRequestBody, err := simplejson.NewJson(wrapper.Request.InflatedBody)
	wrapper.Request.ParsedBody = parsedRequestBody.Interface()
	return wrapper, err
}

func (t *RequestParserTransformer) ShouldExpandRequest() bool {
	return true
}

func (t *RequestParserTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *RequestParserTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *RequestParserTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *RequestParserTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
