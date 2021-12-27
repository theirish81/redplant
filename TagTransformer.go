package main

import "net/http"

type TagTransformer struct {
	Tags []string
}

func (t *TagTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *TagTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *TagTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *TagTransformer) HandleError(_ *http.ResponseWriter) {}

// Transform adds the JWT token to the request
func (t *TagTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for _, tag := range t.Tags {
		val, err := Templ(tag, wrapper)
		if err != nil {
			return nil, err
		}
		if val != "" && val != "<no value>" {
			wrapper.Tags = append(wrapper.Tags, val)
		}
	}
	return wrapper, nil
}

func NewTagTransformer(params map[string]interface{}) (*TagTransformer, error) {
	t := TagTransformer{}
	err := DecodeAndTempl(params, &t, nil, []string{"Tags"})
	return &t, err
}
