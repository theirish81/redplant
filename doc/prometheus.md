# Prometheus support
Support for prometheus metrics is still in its infancy, but will grow.

## Configuration
In the main config file, add the following object:
```yaml
prometheus:
  port: 9002
  path: "/metrics"
```
This will expose Prometheus metrics on port 9002, path: `/metrics`