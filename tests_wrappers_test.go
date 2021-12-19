package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestAPIWrapper_Clone(t *testing.T) {
	wrapper := APIWrapper{Request: &http.Request{Method: "GET"},
							Response: &http.Response{StatusCode: 200},
							RequestBody: []byte("foo"),
							ResponseBody: []byte("foo")}
	w2 := wrapper.Clone()
	if wrapper.Request == w2.Request ||
		&wrapper.RequestBody == &w2.RequestBody ||
		&wrapper.ResponseBody == &w2.ResponseBody {
		t.Error("Clone unsuccessful")
	}
}

func TestAPIWrapper_ExpandRequest(t *testing.T) {
	wrapper := APIWrapper{Request: &http.Request{Method: "GET"},
		Response: &http.Response{StatusCode: 200}}
	wrapper.ExpandRequest()
	if len(wrapper.RequestBody) > 0 {
		t.Error("No body expansion failed")
	}
	wrapper.Request.Body = ioutil.NopCloser(bytes.NewReader([]byte("foo")))
	wrapper.ExpandRequest()
	if string(wrapper.RequestBody) != "foo"{
		t.Error("Request expansion failed")
	}
}

func TestAPIWrapper_ExpandResponse(t *testing.T) {
	wrapper := APIWrapper{Request: &http.Request{Method: "GET"},
		Response: &http.Response{StatusCode: 200}}
	wrapper.ExpandResponse()
	if len(wrapper.ResponseBody) > 0 {
		t.Error("No body expansion failed")
	}
	wrapper.Response.Body = ioutil.NopCloser(bytes.NewReader([]byte("foo")))
	wrapper.ExpandResponse()
	if string(wrapper.ResponseBody) != "foo"{
		t.Error("Request expansion failed")
	}
}

func TestAPIWrapper_Templ(t *testing.T) {
	wrapper := APIWrapper{Request: &http.Request{Method: "GET"},
		Response: &http.Response{StatusCode: 200}}
	res,_ := wrapper.Templ("{{.Request.Method}}")
	if res != "GET" {
		t.Error("wrapper templ not working")
	}
}

func TestReqWithContext(t *testing.T) {
	req := &http.Request{Method: "GET"}
	if GetWrapper(req) != nil {
		t.Error("Empty wrapper retrieval has something wrong")
	}
	req.Header = http.Header{}
	req.Header.Set("Authorization","Basic Zm9vOmJhcg==")
	rule := Rule{Origin: "foobar"}
	req = ReqWithContext(req,&rule)
	if req.Context().Value("wrapper").(*APIWrapper).Rule.Origin != "foobar" {
		t.Error("req context not persisted correctly")
	}
	if req.Context().Value("wrapper").(*APIWrapper).Username != "foo" {
		t.Error("username not present in context")
	}
	wp := GetWrapper(req)
	if wp != req.Context().Value("wrapper") {
		t.Error("wrapper retrieval is broken")
	}
}

func TestAPIMetrics_Measurements(t *testing.T) {
	req := &http.Request{Method: "GET"}
	rule := Rule{Origin: "foobar"}
	req = ReqWithContext(req,&rule)
	wrapper := GetWrapper(req)
	wrapper.Metrics.TransactionEnd = time.Now().Add(time.Duration(10*time.Millisecond))
	wrapper.Metrics.ReqTransStart = time.Now()
	wrapper.Metrics.ResTransStart = time.Now()
	wrapper.Metrics.ReqTransEnd = time.Now().Add(time.Duration(10*time.Millisecond))
	wrapper.Metrics.ResTransEnd = time.Now().Add(time.Duration(10*time.Millisecond))

	if wrapper.Metrics.Transaction() < 10 || wrapper.Metrics.Transaction() > 15 ||
		wrapper.Metrics.ResTransformation() < 10 || wrapper.Metrics.ResTransformation() > 15 ||
		wrapper.Metrics.ReqTransformation() < 10 || wrapper.Metrics.ReqTransformation() > 15 {
		t.Error("Metrics monitoring is not working according to plan")
	}
}