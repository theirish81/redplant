package main

import (
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strings"
)

// LogHelper is wrapper for logrus. It exists to simplify the logging of metadata
type LogHelper struct {
	logger *logrus.Logger
	path   string
	format string
	level  string
}

func (h *LogHelper) extractCaller() string {
	_, file, _, _ := runtime.Caller(3)
	return file[strings.LastIndex(file, "/")+1:]
}

func (h *LogHelper) wrapperToMap(wrapper *APIWrapper) AnyMap {
	data := AnyMap{}
	data["remote_addr"] = wrapper.Request.RemoteAddr
	data["url"] = wrapper.Request.URL.String()
	data["method"] = wrapper.Request.Method
	data["tags"] = wrapper.Tags
	data["id"] = wrapper.ID
	if wrapper.Username != "" {
		data["username"] = wrapper.Username
	}
	if wrapper.Response != nil {
		data["status"] = wrapper.Response.Status
		data["tags"] = wrapper.Tags
	}
	data["module"] = h.extractCaller()
	return data
}

func (h *LogHelper) adjustMeta(meta map[string]any) map[string]any {
	if meta == nil {
		meta = map[string]any{}
	}
	meta["component"] = "redplant"
	if _, ok := meta["module"]; !ok {
		meta["module"] = h.extractCaller()
	}
	return meta
}

func (h *LogHelper) Log(message string, wrapper *APIWrapper, fn func(message string, meta map[string]any)) {
	fn(message, h.wrapperToMap(wrapper))
}

func (h *LogHelper) LogErr(message string, err error, wrapper *APIWrapper, fn func(message string, err error, meta map[string]any)) {
	fn(message, err, h.wrapperToMap(wrapper))
}

func (h *LogHelper) Debug(message string, meta map[string]any) {
	h.logger.WithFields(h.adjustMeta(meta)).Debug(message)
}

func (h *LogHelper) Info(message string, meta map[string]any) {
	h.logger.WithFields(h.adjustMeta(meta)).Info(message)
}

func (h *LogHelper) Warn(message string, err error, meta map[string]any) {
	meta = h.adjustMeta(meta)
	meta["error"] = err
	h.logger.WithFields(meta).Warn(message)
}

func (h *LogHelper) Error(message string, err error, meta map[string]any) {
	meta = h.adjustMeta(meta)
	meta["error"] = err
	h.logger.WithFields(meta).Error(message)
}

func (h *LogHelper) Fatal(message string, err error, meta map[string]any) {
	meta = h.adjustMeta(meta)
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
	return &LogHelper{lx, path, "JSON", level.String()}
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
	return &LogHelper{lx, cfg.Path, cfg.Format, cfg.Level}
}

type STLogHelper struct {
	LogHelper
}

func NewSTLogHelper(config *STLogConfig) *STLogHelper {
	if config != nil {
		return &STLogHelper{LogHelper: *NewLogHelperFromConfig(LoggerConfig{Path: config.Path, Level: config.Level, Format: config.Format})}
	} else {
		return &STLogHelper{LogHelper: *log}
	}
}
