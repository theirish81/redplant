package main

import "github.com/sirupsen/logrus"

type RequestAccessLogSidecar struct {
	channel	chan *APIWrapper
	log		*LogHelper
	block	bool
}

func (s *RequestAccessLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *RequestAccessLogSidecar) Consume(quantity int) {
	for i:=0;i<quantity;i++ {
		go func() {
			for msg := range s.GetChannel() {
				req := msg.Request
				s.log.Info("request access", map[string]interface{}{"remote_addr":req.RemoteAddr, "method": req.Method,"url":req.Host+req.URL.String()})
			}
		}()
	}
}

func (s *RequestAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewRequestAccessLogSidecarFromParams(block bool, params map[string]interface{}) *RequestAccessLogSidecar {
	logger := log
	if path,ok := params["path"]; ok  {
		logger = NewLogHelper(path.(string),logrus.InfoLevel)
	}
	return &RequestAccessLogSidecar{make(chan *APIWrapper),logger,block}
}

type UpstreamAccessLogSidecar struct {
	channel chan *APIWrapper
	log		*LogHelper
	block	bool
}

func (s *UpstreamAccessLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *UpstreamAccessLogSidecar) Consume(quantity int) {
	for i:=0;i<quantity;i++ {
		go func() {
			for msg := range s.GetChannel() {
				res := msg.Response
				req := res.Request
				log.Info("upstream access", map[string]interface{}{"remote_addr":req.RemoteAddr, "method":req.Method, "url":req.URL.String(), "status":res.StatusCode})
			}
		}()
	}
}

func (s *UpstreamAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewUpstreamAccessLogSidecarFromParams(block bool, params map[string]interface{}) *UpstreamAccessLogSidecar {
	logger := log
	if path,ok := params["path"]; ok  {
		logger = NewLogHelper(path.(string),logrus.InfoLevel)
	}
	return &UpstreamAccessLogSidecar{make(chan *APIWrapper), logger, block}
}

type MetricsLogSidecar struct {
	channel chan *APIWrapper
	log		*LogHelper
	block	bool
}

func (s *MetricsLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *MetricsLogSidecar) Consume(quantity int) {
	for i:=0;i<quantity;i++ {
		go func() {
			for msg := range s.GetChannel() {
				log.Info("metrics",map[string]interface{}{"transaction":msg.Metrics.Transaction(),"req_transformation":msg.Metrics.ReqTransformation(), "res_transformation":msg.Metrics.ResTransformation()})
			}
		}()
	}
}

func (s *MetricsLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewMetricsLogSidecarFromParams(block bool, params map[string]interface{}) *MetricsLogSidecar {
	logger := log
	if path,ok := params["path"]; ok  {
		logger = NewLogHelper(path.(string),logrus.InfoLevel)
	}
	return &MetricsLogSidecar{make(chan *APIWrapper), logger, block}
}