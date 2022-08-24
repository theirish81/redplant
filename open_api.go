package main

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers/gorillamux"
	"net/url"
	"regexp"
)

// OARule is a data structure used to map Request and Response config items within the OpenAPI document, defined in
// the `x-redplant` extension
// Request is the request config
// Response is the response config
type OARule struct {
	Request  RequestConfig
	Response ResponseConfig
}

// repl is a regular expression matching an URI parameter as defined in OpenAPI
var repl = regexp.MustCompile(`{.+?}`)
var paramWildcard = ".*"

// OpenAPI2Rules will load a number of OpenAPI files and convert it to a set of RedPlant Rules
// `openAPIConfigs` is a set of OpenAPIConfig, the RedPlant configuration for them
func OpenAPI2Rules(openAPIConfigs map[string]*OpenAPIConfig) RulesMap {
	res := make(RulesMap)
	// for each RedPlant host, one openAPI can be mapped. So for each host, we obtain an OpenAPI config
	for host, cfg := range openAPIConfigs {
		// first we load and parse the OpenAPI spec
		oa, err := loadOpenAPI(*cfg)
		if err != nil {
			log.Error("could not load the OpenAPI spec", err, AnyMap{"file": cfg.File})
			continue
		}
		// Servers can contain multiple items. We will pick one based on the configuration. We pare it into a URL
		// for ease of use. Partial URLs are not acceptable because they defeat the purpose
		serverURL, err := url.Parse(oa.Servers[cfg.ServerIndex].URL)
		if err != nil {
			log.Error("could not parse server URL in OpenAPI spec", err, AnyMap{"url": oa.Servers[cfg.ServerIndex].URL})
			continue
		}
		// As the `Servers` definition may include not only a protocol and a host, but also a partial path, we extract it
		partialPath := serverURL.Path
		// An empty set of RedPlant rules we will fill
		rules := make(map[string]*Rule)
		// For every OpenAPI path, we obtain a path as a string, and its relative configuration
		router, _ := gorillamux.NewRouter(oa)
		for path, operations := range oa.Paths {
			// converting the URI variables into our Regexp format
			px := "^" + string(repl.ReplaceAll([]byte(partialPath+path), []byte(paramWildcard))) + "$"
			// for each method defined in `operations`
			for _, m := range listMethods(operations) {
				op := getOperationByMethod(operations, m)
				// we look for the RedPlant extension
				redExtension := op.Extensions["x-redplant"]
				oaRule := OARule{}
				// if the RedPlant extension has indeed been found...
				if redExtension != nil {
					// ... we unmarshal it into our OARule
					err := json.Unmarshal(redExtension.(json.RawMessage), &oaRule)
					if err != nil {
						log.Error("could not read RedPlant configuration from OpenAPI", err, nil)
						continue
					}
				}
				// Creating the rule, containing the OpenAPI base information, plus the information from the extension
				// if it has been provided
				rule := Rule{Origin: serverURL.String(), StripPrefix: partialPath, Pattern: px,
					Request: oaRule.Request, Response: oaRule.Response, oaOperation: op, oa: oa, oaRouter: &router}
				// Mapping the rule to the path, using the explicit method notation, as this is how
				// OpenAPI works
				rules["["+m+"] "+px] = &rule
			}
		}
		// Finally, assigning the Rules to the host
		res[host] = rules
	}
	return res
}

// getOperationByMethod will take a PathItem and a method, and return the corresponding Operation based on the method
// `operations` the PathItem
// `method` the method
func getOperationByMethod(operations *openapi3.PathItem, method string) *openapi3.Operation {
	switch method {
	case "get":
		return operations.Get
	case "post":
		return operations.Post
	case "patch":
		return operations.Patch
	case "put":
		return operations.Put
	case "delete":
		return operations.Delete
	default:
		return operations.Get
	}
}

// loadOpenAPI loads and parses OpenAPI
// `cfg` is an OpenAPI RedPlant configuration
func loadOpenAPI(cfg OpenAPIConfig) (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromFile(cfg.File)
}

// MergeRules will merge a set of rules with another set or rules. This is necessary because OpenAPI specs
// and regular RedPlant specs can coexist
// `existing` is the existing set of rules
// `new` is the new set of rules
func MergeRules(existing RulesMap, new RulesMap) RulesMap {
	if existing == nil {
		existing = make(RulesMap)
	}
	for domain, routes := range new {
		route, ok := existing[domain]
		if ok {
			for path, val := range routes {
				route[path] = val
			}
		}
		if !ok {
			existing[domain] = routes
		}
	}
	return existing
}

// listMethods will return an array of methods given a PathItem
// `path` is a PathItem
func listMethods(path *openapi3.PathItem) []string {
	methods := make([]string, 0)
	if path.Patch != nil {
		methods = append(methods, "patch")
	}
	if path.Delete != nil {
		methods = append(methods, "delete")
	}
	if path.Put != nil {
		methods = append(methods, "put")
	}
	if path.Post != nil {
		methods = append(methods, "post")
	}
	if path.Get != nil {
		methods = append(methods, "get")
	}
	return methods
}
