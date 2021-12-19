package main

import (
	"errors"
	"regexp"
)

// BarrageRequestTransformer is a transformer that will stop the request if certain preconditions are met
type BarrageRequestTransformer struct {
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
}

// NewBarrageRequestTransformer is the constructor for BarrageRequestTransformer
func NewBarrageRequestTransformer(params map[string]interface{}) (*BarrageRequestTransformer, error) {
	var t BarrageRequestTransformer
	err := DecodeAndTempl(params, &t, nil)
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
	return &t, err
}

// Transform will block the request if the preconditions are not met
func (t *BarrageRequestTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for k, v := range wrapper.Request.Header {
		if t._headerRegexp != nil && t._headerRegexp.MatchString(k+":"+v[0]) {
			return wrapper, errors.New("barraged")
		}
		if t._headerNameRegexp != nil && t._headerNameRegexp.MatchString(k) {
			return wrapper, errors.New("barraged")
		}
		if t._headerValueRegexp != nil && t._headerValueRegexp.MatchString(v[0]) {
			return wrapper, errors.New("barraged")
		}
		if t._bodyRegexp != nil && t._bodyRegexp.Match(wrapper.RequestBody) {
			return wrapper, errors.New("barraged")
		}
	}
	return wrapper, nil
}
