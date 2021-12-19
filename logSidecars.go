package main

import "github.com/sirupsen/logrus"

type RequestAccessLogSidecar struct {
	channel chan *APIWrapper
	log     *LogHelper
	block   bool
	Path    string
}

func (s *RequestAccessLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *RequestAccessLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				req := msg.Request
				s.log.Info("request access", map[string]interface{}{"remote_addr": req.RemoteAddr, "method": req.Method, "url": req.Host + req.URL.String()})
			}
		}()
	}
}

func (s *RequestAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewRequestAccessLogSidecarFromParams(block bool, params map[string]interface{}) (*RequestAccessLogSidecar, error) {
	logger := log
	sidecar := RequestAccessLogSidecar{channel: make(chan *APIWrapper), block: block}
	err := DecodeAndTempl(params, &sidecar, nil)
	if err != nil {
		return nil, err
	}
	if sidecar.Path != "" {
		logger = NewLogHelper(sidecar.Path, logrus.InfoLevel)
	}
	sidecar.log = logger
	return &sidecar, nil
}

type UpstreamAccessLogSidecar struct {
	channel chan *APIWrapper
	log     *LogHelper
	block   bool
	Path    string
}

func (s *UpstreamAccessLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *UpstreamAccessLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				res := msg.Response
				req := res.Request
				log.Info("upstream access", map[string]interface{}{"remote_addr": req.RemoteAddr, "method": req.Method, "url": req.URL.String(), "status": res.StatusCode})
			}
		}()
	}
}

func (s *UpstreamAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewUpstreamAccessLogSidecarFromParams(block bool, params map[string]interface{}) (*UpstreamAccessLogSidecar, error) {
	logger := log
	sidecar := UpstreamAccessLogSidecar{channel: make(chan *APIWrapper), block: block}
	err := DecodeAndTempl(params, &sidecar, nil)
	if err != nil {
		return nil, err
	}
	if sidecar.Path != "" {
		logger = NewLogHelper(sidecar.Path, logrus.InfoLevel)
	}
	sidecar.log = logger
	return &sidecar, err
}

type MetricsLogSidecar struct {
	channel chan *APIWrapper
	log     *LogHelper
	block   bool
	Path    string
}

func (s *MetricsLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *MetricsLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				log.Info("metrics", map[string]interface{}{"transaction": msg.Metrics.Transaction(), "req_transformation": msg.Metrics.ReqTransformation(), "res_transformation": msg.Metrics.ResTransformation()})
			}
		}()
	}
}

func (s *MetricsLogSidecar) ShouldBlock() bool {
	return s.block
}

func NewMetricsLogSidecarFromParams(block bool, params map[string]interface{}) (*MetricsLogSidecar, error) {
	logger := log
	sidecar := MetricsLogSidecar{channel: make(chan *APIWrapper), block: block}
	err := DecodeAndTempl(params, &sidecar, nil)
	if err != nil {
		return nil, err
	}
	if sidecar.Path != "" {
		logger = NewLogHelper(sidecar.Path, logrus.InfoLevel)
	}
	sidecar.log = logger
	return &sidecar, nil
}
