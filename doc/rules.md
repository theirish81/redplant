# Rules
At the very top level, we find the "domain" selection,
which is a regular expression for domains, so: `localhost` is fine as much as `.*.foobar.com` is.

## the paths
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

## request
A collection of request transformers and sidecars which apply to this specific route.

## response
A collection of response transformers and sidecars which apply to this specific route.


### transformers
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
id: basic-auth
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
* `logging` (map[string,any],optional): specific logging to be used for this transformer. See [logging](./logging.md#dedicated-sidecartransformer-logging)

**Check the [request transformers documentation](./request_transformers.md)**

**Check the [response transformers documentation](./response_transformers.md)**

### sidecars
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
  uri: "${Variables.CAPTURE_URI}"
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
* `logging` (map[string,any],optional): specific logging to be used for this sidecar. See [logging](./logging.md#dedicated-sidecartransformer-logging)

**Check the [request sidecars documentation](./request_sidecars.md)**

**Check the [response sidecars documentation](./response_sidecars.md)**