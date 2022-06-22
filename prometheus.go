package main

import "github.com/prometheus/client_golang/prometheus"

type Prometheus struct {
	GlobalInboundRequestsCounter prometheus.Counter
	GlobalOriginRequestsCounter  prometheus.Counter
	InternalErrorsCounter        prometheus.Counter
	CustomCounters               map[string]prometheus.Counter
	CustomSummaries              map[string]prometheus.Summary
}

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
