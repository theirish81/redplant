# Response Sidecars
Sidecars to attach and trigger right after the origin responded.

## Access Log Sidecar
Logs the requested origin URL and the returned response status.

Example:
```yaml
sidecars:
- id: access-log
  workers: 1
  block: true
```

This sidecar has no specific parameter.

## Metrics Log Sidecar
Logs a recording of all the metrics involved in the request/response cycle.

Example:
```yaml
sidecars:
- id: metrics-log
  workers: 1
  block: true
```

This sidecar has no specific parameter. However, logging configuration will apply to determine where and how the metrics
will be logged.



## Capture Sidecar
Captures an API conversation (both request and response), marshals it in a JSON format and either stores it  as a log
entry or sends it via HTTP. For the logging mechanism, the sidecar relies on the
[standard RedPlant logging mechanism](./logging.md).

Example:
```yaml
sidecars:
- id: capture
  workers: 2
  block: false
  params:
    uri: "https://example.com"
    timeout: 10s
    headers:
      foo: "bar"
    requestContentTypeRegexp: '(^$|.*json.*)'
    responseContentTypeRegexp: '.*json.*'
```

params:
* `uri` (string/uri,optional): the URI to send the marshaled conversation to, if you're going to use an HTTP endpoint
  as a destination. If not provided, the system will default to the logging mechanism
* `timeout` (string/duration,optional): in case the URI is HTTP(s), determines what's the timeout
* `headers` (map[string,string],optional): in case the URI is HTTP(s), send these extra headers
* `requestContentTypeRegexp` (string/regexp,optional): capture the conversation only if the request content type
  matches this regular expression
* `responseContentTypeRegexp` (string/regexp,optional): capture the conversation only if the response content type
  matches this regular expression
