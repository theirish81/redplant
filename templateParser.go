package main

import (
	"bytes"
	"github.com/mitchellh/mapstructure"
	"reflect"
)
import "text/template"


func  Templ(data string, scope interface{}) (string,error) {
	templ := template.New("Templ")
	templ,err := templ.Parse(data)
	if err != nil {
		return "",err
	}
	parsed := bytes.NewBufferString("")
	if scope == nil {
		err = templ.Execute(parsed,map[string]interface{}{"Variables":config.Variables})
	} else {
		err = templ.Execute(parsed, scope)
	}
	return parsed.String(),err
}

func DecodeAndTempl(data map[string]interface{}, target interface{}, scope interface{}) error{
	mapstructure.Decode(data,target)
	val := reflect.ValueOf(target).Elem()
	for i:=0; i<val.NumField();i++{
		if val.Field(i).Type().String() == "string" {
				parsed,_ := Templ(val.Field(i).String(),scope)
				val.Field(i).Set(reflect.ValueOf(parsed))
		}

	}
	return nil
}