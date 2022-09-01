package main

import (
	"errors"
	"net/http"
)

type ResponseStatusTransformer struct {
	ActivateOnTags []string
	Code           int
	log            *STLogHelper
}

func NewResponseStatusTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*ResponseStatusTransformer, error) {
	t := ResponseStatusTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	if s, ok := params["code"].(int); !ok {
		return &t, errors.New("status should be an integer")
	} else {
		t.Code = s
		return &t, nil
	}
}

func (t *ResponseStatusTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	wrapper.Response.StatusCode = t.Code
	return wrapper, nil
}

func (t *ResponseStatusTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *ResponseStatusTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *ResponseStatusTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *ResponseStatusTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *ResponseStatusTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
