package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

// CaptureMessage represents the serialization of an API conversation, forwarded to Fortress
type CaptureMessage struct {
	Request  RequestCapture  			`json:"request"`
	Response ResponseCapture 			`json:"response"`
	Definition	map[string]interface{}	`json:"definition"`
	Meta		map[string]interface{}	`json:"meta"`
}

// RequestCapture represents the serialization of an API Request
type RequestCapture struct {
	IP			string				`json:"ip"`
	Body      	string            	`json:"body"`
	Url			string            	`json:"url"`
	Size       	int               	`json:"size"`
	Method     	string				`json:"method"`
	Headers    	map[string][]string	`json:"headers"`
}

// ResponseCapture represents the serialization of an API response
type ResponseCapture struct {
	Body    string					`json:"body"`
	Status  int						`json:"status"`
	Size    int						`json:"size"`
	Headers map[string][]string		`json:"headers"`
}
// CaptureResponse captures the response in a wrapper and populates a CaptureMessage
func CaptureResponse(wrapper *APIWrapper) *CaptureMessage {
	captureMessage := CaptureMessage {
		Request: RequestCapture{
			IP:	addresser.RealIP(wrapper.Request),
			Url: wrapper.Request.URL.String(),
			Method: wrapper.Request.Method,
			Headers: wrapper.Request.Header,
			Body: string(wrapper.RequestBody),
		},
		Response: ResponseCapture{
			Size: len(wrapper.ResponseBody),
			Status: wrapper.Response.StatusCode,
			Headers: wrapper.Response.Header,
			Body: string(wrapper.ResponseBody),
		},
		Definition: map[string]interface{}{"origin":wrapper.Rule.Origin,"pattern":wrapper.Rule.Pattern},
		Meta: make(map[string]interface{}),
	}
	return &captureMessage
}

// CaptureSidecar is the sidecar fo capturing API conversations
type CaptureSidecar struct {
	channel						chan *APIWrapper
	uri							string
	requestContentTypeRegex 	string
	responseContentTypeRegex 	string
	block						bool
	httpClient					*http.Client
	logger						*LogHelper
}

// GetChannel returns the channel for the sidecar
func (s *CaptureSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

// Consume starts the consumption workers
func (s *CaptureSidecar) Consume(quantity int) {
	uri,err := Templ(s.uri,nil)
	if err != nil {
		log.Error("Could not parse capture uri",err,map[string]interface{}{"uri":s.uri})
		return
	}
	s.uri = uri
	var captureFunc func([]byte)
	if strings.HasPrefix(s.uri,"http") {
		s.httpClient = &http.Client{}
		captureFunc = s.CaptureHttp
	} else {
		s.logger = NewLogHelper(s.uri,logrus.InfoLevel)
		captureFunc = s.CaptureLogger
	}

	for i:=0;i<quantity;i++ {
		go func() {
			for msg := range s.GetChannel() {
				func() {
					reqCT := msg.Request.Header.Get("content-type")
					resCT := msg.Response.Header.Get("content-type")
					reqRx, err := regexp.MatchString(s.requestContentTypeRegex, reqCT)
					if err != nil {
						log.Error("Could not parse Request Regexp for CaptureSidecar ", err, map[string]interface{}{"regexp":s.requestContentTypeRegex})
						return
					}
					resRx, err := regexp.MatchString(s.responseContentTypeRegex, resCT)
					if err != nil {
						log.Error("Could not parse Response Regexp for CaptureSidecar ", err, map[string]interface{}{"regexp":s.responseContentTypeRegex})
						return
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
	resp,err := s.httpClient.Post(s.uri,"application/json",reader)
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		log.Error("Error while connecting to logger service ",err,nil)
	}
	if resp.StatusCode >= 400 {
		log.Error("Received "+resp.Status+" status code while connecting to logger service",nil,nil)
	}
}

// CaptureLogger is the implementation of the Logger capture
func (s *CaptureSidecar) CaptureLogger(data []byte) {
	s.logger.Info(string(data),nil)
}

// ShouldBlock should return true if the sidecar should block
func (s *CaptureSidecar) ShouldBlock() bool {
	return s.block
}

// NewCaptureSidecarFromParams is the constructor
func NewCaptureSidecarFromParams(block bool, params map[string]interface{}) (*CaptureSidecar, error) {
	// We need a URI to send the captured data to. It can either be a file or a URL
	uri, ok := params["uri"].(string)
	if !ok  {
		return nil,errors.New("the uri of the capture sidecar is not of the right type")
	}
	// We need a regexp to identify if the request content type can be serialised
	reqRegex, ok := params["requestContentTypeRegex"].(string)
	if !ok {
		return nil,errors.New("the request regex of the capture sidecar is not of the right type")
	}
	// We also need a regexp to identify if the response content type can be serialied
	resRegex, ok := params["responseContentTypeRegex"].(string)
	if !ok {
		return nil,errors.New("the response regex of the capture sidecar is not of the right type")
	}
	return &CaptureSidecar{channel: make(chan *APIWrapper),uri: uri, block: block,
				requestContentTypeRegex: reqRegex,
				responseContentTypeRegex: resRegex},nil
}
