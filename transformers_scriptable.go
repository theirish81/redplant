package main

import (
	"context"
	"errors"
	"github.com/dop251/goja"
	"net/http"
	"os"
)

// ScriptableTransformer is a transformer that uses a JavaScript script
type ScriptableTransformer struct {
	Script         string
	Path           string
	_script        string
	ExpandRequest  bool
	ExpandResponse bool
	ActivateOnTags []string
	log            *STLogHelper
}

// Transform will perform the transformation. The script must return true if the scripts wants the request to move
// forward
func (t *ScriptableTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	runtime := goja.New()
	err := runtime.Set("wrapper", wrapper)
	if err != nil {
		t.log.LogErr("error while creating setting wrapper in runtime", err, wrapper, t.log.Error)
		return wrapper, err
	}
	// run the script
	val, err := runtime.RunString(t._script)
	// if the script failed at running, we return the error
	if err != nil {
		t.log.LogErr("error while running script", err, wrapper, t.log.Error)
		return wrapper, err
	}
	// export the result to a boolean
	res, ok := val.Export().(bool)
	// if the boolean conversion failed, then we return the error
	if !ok {
		t.log.LogErr("script did not return a boolean", nil, wrapper, t.log.Error)
		return wrapper, errors.New("script did not return a boolean")
	}
	// if the script executed fine, and we got a boolean back, and the boolean is true, then we return positively
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

// NewScriptableTransformer is the constructor for ScriptableTransformer
func NewScriptableTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*ScriptableTransformer, error) {
	t := ScriptableTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(context.Background(), params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	if t.Script != "" {
		t._script = t.Script
		return &t, nil
	}
	if t.Path != "" {
		data, err := os.ReadFile(t.Path)
		if err != nil {
			return nil, err
		}
		t._script = string(data)
		return &t, nil
	}
	return nil, errors.New("scriptable_transformer_no_config")

}
