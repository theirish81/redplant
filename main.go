package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
)

var log *LogHelper
var config Config

func main() {

	configFilePath := flag.String("c", "", "Config file path")
	flag.Parse()
	if *configFilePath == "" {
		fmt.Println("apifrp -c [config_file_path]")
		flag.PrintDefaults()
		return
	}

	log = NewLogHelper("", logrus.InfoLevel)

	config = LoadConfig(*configFilePath)
	config.Init()

	router := SetupRouter()

	log.Info("Starting Server", map[string]interface{}{"port": config.Network.Downstream.Port})
	if config.Network.Downstream.Tls != nil {
		log.Fatal("Stopping service", http.ListenAndServeTLS(":"+strconv.Itoa(config.Network.Downstream.Port), config.Network.Downstream.Tls.Cert, config.Network.Downstream.Tls.Key, router), nil)
	} else {
		log.Fatal("Stopping service", http.ListenAndServe(":"+strconv.Itoa(config.Network.Downstream.Port), router), nil)
	}

}
