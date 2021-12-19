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

func DecodeAndTempl(data map[string]interface{}, target interface{}, scope interface{}) error {
	err := mapstructure.Decode(data, target)
	if err != nil {
		return err
	}
	val := reflect.ValueOf(target).Elem()
	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Type().String() == "string" {
			if val.Field(i).CanSet() {
				parsed, err := Templ(val.Field(i).String(), scope)
				if err != nil {
					log.Warn("Error while compiling template", err, map[string]interface{}{"template": val.Field(i).String()})
				}
				val.Field(i).Set(reflect.ValueOf(parsed))
			}
		}

	}
	return nil
}
