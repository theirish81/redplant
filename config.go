package main

import (
	"context"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/routers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/theirish81/yamlRef"
	"github.com/xo/dburl"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Config is the root object of the configuration
// Variables are global variables for the API scope
// Network is the network configuration
// Before is a set of transformers + sidecars to be executed before the rule's set of transformers + sidecars
// After is a set of transformers + sidecars to be executed after the rule's set of transformers + sidecars
// Rules are the routes
// OpenAPI is the OpenAPI way tof configuring rules
// Prometheus is the Prometheus configuration object
type Config struct {
	Variables  StringMap                 `yaml:"variables"`
	Network    Network                   `yaml:"network"`
	Before     BeforeAfterConfig         `yaml:"before"`
	After      BeforeAfterConfig         `yaml:"after"`
	Rules      DomainsMap                `yaml:"rules"`
	OpenAPI    map[string]*OpenAPIConfig `yaml:"openAPI"`
	Prometheus *PrometheusConfig         `yaml:"prometheus"`
}

// DomainsMap is a map of domain=path objects
type DomainsMap map[string]RoutesMap

// RoutesMap is a map of path=rule object
type RoutesMap map[string]*Rule

// RoutesArray is the rules map converted into an array
type RoutesArray []*Rule

// NewRoutesArray will create a RoutesArray out of a RoutesMap
func NewRoutesArray(m RoutesMap) RoutesArray {
	arr := RoutesArray{}
	for pattern, route := range m {
		route.Pattern = pattern
		arr = append(arr, route)
	}
	return arr
}

func (a RoutesArray) Len() int {
	return len(a)
}

func (a RoutesArray) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a RoutesArray) Less(i, j int) bool {
	firstDynamic := a.isDynamic(a[i].Pattern)
	secondDynamic := a.isDynamic(a[j].Pattern)
	if firstDynamic && !secondDynamic {
		return false
	}
	if !firstDynamic && secondDynamic {
		return true
	}
	if len(a[i].Pattern) != len(a[j].Pattern) {
		return len(a[i].Pattern) > len(a[j].Pattern)
	}
	return true
}

// isDynamic will return true if the provided path contains variable parts
func (a RoutesArray) isDynamic(path string) bool {
	return strings.Contains(path, "{") && strings.Contains(path, "}")
}

// ToOrderedRoutes will convert RoutesMap into an array, ordered by sort of a priority
func (m *RoutesMap) ToOrderedRoutes() RoutesArray {
	arr := NewRoutesArray(*m)
	sort.Sort(&arr)
	return arr
}

// OpenAPIConfig is an OpenAPI configuration object
type OpenAPIConfig struct {
	File        string `yaml:"file"`
	ServerIndex int    `yaml:"server_index"`
}

// Rule is an upstream route
// Origin is the origin URL
// StripPrefix the inbound URL path prefix to strip out
// Request is the request transformation pipeline
// Response is the response transformation pipeline
// Pattern is the pattern that matches the path of the URL, in the form of a regexp string
// _pattern is the path component of the pattern. This is derived from the path pattern, key of the rule
// _patternMethod is the method component of the pattern, assuming it's there. This is derived from the path pattern
// key of the rule
// db is the database connection, assuming this rule is using a DBTripper
type Rule struct {
	Origin         string         `yaml:"origin"`
	StripPrefix    string         `yaml:"stripPrefix"`
	Request        RequestConfig  `yaml:"request"`
	Response       ResponseConfig `yaml:"response"`
	Pattern        string         `yaml:"pattern"`
	AllowedMethods []string       `yaml:"allowedMethods"`
	_pattern       string
	_patternMethod string
	oa             *openapi3.T
	oaOperation    *openapi3.Operation
	oaRouter       *routers.Router
	db             *sqlx.DB
}

// RequestConfig is the configuration of the request pipeline
// Transformers is an array of transformer configurations
// Sidecars is an array of sidecar configurations
// _transformers is an array of the actual transformer instances
// _sidecars is an array of the actual transformer instances
type RequestConfig struct {
	Transformers  []TransformerConfig `yaml:"transformers"`
	Sidecars      []SidecarConfig     `yaml:"sidecars"`
	_transformers *RequestTransformers
	_sidecars     *RequestSidecars
}

// ResponseConfig is the configuration of the response pipeline
// Transformers is an array of transformer configurations
// Sidecars is an array of sidecar configurations
// _transformers is an array of the actual transformer instances
// _sidecars is an array of the actual transformer instances
type ResponseConfig struct {
	Transformers  []TransformerConfig `yaml:"transformers"`
	Sidecars      []SidecarConfig     `yaml:"sidecars"`
	_transformers *ResponseTransformers
	_sidecars     *ResponseSidecars
}

