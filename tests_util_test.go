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
