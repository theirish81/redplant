package main

import (
	"bytes"
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"io"
	"net/http"
	"time"
)

type APIRequest struct {
	*http.Request
	InflatedBody []byte
	ParsedBody   any
}

func NewAPIRequest(req *http.Request) *APIRequest {
	return &APIRequest{Request: req}
}

type APIResponse struct {
	*http.Response
	InflatedBody []byte
	ParsedBody   any
}

func NewAPIResponse(res *http.Response) *APIResponse {
	return &APIResponse{Response: res}
}

func (r *APIResponse) Clone() *APIResponse {
	if r == nil {
		return nil
	}
	return &APIResponse{r.Response, r.InflatedBody, r.ParsedBody}
}

func (r *APIRequest) Clone(ctx context.Context) *APIRequest {
	if r == nil {
		return nil
	}
	return &APIRequest{r.Request.Clone(ctx), r.InflatedBody, r.ParsedBody}
}

// APIWrapper wraps a Request and a response
type APIWrapper struct {
	ID             string
	Request        *APIRequest
	Response       *APIResponse
	ResponseWriter http.ResponseWriter
	Claims         *jwt.MapClaims
	Rule           *Rule
	Metrics        *APIMetrics
	Err            error
	Username       string
	Variables      *StringMap
	RealIP         string
	Tags           []string
	ApplyHeaders   http.Header
	// When set to true, it means that the connection has been hijacked. This is the case when websockets
	// are involved
	Hijacked bool
}

// Clone will do sort of a somewhat shallow clone of the wrapper. This is useful when sending the wrapper is being
// sent to a sidecar but also transformers apply. If we didn't clone, results may vary on timing
func (w *APIWrapper) Clone() *APIWrapper {
	return &APIWrapper{ID: w.ID, Request: w.Request.Clone(w.Request.Context()), Response: w.Response.Clone(),
		Claims: w.Claims, Rule: w.Rule, Metrics: w.Metrics, Err: w.Err, RealIP: w.RealIP,
		Tags: w.Tags, ApplyHeaders: w.ApplyHeaders, Hijacked: w.Hijacked}
}

// ExpandRequestIfNeeded determines whether the various transformers and sidecars configured for the route need the
// request expanded. If so, it expands it
func (w *APIWrapper) ExpandRequestIfNeeded() {
	if w.Rule.Request._transformers.ShouldExpandRequest(); w.Rule.Response._transformers.ShouldExpandRequest() ||
		w.Rule.Request._sidecars.ShouldExpandRequest() ||
		w.Rule.Response._sidecars.ShouldExpandRequest() {
		w.ExpandRequest()
	}
}

// ExpandResponseIfNeeded determines whether the various transformers and sidecars configured for the route need the
// // response expanded. If so, it expands it
func (w *APIWrapper) ExpandResponseIfNeeded() {
	if w.Rule.Response._transformers.ShouldExpandResponse() ||
		w.Rule.Response._sidecars.ShouldExpandResponse() {
		w.ExpandResponse()
	}
}

// ExpandRequest will turn the Request body into a byte array, stored in the APIWrapper itself
func (w *APIWrapper) ExpandRequest() {
	if len(w.Request.InflatedBody) == 0 && w.Request.Body != nil {
		w.Request.InflatedBody, _ = io.ReadAll(w.Request.Body)
		w.Request.Body = io.NopCloser(bytes.NewReader(w.Request.InflatedBody))
	}
}

// ExpandResponse will turn the Response body into a byte array, stored in the APIWrapper itself
func (w *APIWrapper) ExpandResponse() {
	if len(w.Response.InflatedBody) == 0 && w.Response.Body != nil {
		w.Response.InflatedBody, _ = io.ReadAll(w.Response.Body)
		w.Response.Body = io.NopCloser(bytes.NewReader(w.Response.InflatedBody))
	}
}

// HasTag will return true if the APIWrapper has hany of the provided tags
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

// Templ will compile the provided template using APIWrapper as scope
func (w *APIWrapper) Templ(data string) (string, error) {
	return template.Templ(data, w)
}

// APIMetrics is a collector of metrics for the transaction
type APIMetrics struct {
	TransactionStart time.Time
	TransactionEnd   time.Time
	ReqTransStart    time.Time
	ReqTransEnd      time.Time
	ResTransStart    time.Time
	ResTransEnd      time.Time
}

// Transaction will return the transaction duration in milliseconds
func (m *APIMetrics) Transaction() int64 {
	return m.TransactionEnd.Sub(m.TransactionStart).Milliseconds()
}

// ReqTransformation will return the duration of the request transformation in milliseconds
func (m *APIMetrics) ReqTransformation() int64 {
	return m.ReqTransEnd.Sub(m.ReqTransStart).Milliseconds()
}

// ResTransformation will return the duration of the response transformation in milliseconds
func (m *APIMetrics) ResTransformation() int64 {
	return m.ResTransEnd.Sub(m.ResTransStart).Milliseconds()
}

// ReqWithContext will add the RedPlant context to the provided request
func ReqWithContext(req *http.Request, responseWriter http.ResponseWriter, rule *Rule) *http.Request {
	ctx := req.Context()
	wrapper := &APIWrapper{Rule: rule, Metrics: &APIMetrics{TransactionStart: time.Now()},
		ID:             uuid.New().String(),
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

// GetWrapper will extract the wrapper from the context in the request
func GetWrapper(req *http.Request) *APIWrapper {
	wrapper := req.Context().Value("wrapper")
	if wrapper == nil {
		return nil
	}
	return wrapper.(*APIWrapper)
}
