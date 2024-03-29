# Prometheus support
Support for Prometheus metrics is evolving and not yet to its final state.

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
- id: metrics-log
  logging:
   prometheus:
    enabled: true
    prefix: test
```
This will enable Prometheus in `metrics-log` which will start publishing its own metrics.
The `prefix` field will prepend a string to the name of the `summary` so that you can better distinguish your series,
but it's totally optional.

## Metrics exposed by component
Not all components will publish Prometheus metrics. Here's an incomplete list of which metrics will be published
if you enable the integration.
More will be added in the future.

### metrics-log
* `transaction` : summary
* `req_transformation` : summary
* `res_transformation` : summary
* `res_transformation` : summary

### access-log (request)
* `request_access` : counter

### access-log (response)
* `upstream_access` : counter

### basic-auth
* `basic_auth_denied`: counter

### cookie-to-token-auth
* `cookie_to_token_auth_denied`: counter

### jwt-auth
* `jwt_auth_denied`: counter

### barrage
* `barraged`: counter

### openapi-validator
* `openapi_validation_failed`: counter

### rate-limiter
* `rate_limited`: counter