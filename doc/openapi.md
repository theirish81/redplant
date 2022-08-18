# OpenAPI support
You can instruct RedPlant on the routes using an OpenAPI v3 specification file.

## Configuration
Once you've placed the OpenAPI spec file(s) in the location that you prefer, reference them at the top level of the
configuration, as in:
```yaml
openAPI:
  'localhost:9001':
      file: config/demoapi.yaml
```
You can reference one file per host. the OpenAPI paths will be converted into RedPlant routes at bootstrap.

## Configuration /2
If you wish to add RedPlant specific configuration details to an OpenAPI spec file, you can do so by using the
`x-redplant` extension. Here's an example:
```yaml
paths:
  '/retail/product':
    get:
      operationId: listProducts
      description: lists all products
      x-redplant:
        response:
          transformers:
            - id: headers
              params:
                set:
                  foo: bar
```

