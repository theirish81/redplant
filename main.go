package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"
)

var log *LogHelper
var config Config
var addresser *IPAddresser

func main() {

	configFilePath := flag.String("c", "", "Config file path")
	flag.Parse()
	if *configFilePath == "" {
		fmt.Println("apifrp -c [config_file_path]")
		flag.PrintDefaults()
		return
	}

	log = NewLogHelper("", logrus.InfoLevel)

	addresser = newIPAddresser()

	router := mux.NewRouter()

	config = LoadConfig(*configFilePath)
	config.Init()

	reverseProxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
		wrapper := GetWrapper(req)
		if wrapper == nil {
			return
		}
		wrapper.Request = req
		wrapper.ExpandRequest()
		wrapper.Rule.Request._sidecars.Run(wrapper.Clone())
		handleURL(wrapper.Rule, req)
		wrapper.Metrics.ReqTransStart = time.Now()
		_, err := wrapper.Rule.Request._transformers.Transform(wrapper)
		wrapper.Metrics.ReqTransEnd = time.Now()
		wrapper.Err = err
	},
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			switch err.Error() {
			case "no_mapping":
				writer.WriteHeader(404)
			case "no_auth":
				writer.WriteHeader(401)
			case "signature is invalid":
				writer.WriteHeader(401)
			case "barraged":
				writer.WriteHeader(403)
			default:
				log.Error("Error while serving resource", err, map[string]interface{}{"url": request.URL.String()})
				writer.WriteHeader(500)
			}
		},
		Transport: configTransport(),
		ModifyResponse: func(response *http.Response) error {
			wrapper := GetWrapper(response.Request)
			if wrapper != nil {
				wrapper := GetWrapper(response.Request)
				wrapper.Response = response
				wrapper.ExpandResponse()
				wrapper.Metrics.ResTransStart = time.Now()
				_, err := wrapper.Rule.Response._transformers.Transform(wrapper)
				if err != nil {
					return err
				}
				wrapper.Metrics.ResTransEnd = time.Now()
				wrapper.Metrics.TransactionEnd = time.Now()
				wrapper.Rule.Response._sidecars.Run(wrapper)
			}
			return nil
		},
	}

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if rules, ok := config.Rules[req.Host]; ok {
			for _, rule := range rules {
				if success := rule._pattern.MatchString(req.URL.Path); success {
					req = ReqWithContext(req, rule)
					break
				}
			}
			reverseProxy.ServeHTTP(w, req)
		}
	})
	log.Info("Starting Server", map[string]interface{}{"port": config.Network.Downstream.Port})
	log.Fatal("Stopping service", http.ListenAndServe(":"+strconv.Itoa(config.Network.Downstream.Port), router), nil)
}
