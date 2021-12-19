package main

// RequestHeaderTransformer transforms the request header by setting or removing headers
type RequestHeaderTransformer struct {
	Set		map[string]string	`mapstructure:"set"`
	Remove []string				`mapstructure:"remove"`
}

func (t *RequestHeaderTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	for hk,hv := range t.Set {
		hv,err := wrapper.Templ(hv)
		if err != nil {
			return wrapper,err
		}
		wrapper.Request.Header.Set(hk, hv)
	}
	for _,rem := range t.Remove {
		wrapper.Request.Header.Del(rem)
	}
	return wrapper,nil
}

func NewRequestHeadersTransformerFromParams(params map[string]interface{}) (*RequestHeaderTransformer,error) {
	var t RequestHeaderTransformer
	err := DecodeAndTempl(params, &t,nil)
	return &t,err
}

type ResponseHeaderTransformer struct {

	Set		map[string]string	`mapstructure:"set"`
	Remove []string				`mapstructure:"remove"`
}

func NewResponseHeadersTransformerFromParams(params map[string]interface{}) (*ResponseHeaderTransformer,error) {
	var t ResponseHeaderTransformer
	err := DecodeAndTempl(params, &t,nil)
	return &t, err
}

func (t *ResponseHeaderTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	for hk,hv := range t.Set {
		wrapper.Response.Header.Set(hk, hv)
	}
	for _,rem := range t.Remove {
		wrapper.Request.Header.Del(rem)
	}
	return wrapper,nil
}