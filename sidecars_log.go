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
				s.log.PrometheusCounterInc("request_access")
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

// NewRequestAccessLogSidecarFromParams constructor for RequestAccessLogSidecar from params
func NewRequestAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*RequestAccessLogSidecar, error) {
	sidecar := RequestAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	sidecar.log.PrometheusCounterInc("request_access")
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
				s.log.PrometheusCounterInc("upstream_access")
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

// NewUpstreamAccessLogSidecarFromParams creates an UpstreamAccessLogSidecar from params
func NewUpstreamAccessLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*UpstreamAccessLogSidecar, error) {
	sidecar := UpstreamAccessLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	sidecar.log.PrometheusRegisterCounter("upstream_access")
	return &sidecar, nil
}

// MetricsLogSidecar a sidecar to log metrics
type MetricsLogSidecar struct {
	channel        chan *APIWrapper
	log            *STLogHelper
	block          bool
	dropOnOverflow bool
	Path           string
	Mode           string
	ActivateOnTags []string
}

func (s *MetricsLogSidecar) GetChannel() chan *APIWrapper {
	return s.channel
}

func (s *MetricsLogSidecar) Consume(quantity int) {
	for i := 0; i < quantity; i++ {
		go func() {
			for msg := range s.GetChannel() {
				s.log.PrometheusSummaryObserve("transaction", msg.Metrics.Transaction())
				s.log.PrometheusSummaryObserve("req_transformation", msg.Metrics.ReqTransformation())
				s.log.PrometheusSummaryObserve("res_transformation", msg.Metrics.ResTransformation())
				s.log.LogWithMeta("metrics", msg, AnyMap{"transaction": msg.Metrics.Transaction(), "req_transformation": msg.Metrics.ReqTransformation(), "res_transformation": msg.Metrics.ResTransformation(), "tags": msg.Tags}, s.log.Info)
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

// NewMetricsLogSidecarFromParams creates a MetricsLogSidecar from params
func NewMetricsLogSidecarFromParams(block bool, queue int, dropOnOverflow bool, activateOnTags []string, logCfg *STLogConfig, _ AnyMap) (*MetricsLogSidecar, error) {
	sidecar := MetricsLogSidecar{channel: make(chan *APIWrapper, queue), block: block, dropOnOverflow: dropOnOverflow, ActivateOnTags: activateOnTags}
	sidecar.log = NewSTLogHelper(logCfg)
	sidecar.log.PrometheusRegisterSummary("transaction")
	sidecar.log.PrometheusRegisterSummary("req_transformation")
	sidecar.log.PrometheusRegisterSummary("res_transformation")
	return &sidecar, nil
}
