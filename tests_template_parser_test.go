package main

import (
	"testing"
)

func TestTempl(t *testing.T) {
	scope := map[string]any{"foo": "bar"}
	data, _ := Templ("yay {{.foo}}", &scope)
	if data != "yay bar" {
		t.Error("Template with provided scope is not working")
	}

	data, err := Templ("foo {{.foo()}}", &scope)
	if err == nil {
		t.Error("Broken template should error out")
	}

	config = Config{}
	config.Variables = map[string]string{"john": "doe"}
	data, _ = Templ("yay {{.Variables.john}}", nil)
	if data != "yay doe" {
		t.Error("Templating with global variables not working")
	}
}

func TestDecodeAndTempl(t *testing.T) {
	config = Config{}
	config.Variables = map[string]string{"john": "doe"}
	data := map[string]any{"Data": "{{.Variables.john}}"}
	type Foo struct {
		Data string
	}
	var foo Foo
	_ = DecodeAndTempl(data, &foo, nil, []string{})
	if foo.Data != "doe" {
		t.Error("DecodeAndTempl doesn't seem to work correctly")
	}
}
