module zap

go 1.23.1

replace github.com/pzx521521/promtail => ../../../promtail

require (
	github.com/pzx521521/promtail v0.0.0-00010101000000-000000000000
	go.uber.org/zap v1.27.0
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/grafana/loki/pkg/push v0.0.0-20241214002250-2de6e16e19f1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241209162323-e6fa225c2576 // indirect
	google.golang.org/grpc v1.69.0 // indirect
	google.golang.org/protobuf v1.35.2 // indirect
)
