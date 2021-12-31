package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// CaptureMessage represents the serialization of an API conversation, forwarded to Fortress
type CaptureMessage struct {
	Request    RequestCapture         `json:"request"`
	Response   ResponseCapture        `json:"response"`
	Definition map[string]interface{} `json:"definition"`
	Meta       map[string]interface{} `json:"meta"`
}

// RequestCapture represents the serialization of an API Request
type RequestCapture struct {
	IP      string              `json:"ip"`
	Body    string              `json:"body"`
	Url     string              `json:"url"`
	Size    int                 `json:"size"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
}

// ResponseCapture represents the serialization of an API response
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
	channel                    chan *APIWrapper
	Uri                        string
	RequestContentTypeRegexp   string
	_requestContentTypeRegexp  *regexp.Regexp
	ResponseContentTypeRegexp  string
	_responseContentTypeRegexp *regexp.Regexp
	block                      bool
	httpClient                 *http.Client
	Headers                    map[string]string
	logger                     *LogHelper
	Format                     string
}

// GetChannel returns the channel for the sidecar
func (s *CaptureSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

// Consume starts the consumption workers
func (s *CaptureSidecar) Consume(quantity int) {
	var captureFunc func([]byte)
	if strings.HasPrefix(s.Uri, "http") {
		s.httpClient = &http.Client{}
		captureFunc = s.CaptureHttp
	} else {
		if strings.HasPrefix(s.Uri, "file://") {
			localUrl, err := url.Parse(s.Uri)
			if err != nil {
				log.Error("Could not parse capture URI. Disabling sidecar", err, nil)
				return
			}
			s.Uri = localUrl.Host + localUrl.Path
		}
		s.logger = NewLogHelperFromConfig(LoggerConfig{Path: s.Uri, Format: s.Format, Level: "info"})
		captureFunc = s.CaptureLogger
	}

	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				func() {
					reqCT := msg.Request.Header.Get("content-type")
					resCT := msg.Response.Header.Get("content-type")
					reqRx := false
					resRx := false
					if s._requestContentTypeRegexp != nil {
						reqRx = s._requestContentTypeRegexp.MatchString(reqCT)
					}
					if s._responseContentTypeRegexp != nil {
						resRx = s._responseContentTypeRegexp.MatchString(resCT)
					}

					if reqRx && resRx {
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

// NewCaptureSidecarFromParams is the constructor
func NewCaptureSidecarFromParams(block bool, params map[string]interface{}) (*CaptureSidecar, error) {
	sidecar := CaptureSidecar{channel: make(chan *APIWrapper), block: block}
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
