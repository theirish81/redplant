package main

import (
	"errors"
	"regexp"
)
// BarrageRequestTransformer is a transformer that will stop the request if certain preconditions are met
type BarrageRequestTransformer struct {
	// HeaderNameRegexp is a regular expression for a forbidden header name
	HeaderNameRegexp		string
	// HeaderValueRegexp is a regular expression for a forbidden header name
	HeaderValueRegexp		string
	// HeaderRegexp is a regular expression for a forbidden full header as in name:value
	HeaderRegexp			string
	// BodyRegexp is a regular expression for a forbidden body
	BodyRegexp				string
}

// NewBarrageRequestTransformer is the constructor for BarrageRequestTransformer
func NewBarrageRequestTransformer(params map[string]interface{}) (*BarrageRequestTransformer,error) {
	var t BarrageRequestTransformer
	err := DecodeAndTempl(params, &t,nil)
	return &t,err
}

// Transform will block the request if the preconditions are not met
func (t *BarrageRequestTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	for k,v := range wrapper.Request.Header {
		if t.HeaderRegexp != "" {
			if matched,_ := regexp.MatchString(t.HeaderRegexp,k+":"+v[0]); matched {
				return wrapper,errors.New("barraged")
			}
		}
		if t.HeaderNameRegexp != "" {
			if matched,_ := regexp.MatchString(t.HeaderNameRegexp,k); matched {
				return wrapper,errors.New("barraged")
			}
		}
		if t.HeaderValueRegexp != "" {
			if matched,_ := regexp.MatchString(t.HeaderValueRegexp,v[0]); matched {
				return wrapper,errors.New("barraged")
			}
		}
	}
	if t.BodyRegexp != "" {
		if matched,_ := regexp.Match(t.BodyRegexp,wrapper.RequestBody); matched {
			return wrapper,errors.New("barraged")
		}
	}
	return wrapper,nil
}