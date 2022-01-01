package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

// SetupRouter will set up the router and return it
func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	// Creating a custom reverse proxy
	reverseProxy := &httputil.ReverseProxy{
		// Custom Director
		Director: func(req *http.Request) {
			wrapper := GetWrapper(req)
			// if wrapper is nil, then we don't have any match, and we shouldn't be here
			if wrapper == nil {
				return
			}
			wrapper.Request = req
			wrapper.ExpandRequestIfNeeded()
			wrapper.Rule.Request._sidecars.Run(wrapper.Clone())
			handleURL(wrapper.Rule, req)
			wrapper.Metrics.ReqTransStart = time.Now()
			_, err := wrapper.Rule.Request._transformers.Transform(wrapper)
			wrapper.Metrics.ReqTransEnd = time.Now()
			wrapper.Err = err
		},
		// Custom  error handler
		ErrorHandler: func(writer http.ResponseWriter, request *http.Request, err error) {
			wrapper := GetWrapper(request)
			if wrapper != nil {
				if wrapper.Rule.Request._transformers.HandleError(err, &writer) {
					return
				}
				if wrapper.Rule.Response._transformers.HandleError(err, &writer) {
					return
				}
			}
			switch err.Error() {
			case "no_mapping":
				writer.WriteHeader(404)
			default:
				log.Error("Error while serving resource", err, map[string]interface{}{"url": request.URL.String()})
				writer.WriteHeader(500)
			}
		},
		// Custom transport
		Transport: configTransport(),
		// Post trip response modification
		ModifyResponse: func(response *http.Response) error {
			wrapper := GetWrapper(response.Request)
			if wrapper != nil {
				wrapper := GetWrapper(response.Request)
				wrapper.Response = response
				wrapper.ExpandResponseIfNeeded()
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
	// Routing based on hostname
	for k, rules := range config.Rules {
		func(rules map[string]*Rule) {
			// Handler for one hostname
			router.Host(k).HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// For each rule for a given hostname...
				for _, rule := range rules {
					// ... if there's match, then we can enrich with a context
					if success := rule._pattern.MatchString(req.URL.Path); success {
						req = ReqWithContext(req, rule)
						break
					}
				}
				reverseProxy.ServeHTTP(w, req)
			}).Name(k)
		}(rules)
	}
	return router
}

// handleURL transforms the URL based on the rules
func handleURL(rule *Rule, req *http.Request) {
	newUrl, _ := url.Parse(rule.Origin)
	reqPath := req.URL.Path
	if len(rule.StripPrefix) > 0 {
		reqPath = strings.Replace(reqPath, rule.StripPrefix, "", 1)
	}
	newUrl.Path = newUrl.Path + reqPath
	req.URL = newUrl
	req.Host = req.URL.Host
}
