package main

import (
	"encoding/json"
	"github.com/getkin/kin-openapi/openapi3"
	"net/url"
	"regexp"
)

type OARule struct {
	Request  RequestConfig
	Response ResponseConfig
}

var repl = regexp.MustCompile(`{.*?}`)

func OpenAPI2Rules(openAPIConfigs map[string]*OpenAPIConfig) map[string]map[string]*Rule {
	res := make(map[string]map[string]*Rule)
	for host, cfg := range openAPIConfigs {
		oa, _ := loadOpenAPI(*cfg)
		serverURL, _ := url.Parse(oa.Servers[cfg.ServerIndex].URL)
		partialPath := serverURL.Path
		paths := make(map[string]*Rule)
		for path, operations := range oa.Paths {
			px := string(repl.ReplaceAll([]byte(partialPath+path), []byte(".*")))
			for _, m := range listMethods(operations) {
				op := getOperationByMethod(operations, m)
				redExtension := op.Extensions["x-redplant"]
				oaRule := OARule{}
				if redExtension != nil {
					err := json.Unmarshal(redExtension.(json.RawMessage), &oaRule)
					if err != nil {
						log.Error("could not read RedPlant configuration from OpenAPI", err, nil)
						continue
					}
				}
				rule := Rule{Origin: serverURL.String(), StripPrefix: partialPath, Pattern: px,
					Request: oaRule.Request, Response: oaRule.Response}
				paths["["+m+"] "+px] = &rule
			}
		}
		res[host] = paths
	}
	return res
}

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

func loadOpenAPI(cfg OpenAPIConfig) (*openapi3.T, error) {
	return openapi3.NewLoader().LoadFromFile(cfg.File)
}

func MergeRules(existing map[string]map[string]*Rule, new map[string]map[string]*Rule) map[string]map[string]*Rule {
	if existing == nil {
		existing = make(map[string]map[string]*Rule)
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
