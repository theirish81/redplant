# Logging
Logging is a key part of any system, but particularly for RedPlant. As one of the objectives of the platform is
giving you insights in what's going on, logging must be fine grained enough to give you exactly what you need, with
the level of details you need.

## logging.yaml (global configuration)

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


## Dedicated sidecar/transformer logging

All defined sidecars and transformers will inherit the default logger. You can, however, specify different behaviours
for each component. For example:

```yaml
- id: basic-auth
  logging:
    path: auth.log
    level: warn
    format: JSON
```

Therefore, you can have a very fine-grained logging strategy by potentially defining your logging needs in each
component.