# RedPlant

![RedPlant](./doc/redplant_small.png "RedPlant")

## Status
[![CircleCI](https://dl.circleci.com/status-badge/img/gh/theirish81/redplant/tree/main.svg?style=svg)](https://dl.circleci.com/status-badge/redirect/gh/theirish81/redplant/tree/main)

## What is RedPlant?

RedPlant is a **reverse proxy** dedicated to **APIs**. I know what you're thinking, there already are so many, so how is
RedPlant different?

### Dedicated to developers
The project is meant to become a swiss-army knife for API development and debugging, both in a test and production
environment.
The modular system of the tool allows developers to build sophisticated transformation pipelines and sidecar tasks,
with a simple yet effective definition system.

Moreover, the availability of modules that stretch beyond the reverse proxy duties, such as the DB->JSON tripper, turn
RedPlant into a trusted testing / CI companion.

### Capable of production traffic
Contrary to most dev-oriented RPs, RedPlant is fully capable of sustaining production-sized traffic. True, not as fast
as NGINX, yet given the complexity of the pipelines you can build, the tradeoff may be worth the tiny extra lag.

### Designed to democratise gateways
In large scale microservices architectures, isolation and black-boxing between areas of the software is crucial to
support a sustainable growth. Sometimes though this is hard to achieve and too complex RPs either induce DevOps to
not fully leverage the software potential, or even discourage the use of one at all. RedPlant aims to equip devs with
a vast array of tools, while maintaining a low complexity.

## Configuration
You can check out an example of the configuration in the `etc` directory.

In short, you have two top level configuration files:

### logging.yaml
Please check our [logging guide](./doc/logging.md)

### config.yaml
This file is where all the magic happens.

**NOTE:** the config.yaml file can reference other YAML files so that you can split the configuration in multiple files
to improve readability. The way you reference another file is the following:
```yaml
key: "$ref:file://another_file.yaml"
```
The content of `another_file.yaml` will become the value of `key`. You can also reference a specific root level object
in another file by using the following notation:
```yaml
key: "$ref:file://another_file.yaml?comp=network"
```

#### variables
In this section you can declare global variables which can be referenced anywhere in the system. Values in this section
can also evaluate [templates](doc/templates.md) to include environment variables. Example:

```yaml
UN: foo
PW: bar
SERVER_NAME: "${SERVER_NAME}"
CAPTURE_URI: file://etc/capture.log
```

In this example `SERVER_NAME` will acquire the value of the environment variable `SERVER_NAME`.

#### network
This section describes everything that has to with both upstream and downstream networking. Example:
```yaml
upstream:
  timeout: 30s
  keepAlive: 30s
  maxIdleConnections: 100
  idleConnectionTimeout: 90s
  expectContinueTimeout: 1s
downstream:
  port: 9001
  tls:
    - host: localhost
      key: etc/server.key
      cert: etc/server.crt
```

#### rules
Rules describe the routes this system will take care of, and how.
**Check the [rules documentation](./doc/rules.md)**


#### before
Each route can declare **transformers** and **sidecars** for both requests and responses. However, global transformers
and sidecars can be applied at a global level. The `before` section describes transformers and sidecars to be applied
**before** the specific route's transformers and sidecars.

**Check the [transformers section in "rules"](./doc/rules.md#transformers)**

**Check the [sidecars section in "rules"](./doc/rules.md#sidecars)**

#### after
The `after` section describes transformers and sidecars to be applied **after** the specific route's transformers and
sidecars.

**Check the [transformers section in "rules"](./doc/rules.md#transformers)**

**Check the [sidecars section in "rules"](./doc/rules.md#sidecars)**

### Templates
It is very useful to reference variables throughout the configuration. Some variables may be evaluated at bootstrap
some others may depend on the API transaction being processed.

Check out [Templates documentation](./doc/templates.md)

### Exotic origins
RedPlant can accept (more or less) exotic origins. See:
* [DB Origin](./doc/db.md) : access your databases with API calls
* [File Origin](./doc/file.md) : transform a route into a static file server
* [Websocket Origin](./doc/websocket.md) : I know, this is not really exotic, but it's currently in the experimental stage

### Observability
Check our [Prometheus metrics support](./doc/prometheus.md)


### OpenAPI support
OpenAPI specification files can be used to instruct RedPlant about the routes, please refer to the
[OpenAPI support](./doc/openapi.md) section


## Running on Docker
Please refer to the [Docker documentation](./doc/docker.md).