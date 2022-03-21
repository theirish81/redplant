package main

import "github.com/sirupsen/logrus"

// RequestAccessLogSidecar logs the inbound access requests
type RequestAccessLogSidecar struct {
	channel        chan *APIWrapper
	log            *LogHelper
	block          bool
	dropOnOverflow bool
	Path           string
	ActivateOnTags []string
}

func (s *RequestAccessLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *RequestAccessLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				req := msg.Request
				s.log.Info("request access", logrus.Fields{"remote_addr": req.RemoteAddr, "method": req.Method, "url": req.Host + req.URL.String(), "tags": msg.Tags})
			}
		}()
	}
}

func (s *RequestAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func (s *RequestAccessLogSidecar) ShouldDropOnOverflow() bool {
	return s.dropOnOverflow
}

func (s *RequestAccessLogSidecar) ShouldExpandRequest() bool {
	return false
}
func (s *RequestAccessLogSidecar) ShouldExpandResponse() bool {
	return false
}

func (s *RequestAccessLogSidecar) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(s.ActivateOnTags)
}

func NewRequestAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, params map[string]interface{}) (*RequestAccessLogSidecar, error) {
	logger := log
	sidecar := RequestAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &sidecar, nil, []string{})
	if err != nil {
		return nil, err
	}
	if sidecar.Path != "" {
		logger = NewLogHelper(sidecar.Path, logrus.InfoLevel)
	}
	sidecar.log = logger
	return &sidecar, nil
}

// UpstreamAccessLogSidecar logs the accesses to the upstream server, once the conversation has happened
type UpstreamAccessLogSidecar struct {
	channel        chan *APIWrapper
	log            *LogHelper
	block          bool
	dropOnOverflow bool
	Path           string
	ActivateOnTags []string
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
				log.Info("upstream access", logrus.Fields{"remote_addr": req.RemoteAddr, "method": req.Method, "url": req.URL.String(), "status": res.StatusCode, "tags": msg.Tags})
			}
		}()
	}
}

func (s *UpstreamAccessLogSidecar) ShouldBlock() bool {
	return s.block
}

func (s *UpstreamAccessLogSidecar) ShouldDropOnOverflow() bool {
	return s.dropOnOverflow
}

func (s *UpstreamAccessLogSidecar) ShouldExpandRequest() bool {
	return false
}
func (s *UpstreamAccessLogSidecar) ShouldExpandResponse() bool {
	return false
}

func (s *UpstreamAccessLogSidecar) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(s.ActivateOnTags)
}

func NewUpstreamAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, params map[string]interface{}) (*UpstreamAccessLogSidecar, error) {
	logger := log
	sidecar := UpstreamAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &sidecar, nil, []string{})
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
	channel          chan *APIWrapper
	log              *LogHelper
	block            bool
	dropOnOverflow   bool
	Path             string
	PrometheusPrefix string
	ActivateOnTags   []string
}

func (s *MetricsLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *MetricsLogSidecar) getPrometheusPrefix() string {
	if s.PrometheusPrefix == "" {
		return "metrics"
	}
	return s.PrometheusPrefix
}

func (s *MetricsLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				if prom != nil {
					prom.CustomSummary(s.getPrometheusPrefix() + "_transaction").Observe(float64(msg.Metrics.Transaction()))
					prom.CustomSummary(s.getPrometheusPrefix() + "_req_transformation").Observe(float64(msg.Metrics.ReqTransformation()))
					prom.CustomSummary(s.getPrometheusPrefix() + "_res_transformation").Observe(float64(msg.Metrics.ResTransformation()))
				}
				log.Info("metrics", logrus.Fields{"transaction": msg.Metrics.Transaction(), "req_transformation": msg.Metrics.ReqTransformation(), "res_transformation": msg.Metrics.ResTransformation(), "tags": msg.Tags})
			}
		}()
	}
}

func (s *MetricsLogSidecar) ShouldBlock() bool {
	return s.block
}

func (s *MetricsLogSidecar) ShouldDropOnOverflow() bool {
	return s.dropOnOverflow
}

func (s *MetricsLogSidecar) ShouldExpandRequest() bool {
	return false
}
func (s *MetricsLogSidecar) ShouldExpandResponse() bool {
	return false
}

func (s *MetricsLogSidecar) IsActive(wrapper *APIWrapper) bool {
	return wrapper.HasTag(s.ActivateOnTags)
}

func NewMetricsLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, params map[string]interface{}) (*MetricsLogSidecar, error) {
	logger := log
	sidecar := MetricsLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	err := DecodeAndTempl(params, &sidecar, nil, []string{})
	if err != nil {
		return nil, err
	}
	if sidecar.Path != "" {
		logger = NewLogHelper(sidecar.Path, logrus.InfoLevel)
	}
	sidecar.log = logger
	return &sidecar, nil
}
