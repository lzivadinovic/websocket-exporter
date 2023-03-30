# Prometheus websocket exporter
Websocket exporter that checks if ws is responding with specific message

Based on https://github.com/sahandhabibi/Websocket-exporter 

### Usage
To probe exporter send GET request with target, message and contains parameter, eg:

```bash
curl localhost:9143/probe?target=wss://endpoint&message=something&contains=test

# response should be something like

# HELP websocket_response_time ( Time until we get EOSE in ms; 0 for failed )
# TYPE websocket_response_time gauge
websocket_response_time 63
# HELP websocket_status_code ( 101 is normal status code for ws )
# TYPE websocket_status_code gauge
websocket_status_code 101
# HELP websocket_successful ( 0 = false , 1 = true )
# TYPE websocket_successful gauge
websocket_successful 1
```


TODO:
- Docs
- Test cases if there is no contains or message request (ping pong)
