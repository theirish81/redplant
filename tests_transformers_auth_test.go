package main

import (
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"os"
	"testing"
)

func TestBasicAuthTransformer_Transform(t *testing.T) {
	transformer, _ := NewBasicAuthTransformer([]string{}, map[string]interface{}{"username": "foo", "password": "bar"})
	if transformer.Username != "foo" || transformer.Password != "bar" {
		t.Error("Wrong param assignment")
	}
	req := http.Request{Header: http.Header{}}
	wrapper := APIWrapper{Request: &req}
	req.Header.Set("Authorization", "Basic Zm9vOmJhcg==")
	_, err := transformer.Transform(&wrapper)
	if err != nil {
		t.Error("Basic auth did not authorise")
	}
	if wrapper.Username != "foo" {
		t.Error("Username not captured")
	}
	req.Header.Set("Authorization", "Basic bananas")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Did not block on invalid basic")
	}
	req.Header.Del("Authorization")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Did not block on missing auth")
	}
}
func TestJWTAuthTransformerPem_Transform(t *testing.T) {
	transformer, _ := NewJWTAuthTransformer([]string{}, map[string]interface{}{"pem": "etc/publicKey"})
	claims := jwt.MapClaims{}
	claims["data"] = "123detectme"

	privateKey, _ := os.ReadFile("etc/privateKey")
	signKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	at := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token, _ := at.SignedString(signKey)
	req := http.Request{Header: http.Header{}}
	req.Header.Set("Authorization", "Bearer "+token)
	wrapper := APIWrapper{Request: &req}
	_, err := transformer.Transform(&wrapper)
	if err != nil {
		t.Error("Auth Failed for no reason")
	}
	c1 := (*wrapper.Claims)["data"]
	c2 := claims["data"].(string)
	if c1 != c2 {
		t.Error("Claims not matching")
	}

	req.Header.Set("Authorization", "Bearer abc")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Auth should have failed with wrong token")
	}

	req.Header.Del("Authorization")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Auth should have failed with no token")
	}
}

func TestJWTAuthTransformerKey_Transform(t *testing.T) {
	transformer, _ := NewJWTAuthTransformer([]string{}, map[string]interface{}{"key": "foobar"})
	claims := jwt.MapClaims{}
	claims["data"] = "123detectme"
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := at.SignedString([]byte("foobar"))
	req := http.Request{Header: http.Header{}}
	req.Header.Set("Authorization", "Bearer "+token)
	wrapper := APIWrapper{Request: &req}
	_, err := transformer.Transform(&wrapper)
	if err != nil {
		t.Error("Auth Failed for no reason")
	}

	c1 := (*wrapper.Claims)["data"]
	c2 := claims["data"].(string)
	if c1 != c2 {
		t.Error("Claims not matching")
	}

	req.Header.Set("Authorization", "Bearer abc")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Auth should have failed with wrong token")
	}

	req.Header.Del("Authorization")
	_, err = transformer.Transform(&wrapper)
	if err == nil {
		t.Error("Auth should have failed with no token")
	}
}
