# Templates
You can *"templetize"* configuration values in multiple places throughout the application.
Templates are strings which contain the tag `${...}`. The content of the curly brackets are expressions which allow you
to select data objects from the current scope.

## The Scope
The rule is simple:
* if the component is meant to operate on or modify an API transaction, the transaction itself is the scope
* if the component has more of a general purpose, then the scope is the `variables` section of the configuration
* when using a template in the variables section of the configuration, the scope is made up of environment variables

## The API transaction scope
This is by far the most complicated scope, so we'll dig into it. Here's the content of the API transaction:

* `Request` (field): the request object
  * `Method` (field): the method used to perform the request
  * `GetHeader(name)` (function): will return the value of a request header
  * `ExpandedBody` (field): an array of bytes representing the content of the request body. This field as a value only
    if a transformer or a sidecar had the need to read the request stream
  * `ParsedBody` (field): a data structure that gets populated by the `parser` transformer if the body is a JSON
* `Response` (field):
  * `StatusCode` (field): the response status code
  * `GetHeader(name)` (function): will return the value of a response header
  * `ExpandedBody` (field): an array of bytes representing the content of the response body. This field as a value only
    if a transformer or a sidecar had the need to read the response stream
  * `ParsedBody` (field): a data structure that gets populated by the `parser` transformer if the body is a JSON