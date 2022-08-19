package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

// LogHelper is wrapper for logrus. It exists to simplify the logging of metadata
type LogHelper struct {
	logger *logrus.Logger
}

func (h *LogHelper) wrapperToMap(wrapper *APIWrapper) AnyMap {
	data := AnyMap{}
	data["remote_addr"] = wrapper.Request.RemoteAddr
	data["url"] = wrapper.Request.URL.String()
	data["method"] = wrapper.Request.Method
	data["tags"] = wrapper.Tags

	if wrapper.Response != nil {
		data["status"] = wrapper.Response.Status
		data["tags"] = wrapper.Tags
	}
	return data
}

func (h *LogHelper) Log(message string, wrapper *APIWrapper, fn func(message string, meta map[string]any)) {
	fn(message, h.wrapperToMap(wrapper))
}

func (h *LogHelper) Info(message string, meta map[string]any) {
	h.logger.WithFields(meta).Info(message)
}

func (h *LogHelper) Warn(message string, err error, meta map[string]any) {
	if meta == nil {
		meta = logrus.Fields{}
	}
	meta["error"] = err
	h.logger.WithFields(meta).Warn(message)
}

func (h *LogHelper) Error(message string, err error, meta map[string]any) {
	if meta == nil {
		meta = logrus.Fields{}
	}
	meta["error"] = err
	h.logger.WithFields(meta).Error(message)
}

func (h *LogHelper) Fatal(message string, err error, meta map[string]any) {
	if meta == nil {
		meta = logrus.Fields{}
	}
	meta["error"] = err
	h.logger.WithFields(meta).Fatal(message)
}

func NewLogHelper(path string, level logrus.Level) *LogHelper {
	lx := logrus.New()
	lx.SetFormatter(&logrus.JSONFormatter{})
	if path == "" {
		lx.SetOutput(os.Stdout)
	} else {
		file, _ := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
		lx.SetOutput(file)
	}
	lx.SetLevel(level)
	return &LogHelper{lx}
}

func NewLogHelperFromConfig(cfg LoggerConfig) *LogHelper {
	lx := logrus.New()
	if cfg.Format == "JSON" {
		lx.SetFormatter(&logrus.JSONFormatter{})
	}
	if cfg.Path == "" {
		lx.SetOutput(os.Stdout)
	} else {
		file, _ := os.OpenFile(cfg.Path, os.O_RDWR|os.O_CREATE, 0755)
		lx.SetOutput(file)
	}
	switch strings.ToLower(cfg.Level) {
	case "debug":
		lx.SetLevel(logrus.DebugLevel)
	default:
		lx.SetLevel(logrus.InfoLevel)
	case "warn":
		lx.SetLevel(logrus.WarnLevel)
	case "error":
		lx.SetLevel(logrus.ErrorLevel)
	case "fatal":
		lx.SetLevel(logrus.FatalLevel)
	}
	return &LogHelper{lx}
}
