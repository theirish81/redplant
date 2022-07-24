package main

// ISidecar is the interface for all sidecars
// Consume will start consuming the messages. It receives an int as parameter that determines how many instances
// of the `consume` go-routines should be launched
// GetChannel will return the inbound channel for the sidecar
// ShouldBlock will return true if the sidecar should block or should be completely asynchronous
// ShouldDropOnOverflow will return true if the inbound messages should be dropped if the channel is full
// ShouldExpandRequest will return true if the request body stream should be expanded (turned into bytes). In other words,
// it will need to return true if the sidecar makes any use of the request body
// ShouldExpandResponse will return true if the response body stream should be expanded (turned into bytes). In other words,
// it will need to return true if the sidecar makes any use of the response body
// IsActive will return true if the sidecar is interested in this message
type ISidecar interface {
	Consume(consumers int)
	GetChannel() chan *APIWrapper
	ShouldBlock() bool
	ShouldDropOnOverflow() bool
	ShouldExpandRequest() bool
	ShouldExpandResponse() bool
	IsActive(wrapper *APIWrapper) bool
}

// RequestSidecars is a collection of sidecars
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

// Push will push a new sidecar to the list of sidecars
func (s *RequestSidecars) Push(sidecar ISidecar) {
	s.sidecars = append(s.sidecars, sidecar)
}

// Run will run all the sidecars against the provided wrapper
func (s *RequestSidecars) Run(wrapper *APIWrapper) {
	for _, sidecar := range s.sidecars {
		// If the sidecar is "active", then run it. "active" means that there's either no "activateOnTags"
		// or the "activateOnTags" matches the tags in the wrapper
		if sidecar.IsActive(wrapper) {
			// If this is meant to be a blocking sidecar, we just run the function
			if sidecar.ShouldBlock() {
				s.runFunc(sidecar, wrapper)
				// Otherwise, we'll just push the message in another goroutine
			} else {
				go s.runFunc(sidecar, wrapper)
			}
		}
	}
}

// runFunc will attempt to send a message to the given sidecar
func (s *RequestSidecars) runFunc(sidecar ISidecar, wrapper *APIWrapper) {
	// Send the message. If the queue is full, drop it
	if sidecar.ShouldDropOnOverflow() {
		select {
		case sidecar.GetChannel() <- wrapper:
		default:
		}
	} else {
		// Send the message. If the queue is full, block
		sidecar.GetChannel() <- wrapper
	}
}

// NewRequestSidecars is the constructor of RequestSidecars
func NewRequestSidecars(sidecars *[]SidecarConfig) *RequestSidecars {
	res := RequestSidecars{}
	for _, s := range *sidecars {
		if s.Workers == 0 {
			s.Workers = 1
		}
		if s.Queue == 0 {
			s.Queue = 1
		}
		switch s.Id {
		case "accessLog":
			sidecar, err := NewRequestAccessLogSidecarFromParams(s.Block, s.Queue, s.DropOnOverflow, s.ActivateOnTags, s.Params)
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

// ResponseSidecars is a collection of response sidecars
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

// Push adds a sidecar to the list of sidecars
func (s *ResponseSidecars) Push(sidecar ISidecar) {
	s.sidecars = append(s.sidecars, sidecar)
}

// Run runs all the sidecars for the given wrapper
func (s *ResponseSidecars) Run(wrapper *APIWrapper) {
	// for every sidecar...
	for _, sidecar := range s.sidecars {
		// If sidecar is active, which means either activateOnTags is empty, or there's a match between
		// activateOnTags and the tags in the wrapper....
		if sidecar.IsActive(wrapper) {
			// If the sidecar should block the transaction in case of an overflowing queue...
			if sidecar.ShouldBlock() {
				s.runFunc(sidecar, wrapper)
			} else {
				// If the sidecar should never block the transaction in case of an overflowing queue...
				go s.runFunc(sidecar, wrapper)
			}
		}
	}
}

// runFunc will attempt to send a message to the given sidecar
func (s *ResponseSidecars) runFunc(sidecar ISidecar, wrapper *APIWrapper) {
	// Send the message. If the queue is full, drop it
	if sidecar.ShouldDropOnOverflow() {
		select {
		case sidecar.GetChannel() <- wrapper:
		default:
		}
	} else {
		// Send the message. If the queue is full, block
		sidecar.GetChannel() <- wrapper
	}
}

// NewResponseSidecars is the constructor of ResponseSidecars
func NewResponseSidecars(sidecars *[]SidecarConfig) *ResponseSidecars {
	res := ResponseSidecars{}
	for _, s := range *sidecars {
		if s.Workers == 0 {
			s.Workers = 1
		}
		if s.Queue == 0 {
			s.Queue = 1
		}
		switch s.Id {
		case "accessLog":
			sidecar, err := NewUpstreamAccessLogSidecarFromParams(s.Block, s.Queue, s.DropOnOverflow, s.ActivateOnTags, s.Params)
			if err != nil {
				log.Error("Could not initialize upstream access log", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}

		case "metricsLog":
			sidecar, err := NewMetricsLogSidecarFromParams(s.Block, s.Queue, s.DropOnOverflow, s.ActivateOnTags, s.Params)
			if err != nil {
				log.Error("Could not initialize metrics log", err, nil)
			} else {
				sidecar.Consume(s.Workers)
				res.Push(sidecar)
			}

		case "capture":
			sidecar, err := NewCaptureSidecarFromParams(s.Block, s.Queue, s.DropOnOverflow, s.ActivateOnTags, s.Params)
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
