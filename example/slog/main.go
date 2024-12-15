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
	c.AddLabels(lokipush.Label{"tag", "tagVal"})
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
