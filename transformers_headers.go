package main

import "net/http"

// RequestHeaderTransformer transforms the request header by setting or removing headers
type RequestHeaderTransformer struct {
	Set            StringMap
	Remove         []string
	ActivateOnTags []string
	log            *STLogHelper
}

func (t *RequestHeaderTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering request header transformation", wrapper, t.log.Debug)
	for hk, hv := range t.Set {
		hv, err := wrapper.Templ(hv)
		if err != nil {
			return wrapper, err
		}
		wrapper.Request.Header.Set(hk, hv)
	}
	for _, rem := range t.Remove {
		wrapper.Request.Header.Del(rem)
	}
	return wrapper, nil
}

func (t *RequestHeaderTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *RequestHeaderTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *RequestHeaderTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *RequestHeaderTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

// NewRequestHeadersTransformerFromParams is the constructor for RequestHeaderTransformer
func NewRequestHeadersTransformerFromParams(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*RequestHeaderTransformer, error) {
	t := RequestHeaderTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := DecodeAndTempl(params, &t, nil, []string{"Set"})
	return &t, err
}

// ResponseHeaderTransformer transforms the response header by setting or removing headers
type ResponseHeaderTransformer struct {
	Set            StringMap
	Remove         []string
	ActivateOnTags []string
	log            *STLogHelper
}

// NewResponseHeadersTransformerFromParams is the constructor for ResponseHeaderTransformer
func NewResponseHeadersTransformerFromParams(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*ResponseHeaderTransformer, error) {
	t := ResponseHeaderTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := DecodeAndTempl(params, &t, nil, []string{"Set"})
	return &t, err
}

func (t *ResponseHeaderTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering response header transformation", wrapper, t.log.Debug)
	for hk, hv := range t.Set {
		wrapper.Response.Header.Set(hk, hv)
	}
	for _, rem := range t.Remove {
		wrapper.Request.Header.Del(rem)
	}
	return wrapper, nil
}

func (t *ResponseHeaderTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *ResponseHeaderTransformer) HandleError(_ *http.ResponseWriter) {

}

func (t *ResponseHeaderTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *ResponseHeaderTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *ResponseHeaderTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
