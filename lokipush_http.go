/*
## [Ingest logs](https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs)
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
        "label": "value"
      },
      "values": [
          [ "<unix epoch in nanoseconds>", "<log line>" ],
          [ "<unix epoch in nanoseconds>", "<log line>" ]
      ]
    }
  ]
}
```
*/

package lokipush

import (
	"bytes"
	"encoding/json"
	"github.com/grafana/loki/pkg/push"
	"strconv"
)

type streamString string

func (s *streamString) MarshalJSON() ([]byte, error) {
	return []byte(string(*s)), nil
}

type streamValues []push.Entry

func (svs *streamValues) MarshalJSON() ([]byte, error) {
	//[["1734135699246000000", "{\"level\":\"info\"}"]]
	ret := bytes.Buffer{}
	ret.WriteRune('[')
	for _, entry := range *svs {
		//timestamp
		ret.WriteString(`["`)
		ret.WriteString(strconv.FormatInt(entry.Timestamp.UnixNano(), 10))
		ret.WriteString(`",`)
		//line
		marshal, err := json.Marshal(entry.Line)
		if err != nil {
			return nil, err
		}
		ret.Write(marshal)
		ret.WriteString(`],`)
	}
	b := ret.Bytes()
	b[len(b)-1] = ']'
	return b, nil
}

type stream struct {
	Stream streamString `json:"stream"`
	Values streamValues `json:"values"`
}

type lokiHttpPushRequest struct {
	Streams []stream `json:"streams"`
}