// STLogConfig is a specialised logging configuration for transformers and sidecars
// Level is the level of the logger
// Path is a path to a file, in case you want to log to a file. Leave empty for stdout
// Format the format of the log. Can either be JSON or simple
// Prometheus holds the Prometheus configuration
type STLogConfig struct {
	Level      string             `yaml:"level"`
	Path       string             `yaml:"path"`
	Format     string             `yaml:"format"`
	Prometheus STPrometheusConfig `yaml:"prometheus"`
}

// STPrometheusConfig is the configuration of Prometheus, for a specific sidecar or transformer
// Enabled will enable the integration, if set to true
// Prefix will be used as a prefix for the collected metric
type STPrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Prefix  string `yaml:"prefix"`
}

// TransformerConfig is the base transformer configuration
// Id is the name of the transformer
// ActivateOnTags is a list of tags for which this transformer will activate
// Params is a map of configuration params for the transformer
type TransformerConfig struct {
	Id             string       `yaml:"id"`
	ActivateOnTags []string     `yaml:"activateOnTags"`
	Logging        *STLogConfig `yaml:"logging"`
	Params         AnyMap       `yaml:"params"`
}

// SidecarConfig is the configuration of a sidecar
// Id is the name of the transformer
// ActivateOnTags is a list of tags for which this sidecar will activate
// Workers is the total number of workers we should have for this sidecar
// Block if set to true, will block the main flow if all sidecars are busy
// DropOnOverflow if set to true, will drop messages if the queue is blocked
// Params is a map of configuration params for the sidecar
type SidecarConfig struct {
	Id             string       `yaml:"id"`
	ActivateOnTags []string     `yaml:"activateOnTags"`
	Workers        int          `yaml:"workers"`
	Queue          int          `yaml:"queue"`
	Block          bool         `yaml:"block"`
	DropOnOverflow bool         `yaml:"blockOnOverflow"`
	Logging        *STLogConfig `yaml:"logging"`
	Params         AnyMap       `yaml:"params"`
}

// BeforeAfterConfig represents a set of transformers + sidecars to be executed before or after the rule's
// transformers + sidecar
// Request is the configuration for the request pipeline
// Response is the configuration for the request pipeline
type BeforeAfterConfig struct {
	Request  RequestConfig  `yaml:"request"`
	Response ResponseConfig `yaml:"response"`
}

// Network is the network configuration
// Upstream is the configuration of the client
// Downstream is the configuration of the web server
type Network struct {
	Upstream   Upstream   `yaml:"upstream"`
	Downstream Downstream `yaml:"downstream"`
}

// Downstream is the downstream configuration
// Port is the port number we should listen on
// Tls is the secure connection configuration
type Downstream struct {
	Port int   `yaml:"port"`
	Tls  []Tls `yaml:"tls"`
}

// Upstream is the upstream configuration
// Timeout is the global timeout as a duration string
// KeepAlive is the keep-alive timeout as a duration string
// MaxIdleConnections is the maximum number of allowed idle connections
// IdleConnectionTimeout is the timeout for an idle connection to be evicted
// ExpectContinueTimeout is the timeout for the "continue" HTTP operation
type Upstream struct {
	Timeout               string `yaml:"timeout"`
	KeepAlive             string `yaml:"keepAlive"`
	MaxIdleConnections    int    `yaml:"maxIdleConnections"`
	IdleConnectionTimeout string `yaml:"idleConnectionTimeout"`
	ExpectContinueTimeout string `yaml:"expectContinueTimeout"`
}

// Tls is the configuration for the secure connection
// Host is the hostname this certificate is for
// Cert is the path to a certificate
// Key is the path to a key
type Tls struct {
	Host string `yaml:"host"`
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// PrometheusConfig is the configuration of the Prometheus metrics endpoint
// Port is the port we're exposing the endpoint to
// Path is the URL path the metrics will be exposed to
type PrometheusConfig struct {
	Port int
	Path string
}

// LoadConfig loads the configuration
func LoadConfig(file string) Config {
	config := Config{}
	// Unfortunately arrays get initialized as nil, and that's not comfortable, so we manually initialize them as empty
	config.Before.Request.Transformers = make([]TransformerConfig, 0)
	config.Before.Request.Sidecars = make([]SidecarConfig, 0)
	config.Before.Response.Transformers = make([]TransformerConfig, 0)
	config.Before.Response.Sidecars = make([]SidecarConfig, 0)
	config.After.Request.Transformers = make([]TransformerConfig, 0)
	config.After.Request.Sidecars = make([]SidecarConfig, 0)
	config.After.Response.Transformers = make([]TransformerConfig, 0)
	config.After.Response.Sidecars = make([]SidecarConfig, 0)

	// Loading the configuration file, merging referenced files
	data, err := yamlRef.MergeAndMarshall(file)
	if err != nil {
		log.Fatal("Could not load the configuration file", err, nil)
	}
	// Unmarshalling the data
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Could not parse the configuration file", err, nil)
	}

	// The Variables section in the configuration can contain templates. A template will be evaluated against
	// environment variables.
	// Loading the environment variables as scope
	envs := getEnvs()
	// For each configuration variable, evaluate templates against environment variables
	for k, v := range config.Variables {
		parsed, _ := template.Templ(context.Background(), v, envs)
		config.Variables[k] = parsed
	}
	return config
}

