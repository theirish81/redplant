package main

import (
	"github.com/sirupsen/logrus"
	"reflect"
	"testing"
)

func TestNewLogHelperFromConfig(t *testing.T) {
	cfg := LoggerConfig{}
	cfg.Path = "foo.log"
	cfg.Level = "info"
	cfg.Format = "JSON"
	logger := NewLogHelperFromConfig(cfg)
	if logger.logger.Level != logrus.InfoLevel {
		t.Error("log level not preserved")
	}
	if reflect.TypeOf(logger.logger.Formatter).String() != "*logrus.JSONFormatter" {
		t.Error("formatter not preserved")
	}
}

func TestNewSTLogHelper(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	cfg := STLogConfig{Level: "debug"}
	logger := NewSTLogHelper(&cfg)
	if logger.logger.Level != logrus.DebugLevel {
		t.Error("log level not preserved")
	}
}
