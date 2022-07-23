package main

import (
	"github.com/neotoolkit/openapi"
	"net/url"
	"os"
	"regexp"
)

var repl = regexp.MustCompile(`{.*?}`)

func OpenAPI2Rules(openAPIConfigs map[string]*OpenAPIConfig) map[string]map[string]*Rule {
	res := make(map[string]map[string]*Rule)
	for host, cfg := range openAPIConfigs {
		op, _ := loadOpenAPI(*cfg)
		serverURL, _ := url.Parse(op.Servers[cfg.ServerIndex].URL)
		partialPath := serverURL.Path
		paths := make(map[string]*Rule)
		for path, actions := range op.Paths {
			px := string(repl.ReplaceAll([]byte(partialPath+path), []byte(".*")))
			rule := Rule{Origin: serverURL.String(), StripPrefix: partialPath, Pattern: px, AllowedMethods: listMethods(actions)}
			paths[px] = &rule
		}
		res[host] = paths
	}
	return res
}

func listMethods(path *openapi.Path) []string {
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

func loadOpenAPI(cfg OpenAPIConfig) (openapi.OpenAPI, error) {
	data, err := os.ReadFile(cfg.File)
	if err != nil {
		return openapi.OpenAPI{}, err
	}
	return openapi.Parse(data)
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
