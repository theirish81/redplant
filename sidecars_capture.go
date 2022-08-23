package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"regexp"
	"time"
)

// CaptureMessage represents the serialization of an API conversation, forwarded to Fortress
// Request is the captured request
// Response is the captured response
// Definition represent meta information of what rules where applied
// Meta is free-hand meta information
type CaptureMessage struct {
	Request    RequestCapture  `json:"request"`
	Response   ResponseCapture `json:"response"`
	Definition AnyMap          `json:"definition"`
	Meta       AnyMap          `json:"meta"`
}

// RequestCapture represents the serialization of an API Request
// IP is the requesting IP address
// Body is the requested body
// Url is the URL being requested
// Size is the size of the body
// Method is the method of the request
// Headers are the request headers
type RequestCapture struct {
	IP      string              `json:"ip"`
	Body    string              `json:"body"`
	Url     string              `json:"url"`
	Size    int                 `json:"size"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
}

// ResponseCapture represents the serialization of an API response
// Body is the response body
// Status is the status code
// Size is the size of the response body
// Headers are the response headers
type ResponseCapture struct {
	Body    string              `json:"body"`
	Status  int                 `json:"status"`
	Size    int                 `json:"size"`
	Headers map[string][]string `json:"headers"`
}

// CaptureResponse captures the response in a wrapper and populates a CaptureMessage
func CaptureResponse(wrapper *APIWrapper) *CaptureMessage {
	captureMessage := CaptureMessage{
		Request: RequestCapture{
			IP:      addresser.RealIP(wrapper.Request.Request),
			Url:     wrapper.Request.URL.String(),
			Method:  wrapper.Request.Method,
			Headers: wrapper.Request.Header,
			Body:    string(wrapper.Request.InflatedBody),
		},
		Response: ResponseCapture{
			Size:    len(wrapper.Response.InflatedBody),
			Status:  wrapper.Response.StatusCode,
			Headers: wrapper.Response.Header,
			Body:    string(wrapper.Response.InflatedBody),
		},
		Definition: AnyMap{"origin": wrapper.Rule.Origin, "pattern": wrapper.Rule.Pattern},
		Meta:       make(AnyMap),
	}
	return &captureMessage
}

// CaptureSidecar is the sidecar fo capturing API conversations
// channel is the go inbound channel
// Uri is the destination of the capture
// RequestContentTypeRegexp is the regexp for the allowed request content type in form of string
// _requestContentTypeRegexp is the compiled regexp for the allowed request content type
// ResponseContentTypeRegexp is the regexp for the allowed response content type in form of string
// _responseContentTypeRegexp is the compiled regexp for the allowed response content type
// block, if true, will put back-pressure on the data flow if all workers are busy
// httpClient is an HTTP Client instance, if we're using a web destination
// Headers is a set of optional request headers we may want to send to the destination
// Timeout is the HTTP client timeout
// logger is logger implementation, if we're using a local logging mechanism
// Format determines the log format for local logging
type CaptureSidecar struct {
	channel                    chan *APIWrapper
	Uri                        string
	RequestContentTypeRegexp   string
	_requestContentTypeRegexp  *regexp.Regexp
	ResponseContentTypeRegexp  string
	_responseContentTypeRegexp *regexp.Regexp
	block                      bool
	dropOnOverflow             bool
	httpClient                 *http.Client
	Headers                    StringMap
	Timeout                    string
	logger                     *STLogHelper
	Format                     string
	ActivateOnTags             []string
}

// GetChannel returns the channel for the sidecar
func (s *CaptureSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

// Consume starts the consumption workers
func (s *CaptureSidecar) Consume(quantity int) {
	var captureFunc func([]byte)
	// If it's a web URL, then we'll use the HTTP capture function
	if IsHTTP(s.Uri) {
		to, err := time.ParseDuration(s.Timeout)
		if err != nil {
			log.Warn("Could not parse HTTP client timeout in Capture sidecar. Defaulting to 5s", err, nil)
			to, _ = time.ParseDuration("5s")
		}
		s.httpClient = &http.Client{Timeout: to}
		captureFunc = s.CaptureHttp
	}
	captureFunc = s.CaptureLogger

	// For each worker...
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				func() {
					// We obtain the content-type of both the request and the response
					reqCT := msg.Request.Header.Get("content-type")
					resCT := msg.Response.Header.Get("content-type")
					reqRx := false
					resRx := false
					// Does the request content type match the regexp?
					if s._requestContentTypeRegexp != nil {
						reqRx = s._requestContentTypeRegexp.MatchString(reqCT)
					}
					// Does the response content type match the regexp?
					if s._responseContentTypeRegexp != nil {
						resRx = s._responseContentTypeRegexp.MatchString(resCT)
					}
					// Do they agree?
					if reqRx && resRx {
						// Then we capture
						capture := CaptureResponse(msg)
						data, _ := json.Marshal(capture)
						captureFunc(data)
					}
				}()
			}
		}()
	}
}

// CaptureHttp is the implementation of the HTTP capture
func (s *CaptureSidecar) CaptureHttp(data []byte) {
	reader := bytes.NewReader(data)
	outboundRequest, err := http.NewRequest("POST", s.Uri, reader)
	if err != nil {
		log.Error("Error creating the request during capture", err, AnyMap{"uri": s.Uri})
		return
	}
	outboundRequest.Header.Set("content-type", "application/json")
	// Setting the custom headers
	for k, v := range s.Headers {
		outboundRequest.Header.Set(k, v)
	}
	resp, err := s.httpClient.Do(outboundRequest)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		log.Error("Error while connecting to logger service ", err, nil)
	}
	if resp.StatusCode >= 400 {
		log.Error("Received "+resp.Status+" status code while connecting to logger service", nil, nil)
	}
}

// CaptureLogger is the implementation of the Logger capture
func (s *CaptureSidecar) CaptureLogger(data []byte) {
	s.logger.Info(string(data), nil)
}

// ShouldBlock should return true if the sidecar should block
func (s *CaptureSidecar) ShouldBlock() bool {
	return s.block
}

func (s *CaptureSidecar) ShouldDropOnOverflow() bool {
	return s.dropOnOverflow
}

func (s *CaptureSidecar) ShouldExpandRequest() bool {
	return true
}

func (s *CaptureSidecar) ShouldExpandResponse() bool {
	return true
}

func (s *CaptureSidecar) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(s.ActivateOnTags)
}

// NewCaptureSidecarFromParams is the constructor
func NewCaptureSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, params AnyMap) (*CaptureSidecar, error) {
	sidecar := CaptureSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	err := template.DecodeAndTempl(params, &sidecar, nil, []string{})
	if err != nil {
		return nil, err
	}
	if sidecar.RequestContentTypeRegexp != "" {
		sidecar._requestContentTypeRegexp, err = regexp.Compile(sidecar.RequestContentTypeRegexp)
		if err != nil {
			return nil, err
		}
	}
	if sidecar.ResponseContentTypeRegexp != "" {
		sidecar._responseContentTypeRegexp, err = regexp.Compile(sidecar.ResponseContentTypeRegexp)
		if err != nil {
			return nil, err
		}
	}
	sidecar.logger = NewSTLogHelper(logCfg)
	return &sidecar, nil
}
