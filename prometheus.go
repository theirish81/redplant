package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
)

// Prometheus is the RedPlant configuration for Prometheus
// InternalErrorsCounter is a global Prometheus counter for errors
// CustomCounters is a map of counters transformers and sidecars can use
// CustomSummaries is a map of summaries transformers and sidecars can use
// customCounterCreationMutex will make sure that no duplicate counters will be created
// customSummaryCreationMutex will make sure that no duplicate summaries will be created
type Prometheus struct {
	InternalErrorsCounter      prometheus.Counter
	CustomCounters             map[string]prometheus.Counter
	CustomSummaries            map[string]prometheus.Summary
	customCounterCreationMutex sync.Mutex
	customSummaryCreationMutex sync.Mutex
}

// NewPrometheus is the Prometheus constructor
func NewPrometheus() *Prometheus {
	prom := Prometheus{}
	iec := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: config.Prometheus.Namespace,
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
	p.customCounterCreationMutex.Lock()
	defer p.customCounterCreationMutex.Unlock()
	if counter, ok := p.CustomCounters[name]; ok {
		return counter
	}
	p.CustomCounters[name] = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: config.Prometheus.Namespace,
		Name:      name,
	})
	_ = prometheus.Register(p.CustomCounters[name])
	return p.CustomCounters[name]
}

// CustomSummary will return a prometheus.Summary instance for the given name. If the counter does not already exist
// it will create one
func (p *Prometheus) CustomSummary(name string) prometheus.Summary {
	p.customSummaryCreationMutex.Lock()
	defer p.customSummaryCreationMutex.Unlock()
	if summary, ok := p.CustomSummaries[name]; ok {
		return summary
	}
	p.CustomSummaries[name] = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: config.Prometheus.Namespace,
		Name:      name,
	})
	_ = prometheus.Register(p.CustomSummaries[name])
	return p.CustomSummaries[name]
}
