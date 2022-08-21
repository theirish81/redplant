# RedPlant

## What is RedPlant?

RedPlant is a **reverse proxy** dedicated to **APIs**. I know what you're thinking, there already are so many, so how is
RedPlant different?

### Dedicated to developers
The project is meant to become a swiss-army knife for API development and debugging. 
The modular system of the tool allows developers to build sophisticated transformation pipelines and sidecar tasks,
with a simple yet effective definition system.

Moreover, the availability of modules that stretch beyond the reverse proxy duties, such as the DB->JSON tripper, turn
RedPlant into a trusted testing / CI companion.

### Capable of production traffic
Contrary to most dev-oriented RPs, RedPlant is fully capable of sustaining production-sized traffic. True, not as fast
as NGINX, yet given the complexity of the pipelines you can build, the tradeoff may be worth the tiny extra lag

### Designed to democratise gateways
In large scale microservices architectures, isolation and black-boxing between areas of the software is crucial to
support a sustainable growth. Sometimes though this is hard to achieve and too complex RPs either induce DevOps to
not fully leverage the software potential, or even discourage the use of one at all. RedPlant aims to equip devs with
a vast array of tools, while maintaining a low complexity.

## Basic configuration
You can check out an example of the configuration in the `etc` directory.

In short, you have two top level configuration files:

### logging.yaml
This file will configure the logging mechanism of the system. The file is not required and if absent, defaults will be
used instead. The configuration is very simple, here's an example:

```yaml
level: INFO
format: simple
path: "my_logs.log"
```

* `level`: determines the log level. Valid values are `DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL` (default: `INFO`)
* `format`: the log format. Valid values are `simple`, `JSON` (default: `simple`)
* `path`: where the logs should be stored. Valid values are `path_to_file.log` or empty (default: empty, logs to STD out)

To select a configuration file pass the `-l path_to_logging.yaml` option to the launch command.

All defined sidecars and transformers will inherit the default logger. You can, however, specify different behaviours
for each component. For example:

```yaml
- id: basicAuth
  logging:
    path: auth.log
    level: warn
    format: JSON
```

Therefore, you can have a very fine grained logging strategy by potentially defining your logging needs in each
component.

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
can also evaluate Go templates to include environment variables. Example:

```yaml
UN: foo
PW: bar
SERVER_NAME: "{{.SERVER_NAME}}"
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

#### before
Each route can declare **transformers** and **sidecars** (we'll discuss them later). However, global transformers and
sidecars can be applied at a global level.
The `before` section describes transformers and sidecars to be applied **before** the specific route's transformers and
sidecars are applied.

#### after
The `after` section describes transformers and sidecars to be applied **after** the specific route's transformers and
sidecars are applied.

#### rules
Rules describe the routes this system will take care of, and how. At the very top level, we find the "domain" selection,
which is a regular expression for domains, so: `localhost` is fine as much as `.*.foobar.com` is.

#### the paths
Direct descendent of the `rules` object, we find paths. Paths are regular expressions as well and describe the path of
the inbound request URLs. The instructions within a path are:
* `origin` (required): a URL representing the origin we should forward the request to
* `stripPrefix` (optional): inbound paths are generally appended as they are to the origin. For example, if the path is
`/foo/abc123` and the origin is `http://example.com/data`, the request will be forwarded to `http://example.com/data/foo/abc123`.
However, this may not be the desired behavior. If we wanted to forward to `http://example.com/data/abc123`, then we give
the `stripPrefix` parameter the value of `/foo`.

In our example, a set of rules with paths will look like this:
```yaml
localhost:9001:
  "/todo/.*":
    origin: https://jsonplaceholder.typicode.com/todos
    stripPrefix: /todo
```

### Paths with explicit method
In some situations you may want to describe substantially different behaviours for different methods.
In this case you can explicitly declare the method you're describing in the path pattern as in:
```yaml
"[get] /todo/.*":
   origin: https://jsonplaceholder.typicode.com/todos
   stripPrefix: /todo
```

#### request
A collection of request transformers and sidecars which apply to this specific route.

### response
A collection of response transformers and sidecars which apply to this specific route.


## Transformers
Transformers are plugins you can apply to a request or a response. Obviously, due to their nature of modifying the
content, transformers are blocking and part of the content negotiation.

**NOTE**: some transformers can be applied to requests and responses alike, others are specialized.

Depending on the purpose of the transformer, it can do two things:
* transform the request or the response by changing URLs, headers and bodies
* alter the flow of the request or the response, by stopping, rejecting or delaying

Transformers can be found in the global `before.request.transformers` and `after.request.transformers` sections or
in the specific route `request.transformers` and `response.transformers`. The configuration is made of two parts,
a global definition that defines the type of transformer and its general behavior, and its specific params.
Example:
```yaml
id: basicAuth
activateOnTags:
  - db
  - fs
params:
  retain: false
  htpasswd: etc/passwords
```

* `id` (string,required): defines the type of transformer
* `activateOnTags` (array[string],optional): if the transaction is tagged with one of these tags, then the transformer will trigger,
otherwise it will not be applied
* `params` (map[string,any],required): the transformer's specific parameters

**Check the [request transformers documentation](./doc/request_transformers.md)**

**Check the [response transformers documentation](./doc/response_transformers.md)**

## Sidecars
Sidecars are operations that will feed on the request/response data, but do not alter the content of the transferred
data. Therefore, they can be executed concurrently to the main data flow.

Sidecars can be found in the global `before.request.sidecars` and `after.request.sidecars` sections or
in the specific route `request.sidecars` and `response.sidecars`. The configuration is made of two parts,
a global definition that defines the type of sidecar and its general behavior, and its specific params.
Example:

```yaml
id: capture
workers: 2
queue: 5
block: true
params:
  uri: "{{.Variables.CAPTURE_URI}}"
  responseContentTypeRegexp: '.*json.*'
  requestContentTypeRegexp: '(^$|.*json.*)'
  format: JSON
```
* `id` (string,required): the name of the sidecar
* `workers` (int,optional): the number of instances of this sidecar (default: 1)
* `queue` (int,optional): the size of the queue for the workers. Meaningful in conjunction with `block` (default: 1)
* `block` (bool,optional): if `true`, the lack of available workers (as in: all busy) combined with a full queue, 
will block the main data flow. This is useful when resources are limited, and we want to avoid a boundless escalation
of used resources (default: false)
* `dropOnOverflow` (bool,optional): if `true`, in case of a full queue, new messages to the sidecars will be dropped until
a slot is freed in the queue.  The combination of `block=false` and `dropOnOverflow=true` puts a hard cap on resource
usage for sidecars, while not limiting the performance of API transactions
* `activateOnTags` (array[string],optional): if the transaction is tagged with one of these tags, then the sidecar will trigger,
  otherwise it will not be applied
* `params`(map[string,any],required): the sidecar's specific parameters

**Check the [request sidecars documentation](./doc/request_sidecars.md)**

**Check the [response sidecars documentation](./doc/response_sidecars.md)**


## Exotic origins
RedPlant can accept (more or less) exotic origins. See:
* [DB Origin](./doc/db.md) : access your databases with API calls
* [Websocket Origin](./doc/websocket.md) : I know, this is not really exotic, but it's currently in the experimental stage

## Observability
Check our [Prometheus metrics support](./doc/prometheus.md)


## OpenAPI support
OpenAPI specification files can be used to instruct RedPlant about the routes, please refer to the
[OpenAPI support](./doc/openapi.md) section