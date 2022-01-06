package main

import (
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/tg123/go-htpasswd"
	"io/ioutil"
	"net/http"
	"strings"
)

// BasicAuthTransformer is a transformer that will block the request in case the credentials do not match the
// expectations
type BasicAuthTransformer struct {

	// Username directly provided in the conf
	Username string `mapstructure:"username"`
	// Password directly provided in the conf
	Password string `mapstructure:"password"`
	// Path to the Htpasswd file
	Htpasswd string `mapstructure:"htpasswd"`
	// If Proxy is set to true, the "proxy-authorization" will be used instead
	Proxy bool
	// If Retain is set to false, the credentials will be removed from the request
	Retain bool
	// _htpasswd is the parsed password file
	_htpasswd *htpasswd.File
	// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
	ActivateOnTags []string
}

// Transform will throw an error if the request doesn't match the basic auth expectations
func (t *BasicAuthTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	// We first detect whether basic credentials are passed over, and we collect them
	username, password, ok := wrapper.Request.BasicAuth()
	// If we have a htpasswd file loaded, then we use that
	if ok && t._htpasswd != nil && t._htpasswd.Match(username, password) {
		wrapper.Username = username
		t.postAuthOperations(wrapper)
		return wrapper, nil
		// If we don't have the file, then we rely on provided username and password
	} else if ok && t.Username == username && t.Password == password {
		wrapper.Username = username
		t.postAuthOperations(wrapper)
		return wrapper, nil
	} else {
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
func NewBasicAuthTransformer(activateOnTags []string, params map[string]interface{}) (*BasicAuthTransformer, error) {
	t := BasicAuthTransformer{ActivateOnTags: activateOnTags, Proxy: false, Retain: true}
	err := DecodeAndTempl(params, &t, nil, []string{})
	// if the path to a Htpasswd file is provided, then we parse it
	if t.Htpasswd != "" {
		t._htpasswd, err = htpasswd.New(t.Htpasswd, htpasswd.DefaultSystems, nil)
		if err != nil {
			return nil, err
		}
	}
	return &t, err
}

// JWTAuthTransformer will block any request without a Bearer token or a token whose signature cannot be verified.
// In addition, it will store claims in the wrapper.
type JWTAuthTransformer struct {
	// _publicKey is the loaded and parsed public key
	_publicKey *rsa.PublicKey
	// _key is the key provided in the conf, in the form of bytes
	_key []byte
	// Pem is the path to a PEM certificate
	Pem string `mapstructure:"pem"`
	// Key is the key provided in the conf, in the form of a string
	Key string `mapstructure:"key"`
	// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
	ActivateOnTags []string
}

func (t *JWTAuthTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *JWTAuthTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *JWTAuthTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

// Transform will block any request without a Bearer token or a token whose signature cannot be verified.
// In addition, it will store claims in the wrapper.
func (t *JWTAuthTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	header := wrapper.Request.Header.Get("authorization")

	// if no Bearer prefix, then we have no auth
	if !strings.HasPrefix(header, "Bearer") {
		return nil, errors.New("no_auth")
	}

	// Extract the token from the header
	token := strings.TrimSpace(header[7:])
	claims := jwt.MapClaims{}
	// Parsing with signature check and claims extraction
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		// If we have a public _key, we'll use that
		if t._publicKey != nil {
			return t._publicKey, nil
		}
		// Otherwise, we'll just use the _key
		return t._key, nil
	})
	// Storing claims in the wrapper
	wrapper.Claims = &claims
	return wrapper, err
}

func (t *JWTAuthTransformer) ErrorMatches(err error) bool {
	return err.Error() == "signature is invalid"
}

func (t *JWTAuthTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(401)
}

// NewJWTAuthTransformer creates a new JWTAuthTransformer from params
func NewJWTAuthTransformer(activateOnTags []string, params map[string]interface{}) (*JWTAuthTransformer, error) {
	t := JWTAuthTransformer{ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	// If there's a "pem" parameter, then we treat that as a path to the certificate
	if t.Pem != "" {
		cert, err := ioutil.ReadFile(t.Pem)
		if err != nil {
			return nil, err
		}
		key, err := jwt.ParseRSAPublicKeyFromPEM(cert)
		t._publicKey = key
		return &t, err
	}
	// If there's a "_key" parameter, we expect the string to be the _key itself
	if t.Key != "" {
		byteKey := []byte(t.Key)
		t._key = byteKey
		return &t, nil
	}
	// If none of the above is verified, then the configuration is missing, and we can error out
	return nil, errors.New("jwt_auth_transformer_no_config")
}

// JWTSignTransformer adds JWT tokens to the request
type JWTSignTransformer struct {
	// _privateKey is the loaded and parsed private key
	_privateKey *rsa.PrivateKey
	// _key is the byte representation of a key
	_key []byte
	// Pem is the path to a PEM certificate
	Pem string `mapstructure:"pem"`
	// Key is the key for that certificate
	Key string `mapstructure:"key"`
	// ExistingClaims if set to true, will take existing claims from the wrapper and resign them
	ExistingClaims bool
	// Claims custom crafted claims
	Claims map[string]interface{}
	// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
	ActivateOnTags []string
}

func (t *JWTSignTransformer) ShouldExpandRequest() bool {
	return false
}

func (t *JWTSignTransformer) ShouldExpandResponse() bool {
	return false
}

func (t *JWTSignTransformer) ErrorMatches(_ error) bool {
	return false
}

func (t *JWTSignTransformer) HandleError(_ *http.ResponseWriter) {}

func (t *JWTSignTransformer) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(t.ActivateOnTags)
}

// Transform adds the JWT token to the request
func (t *JWTSignTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	var token *jwt.Token = nil
	var signedString = ""
	var err error
	// If ExistingClaims is true, then we'll take the claims from the wrapper
	if t.ExistingClaims {
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, wrapper.Claims)
	} else {
		// If it's not true, then we build some hand crafter claims
		claims := jwt.MapClaims{}
		for k, v := range t.Claims {
			if isString(v) {
				parsedClaim, err := Templ(v.(string), wrapper)
				if err != nil {
					return nil, err
				}
				claims[k] = parsedClaim
			} else {
				claims[k] = v
			}
		}
		// Let's create then token
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	}
	// If we have a loaded private key, we use it to sign
	if t._privateKey != nil {
		signedString, err = token.SignedString(t._privateKey)
		if err != nil {
			return nil, err
		}
	} else {
		// If we've been provided a key in the conf, then we use it to sign
		signedString, err = token.SignedString(t._key)
		if err != nil {
			return nil, err
		}
	}
	wrapper.Request.Header.Set("authorization", "Bearer "+signedString)
	return wrapper, nil
}

// NewJWTSignTransformer is the constructor for the JWTSignTransformer
func NewJWTSignTransformer(activateOnTags []string, params map[string]interface{}) (*JWTSignTransformer, error) {
	t := JWTSignTransformer{ActivateOnTags: activateOnTags}

	err := DecodeAndTempl(params, &t, nil, []string{"claims"})
	t.Claims = convertMaps(t.Claims).(map[string]interface{})
	if err != nil {
		return nil, err
	}
	// If there's a "pem" parameter, then we treat that as a path to the certificate
	if t.Pem != "" {
		cert, err := ioutil.ReadFile(t.Pem)
		if err != nil {
			return nil, err
		}
		key, err := jwt.ParseRSAPrivateKeyFromPEM(cert)
		t._privateKey = key
		return &t, err
	}
	// If there's a "_key" parameter, we expect the string to be the _key itself
	if t.Key != "" {
		byteKey := []byte(t.Key)
		t._key = byteKey
		return &t, nil
	}
	// If none of the above is verified, then the configuration is missing, and we can error out
	return nil, errors.New("jwt_sign_transformer_no_config")

}
