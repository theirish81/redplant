package main

import (
	"net/http"
)

// IRequestTransformer is the interface for all request transformers
type IRequestTransformer interface {
	Transform(wrapper *APIWrapper) (*APIWrapper, error)
	ShouldExpandRequest() bool
	ErrorMatches(err error) bool
	HandleError(writer *http.ResponseWriter)
	IsActive(wrapper *APIWrapper) bool
}

// RequestTransformers is the store for all request transformers, associated to a given route
type RequestTransformers struct {
	transformers []IRequestTransformer
}

// ShouldExpandRequest will return true if at least one transformer requires the request to be expanded
func (t *RequestTransformers) ShouldExpandRequest() bool {
	for _, tx := range t.transformers {
		if tx.ShouldExpandRequest() {
			return true
		}
	}
	return false
}

// Transform will process all request transformers for the given wrapper
func (t *RequestTransformers) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for _, transformer := range t.transformers {
		if transformer.IsActive(wrapper) {
			if _, err := transformer.Transform(wrapper); err != nil {
				return wrapper, err
			}
		}
	}
	return wrapper, nil
}

// Push will append a transformer to the transformers
func (t *RequestTransformers) Push(transformer IRequestTransformer) {
	t.transformers = append(t.transformers, transformer)
}

// FindErrorHandler will find a transformer that has the capability of handling a certain error.
// if the transformer is not found, then nil is returned
func (t *RequestTransformers) FindErrorHandler(err error) IRequestTransformer {
	for _, tx := range t.transformers {
		if tx.ErrorMatches(err) {
			return tx
		}
	}
	return nil
}

// HandleError will try to delegate a error to a transformer capable of handling it. If a suitable transformer is found
// then true is returned. False otherwise.
func (t *RequestTransformers) HandleError(err error, writer *http.ResponseWriter) bool {
	handler := t.FindErrorHandler(err)
	if handler == nil {
		return false
	} else {
		handler.HandleError(writer)
		return true
	}
}

// NewRequestTransformers initializes all request transformers, based on their configurations
func NewRequestTransformers(transformers *[]TransformerConfig) (*RequestTransformers, error) {
	res := RequestTransformers{}
	var err error
	for _, t := range *transformers {
		var transformer IRequestTransformer
		switch t.Id {
		case "url":
			transformer, err = NewRequestUrlTransformerFromParams(t.ActivateOnTags, t.Logging, t.Params)
		case "headers":
			transformer, err = NewRequestHeadersTransformerFromParams(t.ActivateOnTags, t.Logging, t.Params)
		case "basic-auth":
			transformer, err = NewBasicAuthTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "jwt-auth":
			transformer, err = NewJWTAuthTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "jwt-sign":
			transformer, err = NewJWTSignTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "cookie-to-token-auth":
			transformer, err = NewCookieToTokenTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "scriptable":
			transformer, err = NewScriptableTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "delay":
			transformer, err = NewDelayTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "barrage":
			transformer, err = NewBarrageRequestTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "tag":
			transformer, err = NewTagTransformer(t.Logging, t.Params)
		case "rate-limiter":
			transformer, err = NewRequestRateLimiterTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "parser":
			transformer, err = NewRequestParserTransformer(t.ActivateOnTags, t.Logging)
		case "openapi-validator":
			transformer, err = NewRequestOpenAPIValidatorTransformer(t.ActivateOnTags, t.Logging)
		case "payload":
			transformer, err = NewRequestPayloadTransformer(t.ActivateOnTags, t.Logging, t.Params)
		}
		if transformer != nil && err == nil {
			res.Push(transformer)
		}
		if err != nil {
			break
		}
	}
	return &res, err
}

// IResponseTransformer is the interface for all response transformers
type IResponseTransformer interface {
	Transform(wrapper *APIWrapper) (*APIWrapper, error)
	ShouldExpandRequest() bool
	ShouldExpandResponse() bool
	ErrorMatches(err error) bool
	HandleError(writer *http.ResponseWriter)
	IsActive(wrapper *APIWrapper) bool
}

// ResponseTransformers is the store for the response transformers for a given route
type ResponseTransformers struct {
	transformers []IResponseTransformer
}

// ShouldExpandRequest will return true if at least one transformer needs the request to be expanded
func (t *ResponseTransformers) ShouldExpandRequest() bool {
	for _, tx := range t.transformers {
		if tx.ShouldExpandRequest() {
			return true
		}
	}
	return false
}

// ShouldExpandResponse will return true if at least one transformer needs the response to be expanded
func (t *ResponseTransformers) ShouldExpandResponse() bool {
	for _, tx := range t.transformers {
		if tx.ShouldExpandResponse() {
			return true
		}
	}
	return false
}

// Transform will process the whole response transformation pipeline
func (t *ResponseTransformers) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	for _, transformer := range t.transformers {
		if transformer.IsActive(wrapper) {
			if wrapper, err := transformer.Transform(wrapper); err != nil {
				return wrapper, err
			}
		}

	}
	return wrapper, nil
}

// Push will append a transformer to the response transformers
func (t *ResponseTransformers) Push(transformer IResponseTransformer) {
	t.transformers = append(t.transformers, transformer)
}

// FindErrorHandler will find a transformer that has the capability of handling a certain error.
// if the transformer is not found, then nil is returned
func (t *ResponseTransformers) FindErrorHandler(err error) IResponseTransformer {
	for _, tx := range t.transformers {
		if tx.ErrorMatches(err) {
			return tx
		}
	}
	return nil
}

// HandleError will try to delegate a error to a transformer capable of handling it. If a suitable transformer is found
// then true is returned. False otherwise.
func (t *ResponseTransformers) HandleError(err error, writer *http.ResponseWriter) bool {
	handler := t.FindErrorHandler(err)
	if handler == nil {
		return false
	} else {
		handler.HandleError(writer)
		return true
	}
}

// NewResponseTransformers initializes all response transformers, based on their configurations
func NewResponseTransformers(transformers *[]TransformerConfig) (*ResponseTransformers, error) {
	res := ResponseTransformers{}
	var err error
	for _, t := range *transformers {
		var transformer IResponseTransformer
		switch t.Id {
		case "headers":
			transformer, err = NewResponseHeadersTransformerFromParams(t.ActivateOnTags, t.Logging, t.Params)
		case "scriptable":
			transformer, err = NewScriptableTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "delay":
			transformer, _ = NewDelayTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "barrage":
			transformer, err = NewBarrageResponseTransformer(t.ActivateOnTags, t.Logging, t.Params)
		case "parser":
			transformer, err = NewResponseParserTransformer(t.ActivateOnTags, t.Logging)
		case "tag":
			transformer, err = NewTagTransformer(t.Logging, t.Params)
		case "payload":
			transformer, err = NewResponsePayloadTransformer(t.ActivateOnTags, t.Logging, t.Params)
		}
		if transformer != nil && err == nil {
			res.Push(transformer)
		}
		if err != nil {
			break
		}
	}
	return &res, nil
}
