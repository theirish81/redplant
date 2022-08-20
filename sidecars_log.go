package main

// RequestAccessLogSidecar logs the inbound access requests
type RequestAccessLogSidecar struct {
	channel        chan *APIWrapper
	log            *STLogHelper
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
				s.log.Log("request access", msg, s.log.Info)
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

func NewRequestAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*RequestAccessLogSidecar, error) {
	sidecar := RequestAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	return &sidecar, nil
}

// UpstreamAccessLogSidecar logs the accesses to the upstream server, once the conversation has happened
type UpstreamAccessLogSidecar struct {
	channel        chan *APIWrapper
	log            *STLogHelper
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
				s.log.Log("upstream access", msg, s.log.Info)
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

func NewUpstreamAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*UpstreamAccessLogSidecar, error) {
	sidecar := UpstreamAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	return &sidecar, nil
}

type MetricsLogSidecar struct {
	channel          chan *APIWrapper
	log              *STLogHelper
	block            bool
	dropOnOverflow   bool
	Path             string
	PrometheusPrefix string
	Mode             string
	ActivateOnTags   []string
}

func (s *MetricsLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *MetricsLogSidecar) getPrometheusPrefix() string {
	if s.PrometheusPrefix == "" {
		return "metrics"
	}
	return "metrics_" + s.PrometheusPrefix
}

func (s *MetricsLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				if s.isPrometheusEnabled() {
					prom.CustomSummary(s.getPrometheusPrefix() + "_transaction").Observe(float64(msg.Metrics.Transaction()))
					prom.CustomSummary(s.getPrometheusPrefix() + "_req_transformation").Observe(float64(msg.Metrics.ReqTransformation()))
					prom.CustomSummary(s.getPrometheusPrefix() + "_res_transformation").Observe(float64(msg.Metrics.ResTransformation()))
				}
				if s.isTextEnabled() {
					s.log.Info("metrics", AnyMap{"transaction": msg.Metrics.Transaction(), "req_transformation": msg.Metrics.ReqTransformation(), "res_transformation": msg.Metrics.ResTransformation(), "tags": msg.Tags})
				}
			}
		}()
	}
}

func (s *MetricsLogSidecar) isPrometheusEnabled() bool {
	return prom != nil && (s.Mode == "prometheus" || s.Mode == "")
}

func (s *MetricsLogSidecar) isTextEnabled() bool {
	return s.Mode == "text" || s.Mode == ""
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

func NewMetricsLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*MetricsLogSidecar, error) {
	sidecar := MetricsLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	return &sidecar, nil
}
