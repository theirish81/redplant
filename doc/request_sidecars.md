# Request Sidecars
Sidecars to attach and trigger before the request is forwarded to the origin.

## Access Log Sidecar
Logs the inbound requests.

Example:
```yaml
sidecars:
- id: accessLog
  workers: 1
  queue: 2
  block: true
```

This sidecar has no specific parameter.