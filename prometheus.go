package main

import "github.com/prometheus/client_golang/prometheus"

// Prometheus is the RedPlant configuration for Prometheus
// GlobalInboundRequestsCounter is the counter for all inbound requests
// GlobalOriginRequestsCounter is the counter for all the origin hits
// CustomCounters is a map of counters transformers and sidecars can use
// CustomSummaries is a map of summaries transformers and sidecars can use
type Prometheus struct {
	GlobalInboundRequestsCounter prometheus.Counter
	GlobalOriginRequestsCounter  prometheus.Counter
	InternalErrorsCounter        prometheus.Counter
	CustomCounters               map[string]prometheus.Counter
	CustomSummaries              map[string]prometheus.Summary
}

// NewPrometheus is the Prometheus constructor
func NewPrometheus() *Prometheus {
	prom := Prometheus{}
	grc := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "redplant",
		Name:      "global_inbound_requests",
		Help:      "global inbound request counter",
	})
	prom.GlobalInboundRequestsCounter = grc
	_ = prometheus.Register(grc)
	gorc := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "redplant",
		Name:      "global_origin_requests",
		Help:      "global requests to the origin counter",
	})
	prom.GlobalOriginRequestsCounter = gorc
	_ = prometheus.Register(gorc)

	iec := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "redplant",
		Name:      "internal_errors",
		Help:      "counting unhandled, unexpected internal errors",
	})
	prom.InternalErrorsCounter = iec
	_ = prometheus.Register(iec)

	prom.CustomCounters = make(map[string]prometheus.Counter)
	prom.CustomSummaries = make(map[string]prometheus.Summary)

	return &prom
}

// CustomCounter will return a prometheus.Counter instance for the given name. If the counter does not already exist
// it will create one
func (p *Prometheus) CustomCounter(name string) prometheus.Counter {
	if counter, ok := p.CustomCounters[name]; ok {
		return counter
	}
	p.CustomCounters[name] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "redplant",
		Name:      name,
	})
	_ = prometheus.Register(p.CustomCounters[name])
	return p.CustomCounters[name]
}

// CustomSummary will return a prometheus.Summary instance for the given name. If the counter does not already exist
// it will create one
func (p *Prometheus) CustomSummary(name string) prometheus.Summary {
	if summary, ok := p.CustomSummaries[name]; ok {
		return summary
	}
	p.CustomSummaries[name] = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "redplant",
		Name:      name,
	})
	_ = prometheus.Register(p.CustomSummaries[name])
	return p.CustomSummaries[name]
}
