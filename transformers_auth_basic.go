package main

import (
	"errors"
	"github.com/tg123/go-htpasswd"
	"net/http"
)

// BasicAuthTransformer is a transformer that will block the request in case the credentials do not match the
// expectations.
// Username directly provided in the conf
// Password directly provided in the conf
// Htpasswd is the path to the Htpasswd file
// Proxy if set to true, the "proxy-authorization" will be used instead
// Retain if set to false, the credentials will be removed from the request
// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
// _htpasswd is the parsed password file
type BasicAuthTransformer struct {
	Username       string
	Password       string
	Htpasswd       string
	Proxy          bool
	Retain         bool
	ActivateOnTags []string
	_htpasswd      *htpasswd.File
	log            *STLogHelper
}

// Transform will throw an error if the request doesn't match the basic auth expectations
func (t *BasicAuthTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	t.log.Log("basic auth triggered", wrapper, t.log.Debug)
	// We first detect whether basic credentials are passed over, and we collect them
	username, password, ok := wrapper.Request.BasicAuth()
	// If we have a htpasswd file loaded, then we use that
	if ok && t._htpasswd != nil && t._htpasswd.Match(username, password) {
		t.log.Log("basic auth accepted", wrapper, t.log.Debug)
		wrapper.Username = username
		t.postAuthOperations(wrapper)
		return wrapper, nil
		// If we don't have the file, then we rely on provided username and password
	} else if ok && t.Username == username && t.Password == password {
		t.log.Log("basic auth accepted", wrapper, t.log.Debug)
		wrapper.Username = username
		t.postAuthOperations(wrapper)
		return wrapper, nil
	} else {
		t.log.Log("basic auth denied", wrapper, t.log.Debug)
		t.log.PrometheusCounterInc("basic_auth_denied")
		// If nothing works, then no_auth
		return nil, errors.New("no_auth")
	}
}

// obtainUsernameAndPassword will extract username and password from the request. If Proxy is set to false, then
// the "authorization" header will be used. If it's set to true, then "proxy-authorization" header will be used
func (t *BasicAuthTransformer) obtainUsernameAndPassword(wrapper *APIWrapper) (string, string, bool) {
	if t.Proxy {
		return parseBasicAuth(wrapper.Request.Header.Get("proxy-authorization"))
	} else {
		return wrapper.Request.BasicAuth()
	}
}

// postAuthOperations will delete the credentials from the request if Retain is set to false
func (t *BasicAuthTransformer) postAuthOperations(wrapper *APIWrapper) {
	// If we must not retain the credentials
	if !t.Retain {
		// If it's a proxy auth, we remove "proxy-authorization" header
		if t.Proxy {
			wrapper.Request.Header.Del("proxy-authorization")
		} else {
			// If it's a reverse-proxy auth, then we remove the "authorization" header
			wrapper.Request.Header.Del("authorization")
		}
	}
}

func (t *BasicAuthTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *BasicAuthTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *BasicAuthTransformer) ErrorMatches(err error) bool {
	return err.Error() == "no_auth"
}

func (t *BasicAuthTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(401)
}

func (t *BasicAuthTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

// NewBasicAuthTransformer creates a BasicAuthTransformer from params
func NewBasicAuthTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*BasicAuthTransformer, error) {
	t := BasicAuthTransformer{ActivateOnTags: activateOnTags, Proxy: false, Retain: true, log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(params, &t, nil, []string{})
	// if the path to a Htpasswd file is provided, then we parse it
	if t.Htpasswd != "" {
		t._htpasswd, err = htpasswd.New(t.Htpasswd, htpasswd.DefaultSystems, nil)
		if err != nil {
			return nil, err
		}
	}
	t.log.PrometheusRegisterCounter("basic_auth_denied")
	return &t, err
}
