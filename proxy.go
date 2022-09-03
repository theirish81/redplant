package main

import (
	"errors"
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
			wrapper.Request = NewAPIRequest(req)

			if !hasMethod(wrapper) {
				wrapper.Err = errors.New("method_not_allowed")
				return
			}
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
				// If the connection has been hijacked, we can't operate on the response anymore.
				// Therefore, we simply return and ignore any further activity.
				if wrapper.Hijacked && err.Error() == "connection_hijacked" {
					return
				}
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
			case "method_not_allowed":
				writer.WriteHeader(405)
			default:
				if prom != nil {
					prom.InternalErrorsCounter.Inc()
				}
				log.Error("Error while serving resource", err, AnyMap{"url": request.URL.String()})
				writer.WriteHeader(500)
			}
		},
		// Custom transport
		Transport: configTransport(),
		// Post trip response modification
		ModifyResponse: func(response *http.Response) error {
			wrapper := GetWrapper(response.Request)
			if wrapper != nil {
				// if the connection has been hijacked by a websocket, we can't operate on the response anymore.
				// The error handler is the only stage that can deactivate further operations, so we're
				// returning an error to trigger that handler.
				if wrapper.Hijacked {
					return errors.New("connection_hijacked")
				}
				wrapper := GetWrapper(response.Request)
				wrapper.Response = NewAPIResponse(response)
				for k, v := range wrapper.ApplyHeaders {
					wrapper.Response.Header.Set(k, v[0])
				}
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
		func(rules RoutesMap) {
			hostRoute := router.Host(k).Subrouter()
			groups := rules.GroupByPattern()
			for _, routes := range groups {
				func(routes []*Rule) {
					hostRoute.HandleFunc(routes[0]._pattern, func(writer http.ResponseWriter, request *http.Request) {
						for _, rule := range routes {
							methodMatch := rule._patternMethod == "" || rule._patternMethod == strings.ToLower(request.Method)
							if methodMatch {
								request = ReqWithContext(request, writer, rule)
								break
							}
						}
						reverseProxy.ServeHTTP(writer, request)
					})
				}(routes)
			}
		}(rules)
	}
	return router
}

// hasMethod will check if the method set in the request is among the ones listed in the Rule.AllowedMethods setting.
// IF Rule.AllowedMethods is nil or empty, then all methods are allowed
func hasMethod(wrapper *APIWrapper) bool {
	if wrapper.Rule.AllowedMethods != nil && len(wrapper.Rule.AllowedMethods) > 0 {
		method := wrapper.Request.Method
		found := false
		for _, m := range wrapper.Rule.AllowedMethods {
			if strings.ToLower(method) == strings.ToLower(m) {
				found = true
			}
			break
		}
		return found
	}
	return true
}

// handleURL transforms the URL based on the rules
func handleURL(rule *Rule, req *http.Request) {
	newUrl, _ := url.Parse(rule.Origin)
	reqPath := req.URL.Path
	if len(rule.StripPrefix) > 0 {
		reqPath = strings.Replace(reqPath, rule.StripPrefix, "", 1)
	}
	newUrl.Path = newUrl.Path + reqPath
	newUrl.RawQuery = req.URL.RawQuery
	req.URL = newUrl
	req.Host = req.URL.Host
}
