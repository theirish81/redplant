# File Origin
The file origin turns a route into an HTTP static file server.

## Setup
```yaml
"/fs/{fn}":
  origin: file://files            # Path to the files directory
  stripPrefix: /fs                # We're stripping the "/fs" from the final URL
```
This will allow the `fs` route to serve files hosted in the `files` directory. The variable in the URL has no specific
purpose other than route-matching, so it can be named anything. However, this route will match a file only, not allowing
directory diving. If you plan on allowing the consumer to dive into all sorts of subdirectories you can always shape
the pattern as `/fs/{fn:.*}`.

This origin is treated no differently from an actual HTTP origin, so all transformations and sidecars can apply.
