package main

import (
	"bytes"
	"context"
	"github.com/dgrijalva/jwt-go"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// APIWrapper wraps a Request and a response
type APIWrapper struct {
	Request        *http.Request
	Response       *http.Response
	ResponseWriter http.ResponseWriter
	RequestBody    []byte
	ResponseBody   []byte
	Claims         *jwt.MapClaims
	Rule           *Rule
	Metrics        *APIMetrics
	Err            error
	Username       string
	Variables      *map[string]string
	RealIP         string
	Tags           []string
	ApplyHeaders   http.Header
	// When set to true, it means that the connection has been hijacked. This is the case when websockets
	// are involved
	Hijacked bool
}

func (w *APIWrapper) Clone() *APIWrapper {
	return &APIWrapper{Request: w.Request.Clone(w.Request.Context()), Response: w.Response, RequestBody: w.RequestBody,
		ResponseBody: w.ResponseBody, Claims: w.Claims, Rule: w.Rule, Metrics: w.Metrics, Err: w.Err, RealIP: w.RealIP,
		Tags: w.Tags, ApplyHeaders: w.ApplyHeaders, Hijacked: w.Hijacked}
}

func (w *APIWrapper) ExpandRequestIfNeeded() {
	if w.Rule.Request._transformers.ShouldExpandRequest(); w.Rule.Response._transformers.ShouldExpandRequest() ||
		w.Rule.Request._sidecars.ShouldExpandRequest() ||
		w.Rule.Response._sidecars.ShouldExpandRequest() {
		w.ExpandRequest()
	}
}

func (w *APIWrapper) ExpandResponseIfNeeded() {
	if w.Rule.Response._transformers.ShouldExpandResponse() ||
		w.Rule.Response._sidecars.ShouldExpandResponse() {
		w.ExpandResponse()
	}
}

// ExpandRequest will turn the Request body into a byte array, stored in the APIWrapper itself
func (w *APIWrapper) ExpandRequest() {
	if len(w.RequestBody) == 0 && w.Request.Body != nil {
		w.RequestBody, _ = io.ReadAll(w.Request.Body)
		w.Request.Body = ioutil.NopCloser(bytes.NewReader(w.RequestBody))
	}
}
func (w *APIWrapper) ExpandResponse() {
	if len(w.ResponseBody) == 0 && w.Response.Body != nil {
		w.ResponseBody, _ = io.ReadAll(w.Response.Body)
		w.Response.Body = ioutil.NopCloser(bytes.NewReader(w.ResponseBody))
	}
}

func (w *APIWrapper) HasTag(tags []string) bool {
	if len(tags) == 0 {
		return true
	}
	for _, a1 := range w.Tags {
		for _, a2 := range tags {
			if a1 == a2 {
				return true
			}
		}
	}
	return false
}

func (w *APIWrapper) Templ(data string) (string, error) {
	return Templ(data, w)
}

type APIMetrics struct {
	TransactionStart time.Time
	TransactionEnd   time.Time
	ReqTransStart    time.Time
	ReqTransEnd      time.Time
	ResTransStart    time.Time
	ResTransEnd      time.Time
}

func (m *APIMetrics) Transaction() int64 {
	return m.TransactionEnd.Sub(m.TransactionStart).Milliseconds()
}
func (m *APIMetrics) ReqTransformation() int64 {
	return m.ReqTransEnd.Sub(m.ReqTransStart).Milliseconds()
}
func (m *APIMetrics) ResTransformation() int64 {
	return m.ResTransEnd.Sub(m.ResTransStart).Milliseconds()
}

func ReqWithContext(req *http.Request, responseWriter http.ResponseWriter, rule *Rule) *http.Request {
	ctx := req.Context()
	wrapper := &APIWrapper{Rule: rule, Metrics: &APIMetrics{TransactionStart: time.Now()},
		Tags:           []string{},
		Variables:      &config.Variables,
		RealIP:         addresser.RealIP(req),
		ResponseWriter: responseWriter,
		ApplyHeaders:   http.Header{}}
	un, _, ok := req.BasicAuth()
	if ok {
		wrapper.Username = un
	}

	ctx = context.WithValue(ctx, "wrapper", wrapper)
	req = req.WithContext(ctx)
	return req
}
func GetWrapper(req *http.Request) *APIWrapper {
	wrapper := req.Context().Value("wrapper")
	if wrapper == nil {
		return nil
	}
	return wrapper.(*APIWrapper)
}
