# Response Transformers
Response transformers will transform the response received from the origin, before forwarding it to the client.

## Header Transformer
Adds, removes and sets custom headers.

example:
```yaml
transformers:
- id: headers
params:
  set:
    X-Proxied: "true"
  remove:
  - User-Agent
```

params:
* `set` (map[string,string],optional): key/values of headers to set
* `remove` (list[string],optional): headers to remove by name

## Barrage Transformer
Will immediately drop the response in case certain conditions are met.

Example:
```yaml
transformers:
- id: barrage
  params:
    bodyRegexp: .*log4j.*
```

params:
* `bodyRegexp` (string,optional): checks the body with a regular expression. If true, will block the request
* `headerNameRegexp` (string,optional): checks the header names with a regular expression. If true, will block the request
* `headerValueRegexp` (string,optional): checks the header values with a regular expression. If true, will block the request
* `headerRegexp` (string,optional): checks the headers `key=value` pairs with a regular expression. If true, will block the request


## Delay Transformer
Will cause a delay in the prosecution of the response.
```yaml
- id: delay
  params:
    min: 1s
    max: 5s
```

params:
* `min` (string/duration,required): the minimum delay being applied to the response
* `max` (string/duration,required): the maximum delay being applied to the response

## Scriptable Transformer
Will apply a transformation described in a JavaScript file.

Example:
```yaml
transformers:
- id: scriptable
  params:
    path:
      etc/scripts/gino.js
```

An example of the script would be:
```javascript
wrapper.Response.Header.Set("gino","pino")
true
```
**Note:** the bool return value is essential, as this transformer can be used to either change the response envelope
or block a response. Return `true` if the flow should continue.

params:
* `path` (string,required): path to a JavaScript script

## Tag Transformer
Will add a tag to the response envelope. Following transformers and sidecars can then be activated if a tag is present.

Example:
```yaml
transformers:
- id: tag
  params:
    tags:
      - db
```
params:
* `tags` (array[string],required): a list of tags to apply to the request.