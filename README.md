# lokipush
[promtail-client](https://github.com/afiskon/promtail-client) has been archived  
I added the following new features:
+ auth
	- [x] BaseAuth by name and password for grafana cloud
+ suport for others log libs
	- [x] slog 
	- [x] zap
	- [x] log(no level default)
	- [x] output both stdout and above libs
+ add push-type by
	- [x] grpc(use org lib, and default)
	- [x] http
	- [x] http-gzip(new)
+ backoff & backup it locally when network error
	- [x] RetryCount & MaxRetryDuration([promtail-client](https://github.com/afiskon/promtail-client) will send always)
	- [x] save it locally when retry failed([promtail-client](https://github.com/afiskon/promtail-client) will sing memory causes memory explosion)
+ loki index(labels) & metadata
	- [x] add labels(loki index)
	- [x] add metadata(grpc only)
# use
```shell
go get github.com/pzx521521/lokipush
```
example for slog:
```go
package main

import (
	"github.com/pzx521521/lokipush"
	"log/slog"
)

func main() {
	pushURL := "https://logs-prod-020.grafana.net/loki/api/v1/push"
	//promtail_test will be a label {"job","promtail_test"} for service_name detect
	c := lokipush.NewConfig(pushURL, "promtail_test")
	//Labels suport
	c = c.AddLabels(lokipush.Label{"tag", "tagVal"})
	//bathauth
	c.UserName = "<your-username>"
	c.Password = "<your-password>"
	//both output with StdOut
	c.StdOut = true
	//PushType defualt is grpc
	c.PushType = lokipush.PUSH_TYPE_HTTP_GZIP
	pusher := lokipush.NewLokiPusher(*c)
	defer pusher.Close()
	slog.SetDefault(slog.New(slog.NewJSONHandler(pusher, &slog.HandlerOptions{
		Level: slog.LevelDebug})))

	slog.Info("hello slog, this is a info msg",
		slog.Int64("id", 123),
		slog.String("name", "slog"))

	slog.With(slog.String("WithKey", "WithVal")).Debug(
		"hello slog, this is a debug msg",
		slog.Int64("id", 123),
		slog.String("name", "slog"))

	slog.Warn("hello slog, this is a warn msg",
		slog.Int64("id", 123),
		slog.String("name", "slog"))

	group := slog.Default().WithGroup("group")
	group.Error("hello slog, this is a error msg",
		slog.Int64("id", 123),
		slog.String("name", "slog"))
	//do something...

	//Loki QL
	// {job="Promtail_test"} | json
}


```
example for zap:  
`exmple/zap`  
example for log:  
`lokipush_test.go/TestLog`  

# reference
## [loki-grpc reference](https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs): 
grpc The default behavior is for the POST body to be a Snappy-compressed Protocol Buffer message:  
These POST requests require the Content-Type HTTP header to be application/x-protobuf
# 
`/loki/api/v1/push` is the endpoint used to send log entries to Loki. The default behavior is for the POST body to be a [Snappy](https://github.com/google/snappy)-compressed [Protocol Buffer](https://github.com/protocolbuffers/protobuf) message:

- [Protocol Buffer definition](https://github.com/grafana/loki/blob/main/pkg/logproto/logproto.proto)
- [Go client library](https://github.com/grafana/loki/blob/main/clients/pkg/promtail/client/client.go)
- [proto push code](https://github.com/grafana/loki/blob/2de6e16e19f1f011fc8b52f493a298ad750e8c64/pkg/pattern/aggregation/push.go#L193)
- [proto real path](https://github.com/grafana/loki/blob/2de6e16e19f1f011fc8b52f493a298ad750e8c64/pkg/push/push.proto)

## loki-http-api reference: [Ingest logs](https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs)

```bash
POST /loki/api/v1/push
```

`/loki/api/v1/push` is the endpoint used to send log entries to Loki. The default behavior is for the POST body to be a [Snappy](https://github.com/google/snappy)-compressed [Protocol Buffer](https://github.com/protocolbuffers/protobuf) message:

- [Protocol Buffer definition](https://github.com/grafana/loki/blob/main/pkg/logproto/logproto.proto)
- [Go client library](https://github.com/grafana/loki/blob/main/clients/pkg/promtail/client/client.go)

These POST requests require the `Content-Type` HTTP header to be `application/x-protobuf`.

Alternatively, if the `Content-Type` header is set to `application/json`, a JSON post body can be sent in the following format:

```json
{
  "streams": [
    {
      "stream": {
        "label": "value",
        "anotherLabel": "anotherValue"
      },
      "values": [
          [ "<unix epoch in nanoseconds>", "<log line>" ],
          [ "<unix epoch in nanoseconds>", "<log line>" ]
      ]
    }
  ]
}
```

You can set `Content-Encoding: gzip` request header and post gzipped JSON.

You can optionally attach [structured metadata](https://grafana.com/docs/loki/latest/get-started/labels/structured-metadata/) to each log line by adding a JSON object to the end of the log line array. The JSON object must be a valid JSON object with string keys and string values. The JSON object should not contain any nested object. The JSON object must be set immediately after the log line. Here is an example of a log entry with some structured metadata attached:


```json
"values": [
    [ "<unix epoch in nanoseconds>", "<log line>", {"trace_id": "0242ac120002", "user_id": "superUser123"}]
]
```

In microservices mode, `/loki/api/v1/push` is exposed by the distributor.

If [`block_ingestion_until`](https://grafana.com/docs/loki/latest/configuration/#limits_config) is configured and push requests are blocked, the endpoint will return the status code configured in `block_ingestion_status_code` (`260` by default) along with an error message. If the configured status code is `200`, no error message will be returned.

## Examples

The following cURL command pushes a stream with the label “foo=bar2” and a single log line “fizzbuzz” using JSON encoding:

```bash
curl -H "Content-Type: application/json" \
  -s -X POST "http://localhost:3100/loki/api/v1/push" \
  --data-raw '{"streams": [{ "stream": { "foo": "bar2" }, "values": [ [ "1570818238000000000", "fizzbuzz" ] ] }]}'
```
