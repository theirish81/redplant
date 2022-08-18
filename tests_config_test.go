package main

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	config = LoadConfig("etc/config.yaml")
	config.Init()
	if len(config.Variables) == 0 {
		t.Error("Variables not parsed correctly")
	}
	if len(config.Before.Request.Transformers) == 0 ||
		len(config.Before.Response.Transformers) == 0 {
		t.Error("Transformers not parsed correctly")
	}
	if len(config.Before.Request.Sidecars) == 0 ||
		len(config.Before.Response.Sidecars) == 0 {
		t.Error("Sidecars not parsed correctly")
	}
	for _, v := range config.Rules {
		for _, v2 := range v {
			if len(v2.Request._transformers.transformers) <= 1 {
				t.Error("Initialization of request transformers failed")
			}
			if len(v2.Request._sidecars.sidecars) < 1 {
				t.Error("Initialization of request sidecars failed")
			}
			if len(v2.Response._transformers.transformers) < 1 {
				t.Error("Initialization of response transformers failed")
			}
			if len(v2.Response._sidecars.sidecars) < 1 {
				t.Error("Initialization of response transformers failed")
			}
		}
	}

}

func TestExtractPattern(t *testing.T) {
	method, path := extractPattern("[get] /bananas")
	if method != "get" || path != "/bananas" {
		t.Error("pattern extraction failed")
	}
	method, path = extractPattern("/bananas")
	if method != "" || path != "/bananas" {
		t.Error("pattern extraction failed in absence of a method")
	}
}

func TestDBPatterns(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	config = LoadConfig("etc/config.yaml")
	config.Init()
	config.Rules["localhost:9001"] = map[string]*Rule{"/db": {Origin: "mysql://foo"}}
	config.Init()
	if config.Rules["localhost:9001"]["/db"].db == nil {
		t.Error("could not initialise DB")
	}
}

func TestLoadLoggerConfig(t *testing.T) {
	logging := "etc/logging.yaml"
	lc, _ := LoadLoggerConfig(&logging)
	if lc.Level == "" || lc.Format == "" {
		t.Error("could not load logger config properly")
	}

	lc, _ = LoadLoggerConfig(nil)
	if lc.Level != "INFO" {
		t.Error("nil logging config does not produce the defaults")
	}
}
