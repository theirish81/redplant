# Request Transformers
Request transformers will transform the inbound request before forwarding it to the origin.

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

## Basic Auth Transformer
Validates basic authentication. You can provide username and password as params (see example) also using variables.
Alternatively, you can use the `htpasswd` param with the path to a htpasswd file.

Example:
```yaml
transformers:
- id: basicAuth 
params:
  username: "username"
  password: "password"
```

## JWT Auth Transformer
Will block any request without a Bearer token or a token whose signature cannot be verified. In addition,
it will store claims in the scope of the request, as the `Claims` variable.

Example:
```yaml
transformers:
- id: jwtAuth 
params:
  key: "some_bytes_here"
```
params:
* `key` (string,optional): you can pass the public key to decode the token in the form of a string
* `pem` (string,optional): path to a public key file

## JWT Sign Transformer
Will add a JWT token to the request in the `Authorization` header.

Example:
```yaml
transformers:
- id: jwtSign
  params:
    pem: /etc/secrets/jwt-public-key/privateKey
```
params:
* `key` (string,optional): you can pass the private key to sign the token in the form of a string
* `pem` (string,optional): path to a private key file
* `claims` (map[string,any],optional): a map of key values representing the claims
* `existingClaims` (bool,optional): if set to true, it will expect claims will be present in the request scope
  (set by jwtAuth) and will produce a token with those claims. This is useful to **re-sign** a token

## Barrage Transformer
Will immediately drop the inbound request in case certain conditions are met.

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

## Rate Limiter Transformer
With the help of a **Redis** server, will rate limit the inbound requests based on certain criteria.

Example:
```yaml
transformers:
- id: rate-limiter
  params:
    redisUri: "redis://:password123@127.0.0.1:6379/1"
    vary: "{{.Request.Header.Get \"Username\"}}"
    limit: 5
    range: 1m
```

params:
* `redisUri` (string,required): the URI to a Redis server
* `vary` (string,required): a string (typically an expression) representing the criteria against which the rate limited
  is applied
* `limit` (int,required): the number of allowed requests
* `range` (string/duration,required): the time frame in which the limit is enforced

## Tag Transformer
Will add a tag to the request envelope. Following transformers and sidecars can then be activated if a tag is present.

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


## Delay Transformer
Will cause a delay in the prosecution of the request.
```yaml
- id: delay
  params:
    min: 1s
    max: 5s
```

params:
* `min` (string/duration,required): the minimum delay being applied to the request
* `max` (string/duration,required): the maximum delay being applied to the request

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
wrapper.Request.Header.Set("gino","pino")
true
```
**Note:** the bool return value is essential, as this transformer can be used to either change the request envelope
or block a request. Return `true` if the flow should continue.

params:
* `path` (string,required): path to a JavaScript script