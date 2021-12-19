package main

import (
	"errors"
	"github.com/dop251/goja"
	"io/ioutil"
)

type ScriptableTransformer struct {
	script	string
}

func (t *ScriptableTransformer) Transform(wrapper *APIWrapper) (*APIWrapper,error) {
	runtime := goja.New()
	err := runtime.Set("wrapper",wrapper)
	if err != nil {
		return wrapper,err
	}
	// run the script
	val,err := runtime.RunString(t.script)
	// if the script failed at running, we return the error
	if err != nil {
		return wrapper,err
	}
	// export the result to a boolean
	res, ok := val.Export().(bool)
	// if the boolean conversion failed, then we return the error
	if !ok {
		return wrapper,errors.New("scriptable transformer: wrong return type in script")
	}
	// if the script executed fine, and we got a boolean back, and the boolean is true, then we return positvely
	if res {
		return wrapper, nil
	}
	// in all other scenarios, the request is rejected
	return wrapper,errors.New("rejected")
}

func NewScriptableTransformer(params map[string]interface{}) (*ScriptableTransformer,error) {
	if script,ok := params["script"]; ok  {
		return &ScriptableTransformer{script: script.(string)},nil
	}
	if path,ok := params["path"]; ok {
		data,err := ioutil.ReadFile(path.(string))
		return &ScriptableTransformer{script: string(data)},err
	}
	return nil,errors.New("scriptable_transformer_no_config")

}