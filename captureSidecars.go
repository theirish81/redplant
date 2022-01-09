package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// CaptureMessage represents the serialization of an API conversation, forwarded to Fortress
type CaptureMessage struct {
	// Request is the captured request
	Request RequestCapture `json:"request"`
	// Response is the captured response
	Response ResponseCapture `json:"response"`
	// Definition represent meta information of what rules where applied
	Definition map[string]interface{} `json:"definition"`
	// Meta is free-hand meta information
	Meta map[string]interface{} `json:"meta"`
}

// RequestCapture represents the serialization of an API Request
type RequestCapture struct {
	// IP is the requesting IP address
	IP string `json:"ip"`
	// Body is the requested body
	Body string `json:"body"`
	// Url is the URL being requested
	Url string `json:"url"`
	// Size is the size of the body
	Size int `json:"size"`
	// Method is the method of the request
	Method string `json:"method"`
	// Headers are the request headers
	Headers map[string][]string `json:"headers"`
}

// ResponseCapture represents the serialization of an API response
type ResponseCapture struct {
	// Body is the response body
	Body string `json:"body"`
	// Status is the status code
	Status int `json:"status"`
	// Size is the size of the response body
	Size int `json:"size"`
	// Headers are the response headers
	Headers map[string][]string `json:"headers"`
}

// CaptureResponse captures the response in a wrapper and populates a CaptureMessage
func CaptureResponse(wrapper *APIWrapper) *CaptureMessage {
	captureMessage := CaptureMessage{
		Request: RequestCapture{
			IP:      addresser.RealIP(wrapper.Request),
			Url:     wrapper.Request.URL.String(),
			Method:  wrapper.Request.Method,
			Headers: wrapper.Request.Header,
			Body:    string(wrapper.RequestBody),
		},
		Response: ResponseCapture{
			Size:    len(wrapper.ResponseBody),
			Status:  wrapper.Response.StatusCode,
			Headers: wrapper.Response.Header,
			Body:    string(wrapper.ResponseBody),
		},
		Definition: map[string]interface{}{"origin": wrapper.Rule.Origin, "pattern": wrapper.Rule.Pattern},
		Meta:       make(map[string]interface{}),
	}
	return &captureMessage
}

// CaptureSidecar is the sidecar fo capturing API conversations
type CaptureSidecar struct {
	// channel is the go inbound channel
	channel chan *APIWrapper
	// Uri is the destination of the capture
	Uri string
	// RequestContentTypeRegexp is the regexp for the allowed request content type in form of string
	RequestContentTypeRegexp string
	// _requestContentTypeRegexp is the compiled regexp for the allowed request content type
	_requestContentTypeRegexp *regexp.Regexp
	// ResponseContentTypeRegexp is the regexp for the allowed response content type in form of string
	ResponseContentTypeRegexp string
	// _responseContentTypeRegexp is the compiled regexp for the allowed response content type
	_responseContentTypeRegexp *regexp.Regexp
	// block, if true, will put back-pressure on the data flow if all workers are busy
	block bool
	// httpClient is an HTTP Client instance, if we're using a web destination
	httpClient *http.Client
	// Headers is a set of optional request headers we may want to send to the destination
	Headers map[string]string
	// Timeout is the HTTP client timeout
	Timeout string
	// logger is logger implementation, if we're using a local logging mechanism
	logger *LogHelper
	// Format determines the log format for local logging
	Format         string
	ActivateOnTags []string
}

// GetChannel returns the channel for the sidecar
func (s *CaptureSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

// Consume starts the consumption workers
func (s *CaptureSidecar) Consume(quantity int) {
	var captureFunc func([]byte)
	// If it's a web URL, then we'll use the HTTP capture function
	if hasPrefixes(s.Uri, []string{"http://", "https://"}) {
		to, err := time.ParseDuration(s.Timeout)
		if err != nil {
			log.Warn("Could not parse HTTP client timeout in Capture sidecar. Defaulting to 5s", err, nil)
			to, _ = time.ParseDuration("5s")
		}
		s.httpClient = &http.Client{Timeout: to}
		captureFunc = s.CaptureHttp
	} else {
		// Otherwise, it's a local file. However, we want to check whether it's a path or a file URI
		if strings.HasPrefix(s.Uri, "file://") {
			localUrl, err := url.Parse(s.Uri)
			if err != nil {
				log.Error("Could not parse capture URI. Disabling sidecar", err, nil)
				return
			}
			s.Uri = localUrl.Host + localUrl.Path
		}
		// Creating logger and assigning local logging function
		s.logger = NewLogHelperFromConfig(LoggerConfig{Path: s.Uri, Format: s.Format, Level: "info"})
		captureFunc = s.CaptureLogger
	}

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
		log.Error("Error creating the request during capture", err, map[string]interface{}{"uri": s.Uri})
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
func NewCaptureSidecarFromParams(block bool, queue int, activateOnTags []string, params map[string]interface{}) (*CaptureSidecar, error) {
	sidecar := CaptureSidecar{channel: make(chan *APIWrapper, queue), block: block, ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &sidecar, nil, []string{})
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
	return &sidecar, nil
}
