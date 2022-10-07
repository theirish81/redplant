package main

import "testing"

func TestPrometheus(t *testing.T) {
	config = Config{Prometheus: &PrometheusConfig{Namespace: "redplant"}}
	p := NewPrometheus()
	if p.CustomSummaries == nil || p.CustomCounters == nil || p.InternalErrorsCounter == nil {
		t.Error("prometheus init did not work")
	}
	c := p.CustomCounter("foo")
	c.Inc()
	if _, ok := p.CustomCounters["foo"]; !ok {
		t.Error("custom counter creation not right")
	}

	s := p.CustomSummary("foos")
	s.Observe(22)
	if _, ok := p.CustomSummaries["foos"]; !ok {
		t.Error("custom summary creation not right")
	}

}
