package main

import (
	"bytes"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"reflect"
)
import "text/template"

// Templ evaluates a template against a scope. If the provided scope is nil, a scope will get created containing
// a "Variables" object, directed from Config
func Templ(data string, scope any) (string, error) {
	templ := template.New("Templ")
	templ, err := templ.Parse(data)
	if err != nil {
		return "", err
	}
	parsed := bytes.NewBufferString("")
	if scope == nil {
		err = templ.Execute(parsed, map[string]any{"Variables": config.Variables})
	} else {
		err = templ.Execute(parsed, scope)
	}
	return parsed.String(), err
}

// DecodeAndTempl will decode a map[string]any into a target data structure. Then it will evaluate all the
// templates found in the decoded structure, against a provided scope (see Templ). Evaluation will not trigger for
// any field listed in the excludeVal array
func DecodeAndTempl(data map[string]any, target any, scope any, excludeEval []string) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	templFieldSet(target, scope, excludeEval)
	return nil
}

// templFieldSet will recursively evaluate templates for a set of fields, against a provided scope (see Templ).
// Any field with a name that is present in the excludedVal array will not be evaluated
func templFieldSet(target any, scope any, excludeEval []string) {
	objectType := reflect.ValueOf(target).Type().String()
	switch objectType {
	// If it's a map of strings...
	case "*map[string]string":
		t2 := target.(*map[string]string)
		// ... we iterate on each element and evaluate
		for k, v := range *t2 {
			(*t2)[k], _ = Templ(v, scope)
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
					parsed, err := Templ(val.Field(i).String(), scope)
					if err != nil {
						log.Warn("Error while compiling template", err, logrus.Fields{"template": val.Field(i).String()})
					}
					// Setting the value
					val.Field(i).Set(reflect.ValueOf(parsed))
				}
				// If it's a map of strings, then we go in recursively
				if objectType == "map[string]string" {
					mp := val.Field(i).Interface().(map[string]string)
					templFieldSet(&mp, scope, excludeEval)
				}
			}
		}
	}

}
