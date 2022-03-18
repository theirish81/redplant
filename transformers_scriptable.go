package main

import (
	"errors"
	"github.com/dop251/goja"
	"io/ioutil"
	"net/http"
)

// ScriptableTransformer is a transformer that uses a JavaScript script
type ScriptableTransformer struct {
	Script         string
	Path           string
	_script        string
	ExpandRequest  bool
	ExpandResponse bool
	ActivateOnTags []string
}

// Transform will perform the transformation. The script must return true if the scripts wants the request to move
// forward
func (t *ScriptableTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	runtime := goja.New()
	err := runtime.Set("wrapper", wrapper)
	if err != nil {
		return wrapper, err
	}
	// run the script
	val, err := runtime.RunString(t._script)
	// if the script failed at running, we return the error
	if err != nil {
		return wrapper, err
	}
	// export the result to a boolean
	res, ok := val.Export().(bool)
	// if the boolean conversion failed, then we return the error
	if !ok {
		return wrapper, errors.New("scriptable transformer: wrong return type in script")
	}
	// if the script executed fine, and we got a boolean back, and the boolean is true, then we return positvely
	if res {
		return wrapper, nil
	}
	// in all other scenarios, the request is rejected
	return wrapper, errors.New("script_rejected")
}

func (t *ScriptableTransformer) ErrorMatches(err error) bool {
	return err.Error() == "script_rejected"
}

func (t *ScriptableTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(403)
}

func (t *ScriptableTransformer) ShouldExpandRequest() bool {
	return t.ExpandRequest
}

func (t *ScriptableTransformer) ShouldExpandResponse() bool {
	return t.ExpandResponse
}

func (t *ScriptableTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

func NewScriptableTransformer(activateOnTags []string, params map[string]interface{}) (*ScriptableTransformer, error) {
	t := ScriptableTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	if t.Script != "" {
		t._script = t.Script
		return &t, nil
	}
	if t.Path != "" {
		data, err := ioutil.ReadFile(t.Path)
		if err != nil {
			return nil, err
		}
		t._script = string(data)
		return &t, nil
	}
	return nil, errors.New("scriptable_transformer_no_config")

}
