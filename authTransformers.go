package main

import (
	"crypto/rsa"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"strings"
)

// BasicAuthTransformer is a transformer that will block the request in case the credentials do not match the
// expectations
type BasicAuthTransformer struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// Transform will throw an error if the request doesn't match the basic auth expectations
func (t *BasicAuthTransformer) Transform(wrapper *APIWrapper) (*APIWrapper, error) {
	username, password, ok := wrapper.Request.BasicAuth()
	if ok && t.Username == username && t.Password == password {
		wrapper.Username = t.Username
		return wrapper, nil
	} else {
		return nil, errors.New("no_auth")
	}
}

// NewBasicAuthTransformer creates a BasicAuthTransformer from params
func NewBasicAuthTransformer(params map[string]interface{}) (*BasicAuthTransformer, error) {
	var t BasicAuthTransformer
	//err := mapstructure.Decode(params,&t)
	err := DecodeAndTempl(params, &t, nil)
	return &t, err
}

// JWTAuthTransformer will block any request without a Bearer token or a token whose signature cannot be verified.
// In addition, it will store claims in the wrapper.
type JWTAuthTransformer struct {
	_publicKey *rsa.PublicKey
	_key       []byte
	Pem        string `mapstructure:"pem"`
	Key        string `mapstructure:"key"`
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

// NewJWTAuthTransformer creates a new JWTAuthTransformer from params
func NewJWTAuthTransformer(params map[string]interface{}) (*JWTAuthTransformer, error) {
	t := JWTAuthTransformer{}
	err := DecodeAndTempl(params, &t, nil)
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
	// If none of the above is verified, then the configuration is missing and we can error out
	return nil, errors.New("jwt_auth_transformer_no_config")
}
