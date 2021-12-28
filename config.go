package main

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/theirish81/yamlRef"
	"github.com/xo/dburl"
	"gopkg.in/yaml.v2"
	"os"
	"regexp"
	"strings"
)

// Config is the root object of the configuration
type Config struct {
	Variables map[string]string           `yaml:"variables"`
	Network   Network                     `yaml:"network"`
	Request   RequestConfig               `yaml:"request"`
	Response  ResponseConfig              `yaml:"response"`
	Rules     map[string]map[string]*Rule `yaml:"rules"`
}

// Rule is an upstream route
type Rule struct {
	Origin   string            `yaml:"origin"`
	Headers  map[string]string `yaml:"headers"`
	Request  RequestConfig     `yaml:"request"`
	Response ResponseConfig    `yaml:"response"`
	Pattern  string            `yaml:"pattern"`
	_pattern *regexp.Regexp
	db       *sqlx.DB
}

// RequestConfig is the configuration of the request pipeline
type RequestConfig struct {
	Transformers  []TransformerConfig `yaml:"transformers"`
	Sidecars      []SidecarConfig     `yaml:"sidecars"`
	_transformers *RequestTransformers
	_sidecars     *RequestSidecars
}

// ResponseConfig is the configuration of the response pipeline
type ResponseConfig struct {
	Transformers  []TransformerConfig `yaml:"transformers"`
	Sidecars      []SidecarConfig     `yaml:"sidecars"`
	_transformers *ResponseTransformers
	_sidecars     *ResponseSidecars
}

// TransformerConfig is the base transformer configuration
type TransformerConfig struct {
	Id     string                 `yaml:"id"`
	Params map[string]interface{} `yaml:"params"`
}

// SidecarConfig is the configuration of a sidecar
type SidecarConfig struct {
	Id      string                 `yaml:"id"`
	Workers int                    `yaml:"workers"`
	Block   bool                   `yaml:"block"`
	Params  map[string]interface{} `yaml:"params"`
}

// Network is the network configuration
type Network struct {
	Upstream   Upstream   `yaml:"upstream"`
	Downstream Downstream `yaml:"downstream"`
}

// Downstream is the downstream configuration
type Downstream struct {
	Port int `yaml:"port"`
}

// Upstream is the upstream configuration
type Upstream struct {
	Timeout               string `yaml:"timeout"`
	KeepAlive             string `yaml:"keepAlive"`
	MaxIdleConnections    int    `yaml:"maxIdleConnections"`
	IdleConnectionTimeout string `yaml:"idleConnectionTimeout"`
	ExpectContinueTimeout string `yaml:"expectContinueTimeout"`
}

// LoadConfig loads the configuration
func LoadConfig(file string) Config {
	config := Config{}
	config.Request.Transformers = make([]TransformerConfig, 0)
	config.Request.Sidecars = make([]SidecarConfig, 0)
	config.Response.Transformers = make([]TransformerConfig, 0)
	config.Response.Sidecars = make([]SidecarConfig, 0)

	data, err := yamlRef.MergeAndMarshall(file)
	if err != nil {
		log.Fatal("Could not load the configuration file", err, nil)
	}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal("Could not parse the configuration file", err, nil)
	}

	envs := getEnvs()
	for k, v := range config.Variables {
		parsed, _ := Templ(v, envs)
		config.Variables[k] = parsed
	}
	return config
}

// Init initialize the configuration
func (c *Config) Init() {
	for _, topRule := range c.Rules {
		for pattern, rule := range topRule {
			var err error
			rule._pattern, err = regexp.Compile(pattern)
			if err != nil {
				log.Fatal("Pattern is not a valida regex", err, nil)
			}
			rule.Origin, err = Templ(rule.Origin, nil)
			if err != nil {
				log.Fatal("Could not parse origin", err, map[string]interface{}{"origin": rule.Origin})
			}
			mergedReqTransformers := append(c.Request.Transformers, rule.Request.Transformers...)
			reqTrans, err := NewRequestTransformers(&mergedReqTransformers)
			if err != nil {
				log.Fatal("Error initializing request transformers ", err, nil)
			}
			rule.Request._transformers = reqTrans

			mergedResTransformers := append(c.Response.Transformers, rule.Response.Transformers...)
			resTrans, err := NewResponseTransformers(&mergedResTransformers)
			if err != nil {
				log.Fatal("Error initializing response transformers ", err, nil)
			}
			rule.Response._transformers = resTrans

			mergedReqSidecars := append(c.Request.Sidecars, rule.Request.Sidecars...)
			rule.Request._sidecars = NewRequestSidecars(&mergedReqSidecars)

			mergedResSidecars := append(c.Response.Sidecars, rule.Response.Sidecars...)
			rule.Response._sidecars = NewResponseSidecars(&mergedResSidecars)

			if strings.HasPrefix(rule.Origin, "postgres://") ||
				strings.HasPrefix(rule.Origin, "mysql://") {
				databaseUrl, err := dburl.Parse(rule.Origin)
				if err != nil {
					log.Fatal("Could not parse the database URI", err, nil)
				}
				rule.db, err = sqlx.Open(databaseUrl.Driver, databaseUrl.DSN)
				if err != nil {
					log.Fatal("Could not connect to the database", err, nil)
				}
			}
		}
	}
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
