package main

import (
	"strings"
)

type RequestUrlTransformer struct {

	OldPrefix string	`yaml:"oldPrefix" mapstructure:"oldPrefix"`

	NewPrefix string	`yaml:"newPrefix" mapstructure:"newPrefix"`

}

func (t *RequestUrlTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	path := wrapper.Request.URL.Path
	if strings.HasPrefix(path,t.OldPrefix) {
		wrapper.Request.URL.Path = strings.Replace(path,t.OldPrefix,t.NewPrefix,1)
	}
	return wrapper,nil
}

func NewRequestUrlTransformerFromParams(params map[string]interface{}) (*RequestUrlTransformer,error) {
	var transformer RequestUrlTransformer
	err := DecodeAndTempl(params,&transformer,nil)
	return &transformer,err
}