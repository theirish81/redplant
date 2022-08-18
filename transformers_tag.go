package main

import "net/http"

// TagTransformer will apply a tag to the request envelope
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

func (t *TagTransformer) IsActive(_ *APIWrapper) bool {
	return true
}

// Transform adds the JWT token to the request
func (t *TagTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for _, tag := range t.Tags {
		val, err := Templ(tag, wrapper)
		if err != nil {
			return nil, err
		}
		// we don't want a <no value> to appear in the tags, so in case that's what's happening, we just don't
		// append the tag
		if val != "" && val != "<no value>" {
			wrapper.Tags = append(wrapper.Tags, val)
		}
	}
	return wrapper, nil
}

// NewTagTransformer is the constructor for TagTransformer
func NewTagTransformer(params map[string]interface{}) (*TagTransformer, error) {
	t := TagTransformer{}
	err := DecodeAndTempl(params, &t, nil, []string{"Tags"})
	return &t, err
}
