package main

type ISidecar interface {
	Consume(consumers int)
	GetChannel() chan *APIWrapper
	ShouldBlock() bool
	ShouldExpandRequest() bool
	ShouldExpandResponse() bool
}

type RequestSidecars struct {
	sidecars []ISidecar
}

func (s *RequestSidecars) ShouldExpandRequest() bool {
	expand := false
	for _, sx := range s.sidecars {
		if sx.ShouldExpandRequest() {
			expand = true
		}
	}
	return expand
}

func (s *RequestSidecars) Push(sidecar ISidecar) {
	s.sidecars = append(s.sidecars, sidecar)
}
func (s *RequestSidecars) Run(wrapper *APIWrapper) {
	for _, sidecar := range s.sidecars {
		f := func() {
			sidecar.GetChannel() <- wrapper
		}
		if sidecar.ShouldBlock() {
			f()
		} else {
			go f()
		}

	}
}

func NewRequestSidecars(sidecars *[]SidecarConfig) *RequestSidecars {
	res := RequestSidecars{}
	for _, s := range *sidecars {
		switch s.Id {
		case "accessLog":
			sidecar, err := NewRequestAccessLogSidecarFromParams(s.Block, s.Params)
			if err != nil {
				log.Error("Could not initialise accessLog", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}

		}
	}
	return &res
}

type ResponseSidecars struct {
	sidecars []ISidecar
}

func (s *ResponseSidecars) ShouldExpandRequest() bool {
	expand := false
	for _, sx := range s.sidecars {
		if sx.ShouldExpandRequest() {
			expand = true
		}
	}
	return expand
}

func (s *ResponseSidecars) ShouldExpandResponse() bool {
	expand := false
	for _, sx := range s.sidecars {
		if sx.ShouldExpandResponse() {
			expand = true
		}
	}
	return expand
}

func (s *ResponseSidecars) Push(sidecar ISidecar) {
	s.sidecars = append(s.sidecars, sidecar)
}
func (s *ResponseSidecars) Run(wrapper *APIWrapper) {
	for _, sidecar := range s.sidecars {
		sidecar.GetChannel() <- wrapper
	}
}

func NewResponseSidecars(sidecars *[]SidecarConfig) *ResponseSidecars {
	res := ResponseSidecars{}
	for _, s := range *sidecars {
		switch s.Id {
		case "accessLog":
			sidecar, err := NewUpstreamAccessLogSidecarFromParams(s.Block, s.Params)
			if err != nil {
				log.Error("Could not initialize upstream access log", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}

		case "metricsLog":
			sidecar, err := NewMetricsLogSidecarFromParams(s.Block, s.Params)
			if err != nil {
				log.Error("Could not initialize metrics log", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}

		case "capture":
			sidecar, err := NewCaptureSidecarFromParams(s.Block, s.Params)
			if err != nil {
				log.Error("Could not initialize capture sidecar. Bypassing. ", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}
		}
	}
	return &res
}
