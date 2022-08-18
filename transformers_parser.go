package main

import (
	"github.com/bitly/go-simplejson"
	"net/http"
)

type RequestParserTransformer struct {
	ActivateOnTags []string
}

func NewRequestParserTransformer(activateOnTags []string) (*RequestParserTransformer, error) {
	return &RequestParserTransformer{ActivateOnTags: activateOnTags}, nil
}

func (t *RequestParserTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	parsedRequestBody, err := simplejson.NewJson(wrapper.RequestBody)
	parsedRequestBody.Interface()
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
