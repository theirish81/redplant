package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"os"
	"strings"
)

// JWTAuthTransformer will block any request without a Bearer token or a token whose signature cannot be verified.
// In addition, it will store claims in the wrapper.
// _publicKey is the loaded and parsed public key
// _key is the key provided in the conf, in the form of bytes
// Pem is the path to a PEM certificate
// Key is the key provided in the conf, in the form of a string
// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
type JWTAuthTransformer struct {
	_publicKey     *rsa.PublicKey
	_key           []byte
	Pem            string
	Key            string
	ActivateOnTags []string
	log            *STLogHelper
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
	t.log.Log("JWT auth triggered", wrapper, t.log.Debug)

	header := wrapper.Request.Header.Get("authorization")

	// if no Bearer prefix, then we have no auth
	if !strings.HasPrefix(header, "Bearer") {
		t.log.Log("no auth header provided", wrapper, t.log.Debug)
		t.log.PrometheusCounterInc("jwt_auth_denied")
		return nil, errors.New("no_auth")
	}

	// Extract the token from the header
	token := strings.TrimSpace(header[7:])
	claims := jwt.MapClaims{}
	// Parsing with signature check and claims extraction
	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (any, error) {
		// If we have a public _key, we'll use that
		if t._publicKey != nil {
			return t._publicKey, nil
		}
		// Otherwise, we'll just use the _key
		return t._key, nil
	})
	// Storing claims in the wrapper
	wrapper.Claims = &claims
	if err != nil {
		t.log.PrometheusCounterInc("jwt_auth_denied")
	}
	return wrapper, err
}

func (t *JWTAuthTransformer) ErrorMatches(err error) bool {
	return err.Error() == "signature is invalid"
}

func (t *JWTAuthTransformer) HandleError(writer *http.ResponseWriter) {
	(*writer).WriteHeader(401)
}

// NewJWTAuthTransformer creates a new JWTAuthTransformer from params
func NewJWTAuthTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*JWTAuthTransformer, error) {
	t := JWTAuthTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}
	err := template.DecodeAndTempl(context.Background(), params, &t, nil, []string{})
	if err != nil {
		return nil, err
	}
	t.log.PrometheusRegisterCounter("jwt_auth_denied")
	// If there's a "pem" parameter, then we treat that as a path to the certificate
	if t.Pem != "" {
		cert, err := os.ReadFile(t.Pem)
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
// _privateKey is the loaded and parsed private key
// _key is the byte representation of a key
// Pem is the path to a PEM certificate
// Key is the key for that certificate
// ExistingClaims if set to true, will take existing claims from the wrapper and resign them
// Claims custom crafted claims
// ActivateOnTags is a list of tags for which this plugin will activate. Leave empty for "always"
type JWTSignTransformer struct {
	_privateKey    *rsa.PrivateKey
	_key           []byte
	Pem            string
	Key            string
	ExistingClaims bool
	Claims         map[string]any
	ActivateOnTags []string
	log            *STLogHelper
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
	t.log.Log("JWT sign triggered", wrapper, t.log.Debug)
	var token *jwt.Token = nil
	var signedString = ""
	var err error
	// If ExistingClaims is true, then we'll take the claims from the wrapper
	if t.ExistingClaims {
		t.log.Log("carrying existing claims to token", wrapper, t.log.Debug)
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, wrapper.Claims)
	} else {
		t.log.Log("composing new claims", wrapper, t.log.Debug)
		// If it's not true, then we build some hand crafter claims
		claims := jwt.MapClaims{}
		for k, v := range t.Claims {
			if isString(v) {
				parsedClaim, err := template.Templ(wrapper.Context, v.(string), wrapper)
				if err != nil {
					t.log.LogErr("unable to parse claims", err, wrapper, t.log.Error)
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
			t.log.LogErr("unable to sign token", err, wrapper, log.Error)
			return nil, err
		}
	} else {
		// If we've been provided a key in the conf, then we use it to sign
		signedString, err = token.SignedString(t._key)
		if err != nil {
			t.log.LogErr("unable to sign token", err, wrapper, log.Error)
			return nil, err
		}
	}
	t.log.Log("JWT sign succeeded", wrapper, t.log.Debug)
	wrapper.Request.Header.Set("authorization", "Bearer "+signedString)
	return wrapper, nil
}

// NewJWTSignTransformer is the constructor for the JWTSignTransformer
func NewJWTSignTransformer(activateOnTags []string, logCfg *STLogConfig, params map[string]any) (*JWTSignTransformer, error) {
	t := JWTSignTransformer{ActivateOnTags: activateOnTags, log: NewSTLogHelper(logCfg)}

	err := template.DecodeAndTempl(context.Background(), params, &t, nil, []string{"Claims"})
	t.Claims = convertMaps(t.Claims).(map[string]any)
	if err != nil {
		return nil, err
	}
	// If there's a "pem" parameter, then we treat that as a path to the certificate
	if t.Pem != "" {
		cert, err := os.ReadFile(t.Pem)
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
