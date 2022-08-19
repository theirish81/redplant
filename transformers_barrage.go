package main

import (
	"errors"
	"net/http"
	"regexp"
)

// BarrageTransformer is a transformer that will stop the request if certain preconditions are met
// HeaderNameRegexp is a regular expression for a forbidden header name
// HeaderValueRegexp is a regular expression for a forbidden header name
// HeaderRegexp is a regular expression for a forbidden full header as in name:value
// BodyRegexp is a regular expression for a forbidden body
// _headerNameRegexp is the compiled version of HeaderNameRegexp
// _headerValueRegexp is the compiled version of HeaderValueRegexp
// _bodyRegexp is the compiled version of BodyRegexp
// response if set to true means that this is a response transformer
// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
type BarrageTransformer struct {
	HeaderNameRegexp   string
	HeaderValueRegexp  string
	HeaderRegexp       string
	BodyRegexp         string
	_headerNameRegexp  *regexp.Regexp
	_headerValueRegexp *regexp.Regexp
	_headerRegexp      *regexp.Regexp
	_bodyRegexp        *regexp.Regexp
	response           bool
	ActivateOnTags     []string
}

// NewBarrageRequestTransformer is the constructor for BarrageTransformer
func NewBarrageRequestTransformer(activateOnTags []string, params map[string]any) (*BarrageTransformer, error) {
	t := BarrageTransformer{ActivateOnTags: activateOnTags}
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

// NewBarrageResponseTransformer is the constructor for the BarrageTransformer dedicated to the request
func NewBarrageResponseTransformer(activateOnTags []string, params map[string]any) (*BarrageTransformer, error) {
	transformer, err := NewBarrageRequestTransformer(activateOnTags, params)
	transformer.response = true
	return transformer, err
}

// Transform will block the request if the preconditions are not met
func (t *BarrageTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	var headers *http.Header
	var body *[]byte
	// As the implementation of this transformer is identical for the request and the response, we determine
	// whether we're targeting the request or the response based on the presence of the response,
	// So here we collect headers and body from the right source
	if t.response {
		headers = &wrapper.Response.Header
		body = &wrapper.ResponseBody
	} else {
		headers = &wrapper.Request.Header
		body = &wrapper.RequestBody
	}
	// For each header, we determine whether one of the regexp matches. If one does, we barrage.
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
	// Similarly, we determine whether the body matches the regexp. If it does, we barrage
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

func (t *BarrageTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
