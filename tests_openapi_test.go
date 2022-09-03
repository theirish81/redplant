package main

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestOpenAPI2Rules(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	cfg := &OpenAPIConfig{File: "etc/openapi.yaml"}
	rules := OpenAPI2Rules(map[string]*OpenAPIConfig{"localhost": cfg, "127.0.0.1": cfg})
	if _, ok := rules["localhost"]; !ok {
		t.Error("did not properly map domain")
	}
	if _, ok := rules["127.0.0.1"]; !ok {
		t.Error("did not properly map domain")
	}
	if _, ok := rules["localhost"]["[get] /api/v3/pet/{petId}"]; !ok {
		t.Error("could not find path")
	}
}

func TestMergeRules(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	config = LoadConfig("etc/config.yaml")
	config.OpenAPI = map[string]*OpenAPIConfig{"localhost:9001": {File: "etc/openapi.yaml"}}
	config.Init()
	if _, ok := config.Rules["localhost:9001"]["[get] /api/v3/pet/{petId}"]; !ok {
		t.Error("merge rules failed")
	}
}
