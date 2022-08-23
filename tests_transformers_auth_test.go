package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"testing"
)

func TestBasicAuthTransformer_Transform(t *testing.T) {
	transformer, _ := NewBasicAuthTransformer([]string{}, nil, map[string]any{"username": "foo", "password": "bar"})
	if transformer.Username != "foo" || transformer.Password != "bar" {
		t.Error("Wrong param assignment")
	}
	ux, _ := url.Parse("http://www.example.com")
	req := http.Request{Header: http.Header{}, URL: ux}
	wrapper := APIWrapper{Request: NewAPIRequest(&req)}
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
	transformer, _ := NewJWTAuthTransformer([]string{}, nil, map[string]any{"pem": "etc/publicKey"})
	claims := jwt.MapClaims{}
	claims["data"] = "123detectme"

	privateKey, _ := os.ReadFile("etc/privateKey")
	signKey, _ := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	at := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token, _ := at.SignedString(signKey)
	ux, _ := url.Parse("http://www.example.com")
	req := http.Request{Header: http.Header{}, URL: ux}
	req.Header.Set("Authorization", "Bearer "+token)
	wrapper := APIWrapper{Request: NewAPIRequest(&req)}
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
	transformer, _ := NewJWTAuthTransformer([]string{}, nil, map[string]any{"key": "foobar"})
	claims := jwt.MapClaims{}
	claims["data"] = "123detectme"
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := at.SignedString([]byte("foobar"))
	ux, _ := url.Parse("http://www.example.com")
	req := http.Request{Header: http.Header{}, URL: ux}
	req.Header.Set("Authorization", "Bearer "+token)
	wrapper := APIWrapper{Request: NewAPIRequest(&req)}
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

func TestJWTSignTransformer_Transform(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	params := map[string]any{
		"pem":    "etc/privateKey",
		"claims": map[string]any{"foo": "bar"},
	}
	transformer, _ := NewJWTSignTransformer([]string{}, nil, params)
	ux, _ := url.Parse("http://example.com")
	wrapper := APIWrapper{Request: &APIRequest{Request: &http.Request{URL: ux, Header: http.Header{}}}}
	_, _ = transformer.Transform(&wrapper)
	if wrapper.Request.Header.Get("Authorization") != "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIifQ.H8CXyGIxW7T0bhfNmvYKVc-6lj0qI0Neg8Q7yw0gqkJk1pbdh0lSGkJXRzTGKUZd81BULeYXlB9_q9h1VVNi3DA21vQn6Dznv900KI2vzB7EIX180pJnpU1WKb8-XyFVBWF9y0yHSWd5LpcOLAXYh3t2JJWpmTv2LdBFsMdHzOIXnchNmEgv_5NVD-EKKKugrffGRw3xneDzS9dxwdkFs17yrQaM_oFFeUoJGoXmWotcax6qZJCjbPgSiLl4qqU6uuwGDxilhJ5XC9Rp-72k606bS4rmVedSx7JlBwsSS8eQ6_aKsskeKzEM6Yjy1McyvbjyU2GHaXDGQytCCsZreOcr0nglMJ12JUAWmKHj-rzL7ulJaDr1ZExD8qrl-Hw9ojw5fmhqcHr7WdXsATX4Vyy7sF7Cfsnfcws3EH1HpC4heNbwSdG0hyBNuRCwPjl9rpKE6tRUDP5H96Tj5JtbuMmWuXp5N8nBZUBxpCTB1RN7fjLMORwoCo9gj-_cAJpH8GrIvH3qEsbm1gYtgaQ2OGuptN5J6Mfwbwas8FefGUS0mPD-mxPROl0nkh9rwdAfA0hZroiod10Bf5ha0-UHxnQOCZ_R5ro--SOTrft68p66hiYln-kH4jkoE67TYh4dLQUTY7uroUmPiQ5lNbjCSI2ZpyTdv7PcDzjDLYmsbnc" {
		t.Error("JWT sign transformer not working as expected")
	}
}
