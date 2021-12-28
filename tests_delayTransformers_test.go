package main

import (
	"testing"
	"time"
)

func TestDelayTransformer_Transform(t *testing.T) {
	transformer, _ := NewDelayTransformer([]string{}, map[string]interface{}{"min": "1s", "max": "3s"})
	wrapper := APIWrapper{}
	before := time.Now()
	_, _ = transformer.Transform(&wrapper)
	after := time.Now()
	if after.Sub(before).Seconds() > 3 || after.Sub(before).Seconds() < 1 {
		t.Error("Delay is not working as expected")
	}
}
