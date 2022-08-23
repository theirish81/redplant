package main

import (
	"encoding/json"
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
	var parsedRequestBody interface{}
	err := json.Unmarshal(wrapper.Request.ExpandedBody, &parsedRequestBody)
	wrapper.Request.ParsedBody = parsedRequestBody
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

type ResponseParserTransformer struct {
	ActivateOnTags []string
	log            *STLogHelper
}

func NewResponseParserTransformer(activateOnTags []string, logCfg *STLogConfig) (*ResponseParserTransformer, error) {
	return &ResponseParserTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}, nil
}

func (t *ResponseParserTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering parse response", wrapper, t.log.Debug)
	var parsedResponseBody interface{}
	err := json.Unmarshal(wrapper.Response.ExpandedBody, &parsedResponseBody)
	wrapper.Response.ParsedBody = parsedResponseBody
	return wrapper, err
}

func (t *ResponseParserTransformer) ShouldExpandRequest() bool {
	return true
}

func (t *ResponseParserTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *ResponseParserTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *ResponseParserTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *ResponseParserTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
