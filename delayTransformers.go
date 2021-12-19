package main

import (
	"math/rand"
	"time"
)

// DelayTransformer will slow the request down by a certain amount
type DelayTransformer struct {
	_min time.Duration
	_max time.Duration
	Min  string `mapstructure:"min"`
	Max  string `mapstructure:"max"`
}

func NewDelayTransformer(params map[string]interface{}) (*DelayTransformer, error) {
	t := DelayTransformer{}
	err := DecodeAndTempl(params, &t, nil)
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
