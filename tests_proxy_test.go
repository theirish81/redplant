package main

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSetupRouter(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	config = LoadConfig("etc/config.yaml")
	config.Init()
	router := SetupRouter()
	route := router.GetRoute("localhost:9001")
	if route == nil {
		t.Error("Route not properly registered")
	}
}
func TestReverseProxy(t *testing.T) {
	log = NewLogHelper("", logrus.InfoLevel)
	config = LoadConfig("etc/config.yaml")
	config.Init()
	router := SetupRouter()
	go func() {
		_ = http.ListenAndServe(":"+strconv.Itoa(config.Network.Downstream.Port), router)
	}()

	d, _ := time.ParseDuration("2s")
	time.Sleep(d)
	res, _ := http.Get("http://foo:bar@localhost:9001/todo/1")
	if !strings.Contains(res.Header.Get("Content-Type"), "application/json") {
		t.Error("Something went wrong with dry run")
	}
	res, _ = http.Get("http://localhost:9001/todo/1")
	if res.StatusCode != 401 {
		t.Error("Something went wrong with dry run")
	}
}
