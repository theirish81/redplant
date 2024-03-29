package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var log *LogHelper
var config Config
var addresser = NewIPAddresser()
var prom *Prometheus
var template RPTemplate

func main() {
	configFilePath := flag.String("c", "", "Path of the main configuration file")
	logFilePath := flag.String("l", "", "Path to the logging configuration file")
	flag.Parse()
	if *configFilePath == "" {
		fmt.Println("redplant -c [config_file_path]")
		flag.PrintDefaults()
		return
	}

	loggingConfig, err := LoadLoggerConfig(logFilePath)
	if err != nil {
		fmt.Println("Could not load logging configuration")
		os.Exit(1)
		return
	}
	log = NewLogHelperFromConfig(loggingConfig)
	config = LoadConfig(*configFilePath)
	startPrometheus()
	config.Init()
	template = NewRPTemplate()
	router := SetupRouter()

	log.Info("Starting Server", AnyMap{"port": config.Network.Downstream.Port})

	server := &http.Server{Addr: ":" + strconv.Itoa(config.Network.Downstream.Port), Handler: router, TLSConfig: setupTLSConfig()}
	if config.Network.Downstream.Tls != nil {
		go func() {
			err = server.ListenAndServeTLS("", "")
			if err.Error() != "http: Server closed" {
				log.Fatal("Error while running web server", err, nil)
			}
		}()
		handleTerm(server)
	} else {
		go func() {
			err = server.ListenAndServe()
			if err.Error() != "http: Server closed" {
				log.Fatal("Error while running web server", err, nil)
			}
		}()
		handleTerm(server)
	}
}

// startPrometheus initializes and starts the Prometheus monitoring functionality
func startPrometheus() {
	if config.Prometheus != nil {
		prom = NewPrometheus()
		go func() {
			promRouter := mux.NewRouter()
			promRouter.Handle(config.Prometheus.Path, promhttp.Handler())
			_ = http.ListenAndServe(":"+strconv.Itoa(config.Prometheus.Port), promRouter)
		}()
	}
}

// setupTLSConfig will set up the TLS specifics if Network.Downstream.Tls was configured
func setupTLSConfig() *tls.Config {
	cfg := &tls.Config{}
	if config.Network.Downstream.Tls != nil {
		for _, certConfig := range config.Network.Downstream.Tls {
			cert, err := tls.LoadX509KeyPair(certConfig.Cert, certConfig.Key)
			if err != nil {
				log.Fatal("Could not load TLS cert", err, nil)
			}
			cfg.Certificates = append(cfg.Certificates, cert)
		}
	}
	return cfg
}

// handleTerm will listen to the termination signals. When a signal is captured, then it will wait 10s of grace period
// before shutting down completely
func handleTerm(server *http.Server) {
	signalChannel := make(chan os.Signal, 1)
	exitChan := make(chan int)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		for {
			<-signalChannel
			log.Info("Graceful shutdown initiated", nil)
			err := server.Shutdown(context.Background())
			if err != nil {
				log.Error("Error while shutting down web server", err, nil)
			}
			duration, _ := time.ParseDuration("10s")
			time.Sleep(duration)
			log.Info("Graceful shutdown completed", nil)
			exitChan <- 0
		}
	}()
	exitCode := <-exitChan
	os.Exit(exitCode)

}
