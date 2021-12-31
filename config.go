package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/theirish81/yamlRef"
	"github.com/xo/dburl"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

// Config is the root object of the configuration
type Config struct {
	// Variables are global variables for the API scope
	Variables map[string]string `yaml:"variables"`
	// Network is the network configuration
	Network Network `yaml:"network"`
	// Before is a set of transformers + sidecars to be executed before the rule's set of transformers + sidecars
	Before BeforeAfterConfig `yaml:"before"`
	// After is a set of transformers + sidecars to be executed after the rule's set of transformers + sidecars
	After BeforeAfterConfig `yaml:"after"`
	// Rules are the routes
	Rules map[string]map[string]*Rule `yaml:"rules"`
}

// Rule is an upstream route
type Rule struct {
	// Origin is the origin URL
	Origin string `yaml:"origin"`
	// Request is the request transformation pipeline
	Request RequestConfig `yaml:"request"`
	// Response is the response transformation pipeline
	Response ResponseConfig `yaml:"response"`
	// Pattern is the pattern that matches the path of the URL, in the form of a regexp string
	Pattern string `yaml:"pattern"`
	// _pattern is the compiled Pattern
	_pattern *regexp.Regexp
	// db is the database connection, assuming this rule is using a DBTripper
	db *sqlx.DB
}

// RequestConfig is the configuration of the request pipeline
type RequestConfig struct {
	// Transformers is an array of transformer configurations
	Transformers []TransformerConfig `yaml:"transformers"`
	// Sidecars is an array of sidecar configurations
	Sidecars []SidecarConfig `yaml:"sidecars"`
	// _transformers is an array of the actual transformer instances
	_transformers *RequestTransformers
	// _sidecars is an array of the actual transformer instances
	_sidecars *RequestSidecars
}

// ResponseConfig is the configuration of the response pipeline
type ResponseConfig struct {
	// Transformers is an array of transformer configurations
	Transformers []TransformerConfig `yaml:"transformers"`
	// Sidecars is an array of sidecar configurations
	Sidecars []SidecarConfig `yaml:"sidecars"`
	// _transformers is an array of the actual transformer instances
	_transformers *ResponseTransformers
	// _sidecars is an array of the actual transformer instances
	_sidecars *ResponseSidecars
}

// TransformerConfig is the base transformer configuration
type TransformerConfig struct {
	// Id is the name of the transformer
	Id string `yaml:"id"`
	// ActivateOnTags is a list of tags for which this transformer will activate
	ActivateOnTags []string `yaml:"activateOnTags"`
	// Params is a map of configuration params for the transformer
	Params map[string]interface{} `yaml:"params"`
}

// SidecarConfig is the configuration of a sidecar
type SidecarConfig struct {
	// Id is the name of the transformer
	Id string `yaml:"id"`
	// ActivateOnTags is a list of tags for which this sidecar will activate
	ActivateOnTags []string `yaml:"activateOnTags"`
	// Workers is the total number of workers we should have for this sidecar
	Workers int `yaml:"workers"`
	// Block if set to true, will block the main flow if all sidecars are busy
	Block bool `yaml:"block"`
	// Params is a map of configuration params for the sidecar
	Params map[string]interface{} `yaml:"params"`
}

// BeforeAfterConfig represents a set of transformers + sidecars to be executed before or after the rule's
// transformers + sidecar
type BeforeAfterConfig struct {
	// Request is the configuration for the request pipeline
	Request RequestConfig `yaml:"request"`
	// Response is the configuration for the request pipeline
	Response ResponseConfig `yaml:"response"`
}

// Network is the network configuration
type Network struct {
	// Upstream is the configuration of the client
	Upstream Upstream `yaml:"upstream"`
	// Downstream is the configuration of the web server
	Downstream Downstream `yaml:"downstream"`
}

// Downstream is the downstream configuration
type Downstream struct {
	// Port is the port number we should listen on
	Port int `yaml:"port"`
	// Tls is the secure connection configuration
	Tls []Tls `yaml:"tls"`
}

// Upstream is the upstream configuration
type Upstream struct {
	// Timeout is the global timeout as a duration string
	Timeout string `yaml:"timeout"`
	// KeepAlive is the keep-alive timeout as a duration string
	KeepAlive string `yaml:"keepAlive"`
	// MaxIdleConnection is the maximum number of allowed idle connections
	MaxIdleConnections int `yaml:"maxIdleConnections"`
	// IdleConnectionTimeout is the timeout for an idle connection to be evicted
	IdleConnectionTimeout string `yaml:"idleConnectionTimeout"`
	// ExpectContinueTimeout is the timeout for the "continue" HTTP operation
	ExpectContinueTimeout string `yaml:"expectContinueTimeout"`
}

// Tls is the configuration for the secure connection
type Tls struct {
	Host string `yaml:"host"`
	// Cert is the path to a certificate
	Cert string `yaml:"cert"`
	// Key is the path to a key
	Key string `yaml:"key"`
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
		parsed, _ := Templ(v, envs)
		config.Variables[k] = parsed
	}
	return config
}

// Init initialize the configuration
func (c *Config) Init() {
	// For every domain definition
	for _, topRule := range c.Rules {
		// For every rule within the domain definition
		for pattern, rule := range topRule {
			var err error
			// Compile the pattern regexp and store it
			rule._pattern, err = regexp.Compile(pattern)
			if err != nil {
				log.Fatal("Pattern is not a valida regex", err, nil)
			}
			// The origin may be a template, so we evaluate it
			rule.Origin, err = Templ(rule.Origin, nil)
			if err != nil {
				log.Fatal("Could not parse origin", err, map[string]interface{}{"origin": rule.Origin})
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
			rule.Request._sidecars = NewRequestSidecars(&mergedReqSidecars)

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
		}
	}
}

type LoggerConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Path   string `yaml:"path"`
}

func LoadLoggerConfig(path *string) (LoggerConfig, error) {
	cfg := LoggerConfig{}
	if path == nil || *path == "" {
		cfg.Level = "INFO"
		cfg.Format = "simple"
		cfg.Path = ""
		return cfg, nil
	}
	fileContent, err := ioutil.ReadFile(*path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(fileContent, &cfg)
	return cfg, err
}
