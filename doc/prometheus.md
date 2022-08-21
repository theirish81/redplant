# Prometheus support
Support for prometheus metrics is still in its infancy, but will grow.

## General configuration
In the main config file, add the following object:
```yaml
prometheus:
  port: 9002
  path: "/metrics"
```
This will expose Prometheus metrics on port 9002, path: `/metrics`


## Sidecar / Transformer configuration
As a default, RedPlant will only publish the application performance metrics, but more metrics are available by
configuring Prometheus at the sidecar and transformer level. By enabling Prometheus in these components, they will
start publishing metrics on their activity.

In a sidecar or transformer configuration, add or extend the logging section, as in:

```yaml
- id: metricsLog
  logging:
   prometheus:
    enabled: true
    prefix: test
```
This will enable Prometheus in `metricsLog` which will start publishing its own metrics.
The `prefix` field will prepend a string to the name of the `summary` so that you can better distinguish your series,
but it's totally optional.

## Metrics exposed by component

### metricsLog
* `transaction` : summary
* `req_transformation` : summary
* `res_transformation` : summary
* `res_transformation` : summary

### accessLog (request)
* `request_access` : counter

### accessLog (response)
* `upstream_access` : counter

### basicAuth
* `basic_auth_denied`: counter

### cookieToTokenAuth
* `cookie_to_token_auth_denied`: counter

### jwtAuth
* `jwt_auth_denied`: counter

### barrage
* `barraged`: counter

### openapi_validator
* `openapi_validation_failed`: counter

### rate-limiter
* `rate_limited`: counter