// Init initialize the configuration
func (c *Config) Init() {
	if c.OpenAPI != nil {
		c.Rules = MergeRules(c.Rules, OpenAPI2Rules(c.OpenAPI))
	}
	// For every domain definition
	for domain, topRule := range c.Rules {
		// For every rule within the domain definition
		for _, rule := range topRule.ToOrderedRoutes() {
			var err error
			// separate the method and the actual pattern
			rule._patternMethod, rule._pattern = extractPattern(rule.Pattern)
			// The origin may be a template, so we evaluate it
			rule.Origin, err = template.Templ(context.Background(), rule.Origin, nil)
			if err != nil {
				log.Fatal("Could not parse origin", err, AnyMap{"origin": rule.Origin})
			}
			// Before, Rule and After request transformers configuration are merged into one array...
			mergedReqTransformers := append(append(c.Before.Request.Transformers, rule.Request.Transformers...), c.After.Request.Transformers...)
			// ... and then transformers get initialized
			rule.Request._transformers, err = NewRequestTransformers(&mergedReqTransformers)
			if err != nil {
				log.Fatal("Error initializing request transformers ", err, nil)
			}

			// Before, Rule and After response transformers configuration are merged into one array...
			mergedResTransformers := append(append(c.Before.Response.Transformers, rule.Response.Transformers...), c.After.Response.Transformers...)
			// ... and then transformers get initialized
			rule.Response._transformers, err = NewResponseTransformers(&mergedResTransformers)
			if err != nil {
				log.Fatal("Error initializing response transformers ", err, nil)
			}

			// Before, Rule and After request sidecars configuration are merged into one array...
			mergedReqSidecars := append(append(c.Before.Request.Sidecars, rule.Request.Sidecars...), c.After.Request.Sidecars...)
			// ... and then sidecars get initialized
			rule.Request._sidecars = NewRequestSidecars(mergedReqSidecars)

			// Before, Rule and After response sidecars configuration are merged into one array...
			mergedResSidecars := append(append(c.Before.Response.Sidecars, rule.Response.Sidecars...), c.After.Response.Sidecars...)
			// ... and then sidecars get initialized
			rule.Response._sidecars = NewResponseSidecars(&mergedResSidecars)

			// If the origin is a URI to a DB...
			if hasPrefixes(rule.Origin, []string{"postgres://", "mysql://"}) {
				// Parse the URI
				databaseUrl, err := dburl.Parse(rule.Origin)
				if err != nil {
					log.Fatal("Could not parse the database URI", err, nil)
				}
				// Open the connection and store the reference
				rule.db, err = sqlx.Open(databaseUrl.Driver, databaseUrl.DSN)
				if err != nil {
					log.Fatal("Could not connect to the database", err, nil)
				}
			}
			log.Info("route registered", AnyMap{"pattern": rule.Pattern, "domain": domain})
		}
	}
}

// LoggerConfig is the logger configuration
type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Path   string `yaml:"path"`
}

// LoadLoggerConfig loads the logger configuration from the provided file path
func LoadLoggerConfig(path *string) (LoggerConfig, error) {
	cfg := LoggerConfig{}
	if path == nil || *path == "" {
		cfg.Level = "INFO"
		cfg.Format = "simple"
		cfg.Path = ""
		return cfg, nil
	}
	fileContent, err := os.ReadFile(*path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(fileContent, &cfg)
	return cfg, err
}

// methodFinderRegexp will find the explicit method at the beginning of the path, if defined
var methodFinderRegexp, _ = regexp.Compile("^\\[(get|post|put|patch|delete|options)\\]")

// extractPattern will separate the explicit method from the path, if that's how the path definition was composed, and
// return the two components. If the path does not contain an explicit method, then it will return an empty string
// for the method, and the path as second return value
func extractPattern(path string) (string, string) {
	if hasPrefixes(path, []string{"[get]", "[post]", "[put]", "[patch]", "[delete]", "[options]"}) {
		method := strings.Replace(strings.Replace(string(methodFinderRegexp.Find([]byte(path))), "[", "", 1), "]", "", 1)
		cleanPath := strings.TrimSpace(strings.SplitN(path, "]", 2)[1])
		return method, cleanPath
	}
	return "", path
}
