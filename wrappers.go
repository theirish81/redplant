package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"time"
)

// APIRequest is a wrapper around http.Request
type APIRequest struct {
	*http.Request
	ExpandedBody []byte
	ParsedBody   any
	UrlVars      map[string]string
}

// NewAPIRequest is the constructor for APIRequest
func NewAPIRequest(req *http.Request) *APIRequest {
	return &APIRequest{Request: req, UrlVars: mux.Vars(req)}
}

// APIResponse is the wrapper around http.Response
type APIResponse struct {
	*http.Response
	ExpandedBody []byte
	ParsedBody   any
}

// NewAPIResponse is the constructor for APIResponse
func NewAPIResponse(res *http.Response) *APIResponse {
	return &APIResponse{Response: res}
}

// Clone shallow clones the response
func (r *APIResponse) Clone() *APIResponse {
	if r == nil {
		return nil
	}
	return &APIResponse{r.Response, r.ExpandedBody, r.ParsedBody}
}

// Clone shallow clones the request
func (r *APIRequest) Clone(ctx context.Context) *APIRequest {
	if r == nil {
		return nil
	}
	return &APIRequest{r.Request.Clone(ctx), r.ExpandedBody, r.ParsedBody, r.UrlVars}
}

// APIWrapper wraps a Request and a response
type APIWrapper struct {
	ID             string
	Context        context.Context
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
	return &APIWrapper{ID: w.ID, Context: w.Context, Request: w.Request.Clone(w.Request.Context()), Response: w.Response.Clone(),
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
	if len(w.Request.ExpandedBody) == 0 && w.Request.Body != nil {
		rawBody, _ := io.ReadAll(w.Request.Body)
		rawReader := bytes.NewReader(rawBody)
		if IsGZIP(w.Request.TransferEncoding) {
			gzipReader, _ := gzip.NewReader(rawReader)
			w.Request.ExpandedBody, _ = io.ReadAll(gzipReader)
		} else {
			w.Request.ExpandedBody, _ = io.ReadAll(rawReader)
		}
		w.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
	}
}

// ExpandResponse will turn the Response body into a byte array, stored in the APIWrapper itself
func (w *APIWrapper) ExpandResponse() {
	if len(w.Response.ExpandedBody) == 0 && w.Response.Body != nil {
		rawBody, err := io.ReadAll(w.Response.Body)
		if err != nil {
			log.LogErr("could not read from response body stream", err, w, log.Warn)
		}
		if rawBody != nil && len(rawBody) > 0 {
			rawReader := bytes.NewReader(rawBody)
			if w.Response.Uncompressed {
				w.Response.ExpandedBody, err = io.ReadAll(rawReader)
				if err != nil {
					log.LogErr("could not read uncompressed response raw body", err, w, log.Warn)
				}
			} else {
				gzipReader, err := gzip.NewReader(rawReader)
				if err == nil {
					w.Response.ExpandedBody, _ = io.ReadAll(gzipReader)
				} else {
					log.LogErr("could not decompress response body. Falling back to uncompressed", err, w, log.Warn)
					w.Response.Uncompressed = true
					w.Response.Body = io.NopCloser(bytes.NewReader(rawBody))
					w.ExpandResponse()
					return
				}
			}
		}
		w.Response.Body = io.NopCloser(bytes.NewReader(rawBody))
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
func (w *APIWrapper) Templ(ctx context.Context, data string) (string, error) {
	return template.Templ(ctx, data, w)
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
		Context:        ctx,
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
