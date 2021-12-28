package main

import (
	"math/rand"
	"net/http"
	"time"
)

// DelayTransformer will slow the request down by a certain amount
type DelayTransformer struct {
	_min           time.Duration
	_max           time.Duration
	Min            string `mapstructure:"min"`
	Max            string `mapstructure:"max"`
	ActivateOnTags []string
}

func NewDelayTransformer(activateOnTags []string, params map[string]interface{}) (*DelayTransformer, error) {
	t := DelayTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	t._min, err = time.ParseDuration(t.Min)
	if err != nil {
		return nil, err
	}
	t._max, err = time.ParseDuration(t.Max)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *DelayTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	timeRange := t._max.Nanoseconds() - t._min.Nanoseconds()
	nanos := t._min.Nanoseconds() + rand.Int63n(timeRange)
	time.Sleep(time.Duration(nanos))
	return wrapper, nil
}

func (t *DelayTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *DelayTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *DelayTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *DelayTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *DelayTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
