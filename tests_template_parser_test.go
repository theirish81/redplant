package main

import (
	"context"
	"testing"
)

func TestTempl(t *testing.T) {
	template = NewRPTemplate()
	scope := map[string]any{"foo": "bar"}
	data, _ := template.Templ(context.Background(), "yay ${foo}", &scope)
	if data != "yay bar" {
		t.Error("Template with provided scope is not working")
	}

	data, err := template.Templ(context.Background(), "foo ${foo()}", &scope)
	if err == nil {
		t.Error("Broken template should error out")
	}

	config = Config{}
	config.Variables = StringMap{"john": "doe"}
	data, _ = template.Templ(context.Background(), "yay ${Variables.john}", nil)
	if data != "yay doe" {
		t.Error("Templating with global variables not working")
	}
}

func TestDecodeAndTempl(t *testing.T) {
	template = NewRPTemplate()
	config = Config{}
	config.Variables = StringMap{"john": "doe"}
	data := map[string]any{"Data": "${Variables.john}"}
	type Foo struct {
		Data string
	}
	var foo Foo
	_ = template.DecodeAndTempl(context.Background(), data, &foo, nil, []string{})
	if foo.Data != "doe" {
		t.Error("DecodeAndTempl doesn't seem to work correctly")
	}

	var foo2 StringMap
	_ = template.DecodeAndTempl(context.Background(), data, &foo2, nil, []string{})
	if foo.Data != "doe" {
		t.Error("DecodeAndTempl doesn't seem to work correctly")
	}
	var foo3 map[string]string
	_ = template.DecodeAndTempl(context.Background(), data, &foo3, nil, []string{})
	if foo.Data != "doe" {
		t.Error("DecodeAndTempl doesn't seem to work correctly")
	}
}
