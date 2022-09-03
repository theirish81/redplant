package main

import (
	"encoding/base64"
	"os"
	"reflect"
	"strings"
)

type StringMap map[string]string
type AnyMap map[string]any
type RulesMap map[string]RoutesMap
type RoutesMap map[string]*Rule

type PathGroupedRoutes map[string][]*Rule

func (m *RoutesMap) GroupByPattern() PathGroupedRoutes {
	groups := PathGroupedRoutes{}
	for _, rule := range *m {
		pattern := rule._pattern
		if _, ok := groups[pattern]; !ok {
			groups[pattern] = make([]*Rule, 0)
		}
		groups[pattern] = append(groups[pattern], rule)
	}
	return groups
}

// stringInArray will search a string in an array of strings and return true if the string is found
func stringInArray(search string, array []string) bool {
	for _, sx := range array {
		if search == sx {
			return true
		}
	}
	return false
}

// getFieldName given a reflect.Value, it will return the name of the field for a given index
func getFieldName(val reflect.Value, index int) string {
	structField := reflect.Indirect(val).Type().Field(index)
	return structField.Name
}

// getEnvs converts environment variables to a map
func getEnvs() StringMap {
	envs := make(StringMap)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}
	return envs
}

// hasPrefixes will check whether an input string has one of the provided prefixes
func hasPrefixes(data string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(data, prefix) {
			return true
		}
	}
	return false
}

// isString given an interface, it will return true if the data type is a string
func isString(data any) bool {
	if data == nil {
		return false
	}
	return reflect.ValueOf(data).Type().String() == "string"
}

// parseBasicAuth will parse a basic auth header value
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

// convertMaps will recursively go through a nested structure and converting map[any]any to
// map[string]any
func convertMaps(intf any) any {
	switch obj := intf.(type) {
	case map[string]any:
		for k, v := range obj {
			obj[k] = convertMaps(v)
		}
	case map[any]any:
		nuMap := map[string]any{}
		for k, v := range obj {
			nuMap[k.(string)] = convertMaps(v)
		}
		return nuMap
	case []any:
		for index, object := range obj {
			obj[index] = convertMaps(object)
		}
	}
	return intf
}

// IsHTTP will tell you if the given string is somewhat likely to be an HTTP(s) URL
func IsHTTP(file string) bool {
	return hasPrefixes(file, []string{"http://", "https://"})
}

func IsGZIP(transferEncoding []string) bool {
	for _, tx := range transferEncoding {
		if strings.Contains(tx, "gzip") {
			return true
		}
	}
	return false
}
