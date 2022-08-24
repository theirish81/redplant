package main

import (
	"context"
	"net/http"
)

// TagTransformer will apply a tag to the request envelope
type TagTransformer struct {
	Tags []string
	log  *STLogHelper
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
	t.log.Log("triggering tag transformer", wrapper, t.log.Debug)
	for _, tag := range t.Tags {
		val, err := template.Templ(wrapper.Context, tag, wrapper)
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
func NewTagTransformer(logCfg *STLogConfig, params map[string]any) (*TagTransformer, error) {
	t := TagTransformer{log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(context.Background(), params, &t, nil, []string{"Tags"})
	return &t, err
}
