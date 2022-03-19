# Response Sidecars
Sidecars to attach and trigger right after the origin responded.

## Access Log Sidecar
Logs the requested origin URL and the returned response status.

Example:
```yaml
sidecars:
- id: accessLog
  workers: 1
  block: true
```

This sidecar has no specific parameter.

## Metrics Log Sidecar
Logs a recording of all the metrics involved in the request/response cycle.

Example:
```yaml
sidecars:
- id: metricsLog
  workers: 1
  block: true
```

This sidecar has no specific parameter.


## Capture Sidecar
Captures an API conversation (both request and response), marshals it in a JSON format and either stores it
to a file or sends it via HTTP.

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
* `uri` (string/uri,required): the URI to send the marshaled conversation to. It can be either http(s):// or file://
* `timeout` (string/duration,optional): in case the URI is HTTP(s), determines what's the timeout
* `headers` (map[string,string],optional): in case the URI is HTTP(s), send these extra headers
* `requestContentTypeRegexp` (string/regexp,optional): capture the conversation only if the request content type
  matches this regular expression
* `responseContentTypeRegexp` (string/regexp,optional): capture the conversation only if the response content type
  matches this regular expression
