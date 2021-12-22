package main

import (
	"bytes"
	"github.com/mitchellh/mapstructure"
	"reflect"
)
import "text/template"

func Templ(data string, scope interface{}) (string, error) {
	templ := template.New("Templ")
	templ, err := templ.Parse(data)
	if err != nil {
		return "", err
	}
	parsed := bytes.NewBufferString("")
	if scope == nil {
		err = templ.Execute(parsed, map[string]interface{}{"Variables": config.Variables})
	} else {
		err = templ.Execute(parsed, scope)
	}
	return parsed.String(), err
}

func DecodeAndTempl(data map[string]interface{}, target interface{}, scope interface{}, excludeEval []string) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	templFieldSet(target, scope, excludeEval)
	return nil
}

func templFieldSet(target interface{}, scope interface{}, excludeEval []string) {
	objectType := reflect.ValueOf(target).Type().String()
	switch objectType {
	case "*map[string]string":
		t2 := target.(*map[string]string)
		for k, v := range *t2 {
			(*t2)[k], _ = Templ(v, scope)
		}
	default:
		val := reflect.ValueOf(target).Elem()
		for i := 0; i < val.NumField(); i++ {
			if !stringInArray(getFieldName(val, i), excludeEval) && val.Field(i).CanSet() {
				objectType := val.Field(i).Type().String()
				if objectType == "string" {
					parsed, err := Templ(val.Field(i).String(), scope)
					if err != nil {
						log.Warn("Error while compiling template", err, map[string]interface{}{"template": val.Field(i).String()})
					}
					val.Field(i).Set(reflect.ValueOf(parsed))
				}
				if objectType == "map[string]string" {
					mp := val.Field(i).Interface().(map[string]string)
					templFieldSet(&mp, scope, excludeEval)
				}
			}
		}
	}

}
