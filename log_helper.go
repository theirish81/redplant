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

// extractCaller extracts the caller module for the log call
func (h *LogHelper) extractCaller() string {
	_, file, _, _ := runtime.Caller(3)
	return file[strings.LastIndex(file, "/")+1:]
}

// wrapperToMap extracts meaningful metadata from a wrapper for logging purposes
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

// adjustMeta will make sure meta exists, add the component and the module field and return the updated meta
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

// addMeta will add `more` fields to meta, and return the result
func (h *LogHelper) addMeta(meta map[string]any, more map[string]any) map[string]any {
	for k, v := range more {
		meta[k] = v
	}
	return meta
}

// Log creates a log entry
func (h *LogHelper) Log(message string, wrapper *APIWrapper, fn func(message string, meta map[string]any)) {
	fn(message, h.wrapperToMap(wrapper))
}

// LogWithMeta creates a log entry with meta
func (h *LogHelper) LogWithMeta(message string, wrapper *APIWrapper, meta map[string]any, fn func(message string, meta map[string]any)) {
	fn(message, h.addMeta(h.wrapperToMap(wrapper), meta))
}

// LogWithErrorMeta creates a log entry with meta and error
func (h *LogHelper) LogWithErrorMeta(message string, err error, wrapper *APIWrapper, meta map[string]any, fn func(message string, err error, meta map[string]any)) {
	fn(message, err, h.addMeta(h.wrapperToMap(wrapper), meta))
}

// LogErr creates a log entry with error
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

// NewLogHelper creates a new root logger
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

// NewLogHelperFromConfig creates a new root logger from configuration
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

// STLogHelper is a logger specialised for sidecars and transformers. It also take scare of publishing prometheus
// metrics for sidecars and transformers
type STLogHelper struct {
	LogHelper
	prometheusEnabled bool
	prometheusPrefix  string
}

// PrometheusRegisterCounter registers a Prometheus counter with the given name
func (l *STLogHelper) PrometheusRegisterCounter(name string) {
	if prom != nil && l.prometheusEnabled {
		prom.CustomCounter(l.prometheusPrefix + name)
	}
}

// PrometheusRegisterSummary registers a Prometheus summary with the given name
func (l *STLogHelper) PrometheusRegisterSummary(name string) {
	if prom != nil && l.prometheusEnabled {
		prom.CustomSummary(l.prometheusPrefix + name)
	}
}

// PrometheusCounterAdd triggers an `Add` for the counter with the given name
func (l *STLogHelper) PrometheusCounterAdd(name string, inc int) {
	if prom != nil {
		prom.CustomCounter(l.prometheusPrefix + name).Add(float64(inc))
	}
}

// PrometheusCounterInc triggers an `Inc` for the counter with the given name
func (l *STLogHelper) PrometheusCounterInc(name string) {
	if prom != nil {
		prom.CustomCounter(l.prometheusPrefix + name).Inc()
	}
}

// PrometheusSummaryObserve triggers an `Observe` for the summary with the given name
func (l *STLogHelper) PrometheusSummaryObserve(name string, data int64) {
	if prom != nil {
		prom.CustomSummary(l.prometheusPrefix + name).Observe(float64(data))
	}
}

// NewSTLogHelper creates an STLogHelper
func NewSTLogHelper(config *STLogConfig) *STLogHelper {
	helper := STLogHelper{LogHelper: *log}
	if config != nil {
		helper.LogHelper = *NewLogHelperFromConfig(LoggerConfig{Path: config.Path, Level: config.Level, Format: config.Format})
		if config.Prometheus.Enabled {
			helper.prometheusEnabled = true
			if len(config.Prometheus.Prefix) > 0 {
				helper.prometheusPrefix = config.Prometheus.Prefix + "_"
			}

		}
	}
	return &helper
}
