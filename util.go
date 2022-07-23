package main

import (
	"encoding/base64"
	"errors"
	"net/url"
	"os"
	"reflect"
	"strings"
)

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
func getEnvs() *map[string]string {
	envs := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = pair[1]
	}
	return &envs
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
func isString(data interface{}) bool {
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

// convertMaps will recursively go through a nested structure and converting map[interface{}]interface{} to
// map[string]interface{}
func convertMaps(intf interface{}) interface{} {
	switch obj := intf.(type) {
	case map[string]interface{}:
		for k, v := range obj {
			obj[k] = convertMaps(v)
		}
	case map[interface{}]interface{}:
		nuMap := map[string]interface{}{}
		for k, v := range obj {
			nuMap[k.(string)] = convertMaps(v)
		}
		return nuMap
	case []interface{}:
		for index, object := range obj {
			obj[index] = convertMaps(object)
		}
	}
	return intf
}

func FileNameFormat(file string) (string, error) {
	if strings.HasPrefix(file, "file://") {
		localUrl, err := url.Parse(file)
		if err != nil {
			return "", errors.New("wrong file format")
		}
		return localUrl.Host + localUrl.Path, nil
	}
	return file, nil
}

func IsHTTP(file string) bool {
	return hasPrefixes(file, []string{"http://", "https://"})
}
