package main

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

func main() {
	zlog, pusher := newLogger()
	defer func(zlog *zap.Logger) {
		err := zlog.Sync()
		if err != nil {
			fmt.Printf("%v\n", err)
		}
		pusher.Close()
	}(zlog)
	zlog.Info("hello slog, this is a info msg",
		zap.Int64("id", 123),
		zap.String("name", "slog"))

	zlog.Debug(
		"hello slog, this is a debug msg",
		zap.Int64("id", 123),
		zap.String("name", "slog"))

	zlog.Warn("hello slog, this is a warn msg",
		zap.Int64("id", 123),
		zap.String("name", "slog"))

	zlog.Error("hello slog, this is a error msg",
		zap.Int64("id", 123),
		zap.String("name", "slog"))
	//do something...
	time.Sleep(time.Second)
	//{job="Promtail_test"} | json

}
