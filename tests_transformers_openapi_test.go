package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"testing"
)

func TestRequestOpenAPISchemaTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	r, _ := http.NewRequest("GET", "https://petstore3.swagger.io/api/v3/pet/1?api_key=foobar", nil)
	req := NewAPIRequest(r)
	wrapper := APIWrapper{Request: req}
	wrapper.Rule = &Rule{}
	rules := OpenAPI2Rules(map[string]*OpenAPIConfig{"localhost:9001": {File: "etc/openapi.yaml"}})
	rule := rules["localhost:9001"]["[get] ^/api/v3/pet/.*$"]
	wrapper.Rule = rule

	transformer, _ := NewRequestOpenAPIValidatorTransformer([]string{}, nil)
	if _, err := transformer.Transform(&wrapper); err != nil {
		t.Error("something went wrong while validating a fully valid request")
	}

	r, _ = http.NewRequest("GET", "https://petstore3.swagger.io/api/v3/pet/a?api_key=foobar", nil)
	req = NewAPIRequest(r)
	wrapper.Request = req
	if _, err := transformer.Transform(&wrapper); err == nil {
		t.Error("non valid request should return an error")
	}
}
