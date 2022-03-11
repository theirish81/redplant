# Websocket Origin (experimental)
Our support of websocket is bare bones and experimental.

## Setup
Configure your route as follows:
```yaml
"/ws":
    origin: wss://example/v3/channel_1
    stripPrefix: /ws
```
Where the value of `origin` is your websocket URL. In the example we're also stripping out `/ws` that would be
otherwise appended.

## Usage
This tripper will **blindly** proxy websockets connections. The request transformations will apply for establishing
the connection, but the messages travelling in the websocket will be left untouched and unobserved. There's a plan
to create transformers dedicated to websockets, but they do not exist at the moment.