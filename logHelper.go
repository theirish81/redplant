package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

type LogHelper struct {
	logger *logrus.Logger

}

func (h *LogHelper) Info(message string, meta map[string]interface{} ){
	h.logger.WithFields(meta).Info(message)
}

func (h *LogHelper) Warn(message string, err error, meta map[string]interface{} ){
	meta["error"] = err
	h.logger.WithFields(meta).Warn(message)
}

func (h *LogHelper) Error(message string, err error, meta map[string]interface{} ){
	meta["error"] = err
	h.logger.WithFields(meta).Error(message)
}

func (h *LogHelper) Fatal(message string, err error, meta map[string]interface{} ){
	meta["error"] = err
	h.logger.WithFields(meta).Fatal(message)
}

func NewLogHelper(path string, level logrus.Level) *LogHelper {
	lx := logrus.New()
	lx.SetFormatter(&logrus.JSONFormatter{})
	if path == "" {
		lx.SetOutput(os.Stdout)
	}  else {
		file,_ := os.OpenFile(path,os.O_RDWR|os.O_CREATE, 0755)
		lx.SetOutput(file)
	}
	lx.SetLevel(level)
	return &LogHelper{lx}
}