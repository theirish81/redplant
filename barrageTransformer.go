package main

import (
	"errors"
	"net/http"
	"regexp"
)

// BarrageTransformer is a transformer that will stop the request if certain preconditions are met
type BarrageTransformer struct {
	// HeaderNameRegexp is a regular expression for a forbidden header name
	HeaderNameRegexp string
	// HeaderValueRegexp is a regular expression for a forbidden header name
	HeaderValueRegexp string
	// HeaderRegexp is a regular expression for a forbidden full header as in name:value
	HeaderRegexp string
	// BodyRegexp is a regular expression for a forbidden body
	BodyRegexp         string
	_headerNameRegexp  *regexp.Regexp
	_headerValueRegexp *regexp.Regexp
	_headerRegexp      *regexp.Regexp
	_bodyRegexp        *regexp.Regexp
	response           bool
}

// NewBarrageRequestTransformer is the constructor for BarrageTransformer
func NewBarrageRequestTransformer(params map[string]interface{}) (*BarrageTransformer, error) {
	var t BarrageTransformer
	err := DecodeAndTempl(params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	if t.HeaderNameRegexp != "" {
		t._headerNameRegexp, err = regexp.Compile(t.HeaderRegexp)
		if err != nil {
			return nil, err
		}
	}
	if t.HeaderValueRegexp != "" {
		t._headerValueRegexp, err = regexp.Compile(t.HeaderValueRegexp)
		if err != nil {
			return nil, err
		}
	}
	if t.HeaderRegexp != "" {
		t._headerRegexp, err = regexp.Compile(t.HeaderRegexp)
		if err != nil {
			return nil, err
		}
	}
	if t.BodyRegexp != "" {
		t._bodyRegexp, err = regexp.Compile(t.BodyRegexp)
		if err != nil {
			return nil, err
		}
	}
	t.response = false
	return &t, err
}

func NewBarrageResponseTransformer(params map[string]interface{}) (*BarrageTransformer, error) {
	transformer, err := NewBarrageRequestTransformer(params)
	transformer.response = true
	return transformer, err
}

// Transform will block the request if the preconditions are not met
func (t *BarrageTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	var headers *http.Header
	var body *[]byte
	if t.response {
		headers = &wrapper.Response.Header
		body = &wrapper.ResponseBody
	} else {
		headers = &wrapper.Request.Header
		body = &wrapper.RequestBody
	}
	for k, v := range *headers {
		if t._headerRegexp != nil && t._headerRegexp.MatchString(k+":"+v[0]) {
			return wrapper, errors.New("barraged")
		}
		if t._headerNameRegexp != nil && t._headerNameRegexp.MatchString(k) {
			return wrapper, errors.New("barraged")
		}
		if t._headerValueRegexp != nil && t._headerValueRegexp.MatchString(v[0]) {
			return wrapper, errors.New("barraged")
		}
	}
	if t._bodyRegexp != nil && t._bodyRegexp.Match(*body) {
		return wrapper, errors.New("barraged")
	}
	return wrapper, nil
}

func (t *BarrageTransformer) ErrorMatches(err error) bool {
	return err.Error() == "barraged"
}

func (t *BarrageTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(403)
}

func (t *BarrageTransformer) ShouldExpandRequest() bool {
	return t._bodyRegexp != nil
}

func (t *BarrageTransformer) ShouldExpandResponse() bool {
	return t._bodyRegexp != nil
}
