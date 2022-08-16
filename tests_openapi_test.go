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
	if _, ok := rules["localhost"]["[get] ^/v1/pets/.*$"]; !ok {
		t.Error("could not find path")
	}
}
