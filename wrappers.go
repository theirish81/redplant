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
	Request  *http.Request
	Response *http.Response
	RequestBody		[]byte
	ResponseBody	[]byte
	Claims			*jwt.MapClaims
	Rule			*Rule
	Metrics			*APIMetrics
	Err				error
	Username		string
	Variables		*map[string]string
}

func (w *APIWrapper) Clone() *APIWrapper {
	return &APIWrapper{Request: w.Request.Clone(w.Request.Context()), Response: w.Response, RequestBody: w.RequestBody,
		ResponseBody: w.ResponseBody, Claims:w.Claims, Rule: w.Rule, Metrics: w.Metrics, Err: w.Err}
}

// ExpandRequest will turn the Request body into a byte array, stored in the APIWrapper itself
func (w *APIWrapper) ExpandRequest() {
	if len(w.RequestBody) == 0 && w.Request.Body != nil {
		w.RequestBody,_ = io.ReadAll(w.Request.Body)
		w.Request.Body = ioutil.NopCloser(bytes.NewReader(w.RequestBody))
	}
}
func (w *APIWrapper) ExpandResponse() {
	if len(w.ResponseBody) == 0 && w.Response.Body != nil {
		w.ResponseBody, _ = io.ReadAll(w.Response.Body)
		w.Response.Body = ioutil.NopCloser(bytes.NewReader(w.ResponseBody))
	}
}

func (w *APIWrapper) Templ(data string) (string,error) {
	return Templ(data,w)
}

type APIMetrics struct {
	TransactionStart	time.Time
	TransactionEnd		time.Time
	ReqTransStart		time.Time
	ReqTransEnd			time.Time
	ResTransStart		time.Time
	ResTransEnd			time.Time

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

func ReqWithContext(req *http.Request, rule *Rule) *http.Request {
	ctx := req.Context()
	wrapper := &APIWrapper{Rule: rule,Metrics: &APIMetrics{TransactionStart: time.Now()}, Variables: &config.Variables}
	un,_,ok := req.BasicAuth()
	if ok {
		wrapper.Username = un
	}

	ctx = context.WithValue(ctx,"wrapper", wrapper)
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