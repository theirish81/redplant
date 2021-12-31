package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var log *LogHelper
var config Config

func main() {

	configFilePath := flag.String("c", "", "Path of the main configuration file")
	logFilePath := flag.String("l", "", "Path to the logging configuration file")
	flag.Parse()
	if *configFilePath == "" {
		fmt.Println("apifrp -c [config_file_path]")
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
	config.Init()

	router := SetupRouter()

	log.Info("Starting Server", map[string]interface{}{"port": config.Network.Downstream.Port})
	if config.Network.Downstream.Tls != nil {
		server := &http.Server{Addr: ":" + strconv.Itoa(config.Network.Downstream.Port), Handler: router, TLSConfig: setupTLSConfig()}
		log.Fatal("Stopping service", server.ListenAndServeTLS("", ""), nil)
	} else {
		log.Fatal("Stopping service", http.ListenAndServe(":"+strconv.Itoa(config.Network.Downstream.Port), router), nil)
	}
}

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
