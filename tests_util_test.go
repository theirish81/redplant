package main

import (
	"reflect"
	"testing"
)

func TestConvertMaps(t *testing.T) {
	data := map[interface{}]interface{}{}
	data["foo"] = map[interface{}]interface{}{"foo": map[interface{}]interface{}{"foo": 22}}
	d2 := convertMaps(data)
	if reflect.ValueOf(d2).Type().String() != "map[string]interface {}" {
		t.Error("First level conversion failed")
	}
	if reflect.ValueOf(d2.(map[string]interface{})["foo"]).Type().String() != "map[string]interface {}" {
		t.Error("Second level conversion failed")
	}
}

func TestIsString(t *testing.T) {
	if !isString("foo") {
		t.Error("mis-identification of string")
	}
	if isString(2) {
		t.Error("mis-identification of int")
	}
	if isString(true) {
		t.Error("mis-identification of bool")
	}
	if isString(nil) {
		t.Error("mis-identification of nil")
	}
}

func TestParseBasicAuth(t *testing.T) {
	if un, pw, ok := parseBasicAuth("Basic Zm9vOmJhcg=="); !ok || un != "foo" || pw != "bar" {
		t.Error("error parsing basic auth")
	}
	if _, _, ok := parseBasicAuth("Basic foobar"); ok {
		t.Error("found auth where there's none")
	}
	if _, _, ok := parseBasicAuth("Basic foobar"); ok {
		t.Error("found auth where there's none")
	}
	if _, _, ok := parseBasicAuth("Basic "); ok {
		t.Error("found auth where there's none")
	}
	if _, _, ok := parseBasicAuth("foo"); ok {
		t.Error("found auth where there's none")
	}
}
