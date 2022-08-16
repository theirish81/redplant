package main

import (
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"net/http"
	"strings"
)

type RequestOpenAPISchemaTransformer struct {
	ActivateOnTags []string
}

func NewRequestOpenAPIValidatorTransformer(activateOnTags []string) (*RequestOpenAPISchemaTransformer, error) {
	return &RequestOpenAPISchemaTransformer{ActivateOnTags: activateOnTags}, nil
}

func (t *RequestOpenAPISchemaTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	route, params, err := (*wrapper.Rule.oaRouter).FindRoute(wrapper.Request)
	if err != nil {
		return wrapper, err
	}
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    wrapper.Request,
		PathParams: params,
		Route:      route,
	}
	err = openapi3filter.ValidateRequest(wrapper.Request.Context(), requestValidationInput)
	if err != nil {
		err = errors.New("validation_error: " + err.Error())
	}
	return wrapper, err
}

func (t *RequestOpenAPISchemaTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *RequestOpenAPISchemaTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *RequestOpenAPISchemaTransformer) ErrorMatches(err error) bool {
	return strings.HasPrefix(err.Error(), "validation_error")
}

func (t *RequestOpenAPISchemaTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(400)
}

func (t *RequestOpenAPISchemaTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}
