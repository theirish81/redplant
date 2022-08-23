package main

import (
	"context"
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/theirish81/gowalker"
	"net/http"
	"reflect"
)

type RPTemplate struct {
	functions *gowalker.Functions
}

func NewRPTemplate() RPTemplate {
	t := RPTemplate{functions: gowalker.NewFunctions()}
	t.functions.Add("GetHeader", func(context context.Context, data any, params ...string) (any, error) {
		switch d := data.(type) {
		case http.Request:
			return d.Header.Get(params[0]), nil
		case http.Response:
			return d.Header.Get(params[0]), nil
		default:
			return "", errors.New("cannot invoke GetHeader function against this type")
		}
	})
	return t
}

// Templ evaluates a template against a scope. If the provided scope is nil, a scope will get created containing
// a "Variables" object, directed from Config
func (t *RPTemplate) Templ(data string, scope any) (string, error) {
	if scope == nil {
		return gowalker.Render(context.TODO(), data, AnyMap{"Variables": config.Variables}, t.functions)
	} else {
		return gowalker.Render(context.TODO(), data, scope, t.functions)
	}
}

// DecodeAndTempl will decode a map[string]any into a target data structure. Then it will evaluate all the
// templates found in the decoded structure, against a provided scope (see Templ). Evaluation will not trigger for
// any field listed in the excludeVal array
func (t *RPTemplate) DecodeAndTempl(data map[string]any, target any, scope any, excludeEval []string) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	t.templFieldSet(target, scope, excludeEval)
	return nil
}

// templFieldSet will recursively evaluate templates for a set of fields, against a provided scope (see Templ).
// Any field with a name that is present in the excludedVal array will not be evaluated
func (t *RPTemplate) templFieldSet(target any, scope any, excludeEval []string) {
	objectType := reflect.ValueOf(target).Type().String()
	switch objectType {
	// If it's a map of strings...
	case "*main.StringMap":
		t2 := target.(*StringMap)
		// ... we iterate on each element and evaluate
		for k, v := range *t2 {
			(*t2)[k], _ = t.Templ(v, scope)
		}
	case "*map[string]string":
		t2 := target.(*map[string]string)
		// ... we iterate on each element and evaluate
		for k, v := range *t2 {
			(*t2)[k], _ = t.Templ(v, scope)
		}
	default:
		// If it's any other object
		val := reflect.ValueOf(target).Elem()
		// For each field...
		for i := 0; i < val.NumField(); i++ {
			// if the key is not among the excluded, and it can be set
			if !stringInArray(getFieldName(val, i), excludeEval) && val.Field(i).CanSet() {
				objectType := val.Field(i).Type().String()
				// If it's a string, then we can proceed
				if objectType == "string" {
					// Evaluating the template
					parsed, err := t.Templ(val.Field(i).String(), scope)
					if err != nil {
						log.Warn("Error while compiling template", err, AnyMap{"template": val.Field(i).String()})
					}
					// Setting the value
					val.Field(i).Set(reflect.ValueOf(parsed))
				}
				// If it's a map of strings, then we go in recursively
				if objectType == "map[string]string" {
					mp := val.Field(i).Interface().(map[string]string)
					t.templFieldSet(&mp, scope, excludeEval)
				}
			}
		}
	}

}
