package main

import (
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"net/http"
	"strings"
)

type RequestOpenAPISchemaTransformer struct {
	ActivateOnTags []string
	log            *STLogHelper
}

func NewRequestOpenAPIValidatorTransformer(activateOnTags []string, logCfg *STLogConfig) (*RequestOpenAPISchemaTransformer, error) {
	t := RequestOpenAPISchemaTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	t.log.PrometheusRegisterCounter("openapi_validation_failed")
	return &t, nil
}

func (t *RequestOpenAPISchemaTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("triggering request OpenAPI schema transformer", wrapper, t.log.Debug)
	route, params, err := (*wrapper.Rule.oaRouter).FindRoute(wrapper.Request.Request)
	if err != nil {
		t.log.LogErr("problem finding route in OpenAPI", err, wrapper, t.log.Warn)
		return wrapper, err
	}
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    wrapper.Request.Request,
		PathParams: params,
		Route:      route,
	}
	err = openapi3filter.ValidateRequest(wrapper.Request.Context(), requestValidationInput)
	if err != nil {
		t.log.PrometheusCounterInc("openapi_validation_failed")
		t.log.LogErr("validation error in OpenAPI", err, wrapper, t.log.Warn)
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
