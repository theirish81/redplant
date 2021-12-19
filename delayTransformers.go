package main

import (
	"math/rand"
	"time"
)

// DelayTransformer will slow the request down by a certain amount
type DelayTransformer struct {
	Min		time.Duration
	Max		time.Duration
}

// DelayConfig is the configuration for the DelayTransformer
type DelayConfig struct {
	Min	string	`mapstructure:"min"`
	Max	string	`mapstructure:"max"`
}

func NewDelayTransformer(params map[string]interface{}) (*DelayTransformer,error) {
	var cfg DelayConfig
	err := DecodeAndTempl(params,&cfg,nil)
	if err != nil {
		return nil,err
	}
	minObj,err := time.ParseDuration(cfg.Min)
	if err != nil {
		return nil,err
	}
	maxObj,err := time.ParseDuration(cfg.Max)
	if err != nil {
		return nil,err
	}
	return &DelayTransformer{minObj,maxObj},nil
}

func (t *DelayTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	timeRange := t.Max.Nanoseconds()-t.Min.Nanoseconds()
	nanos := t.Min.Nanoseconds()+rand.Int63n(timeRange)
	time.Sleep(time.Duration(nanos))
	return wrapper,nil
}