package main

import (
	"github.com/pzx521521/promtail"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"os"
)

var ip, _ = LocalIP()
var host, _ = os.Hostname()
var id = ip + "-" + host

func newLogger() (*zap.Logger, *lokipush.LokiPusher) {
	pushURL := "https://logs-prod-020.grafana.net/loki/api/v1/push"
	c := lokipush.NewConfig(pushURL, "promtail_test", lokipush.PUSH_TYPE_GRPC)
	c = c.AddLabels(lokipush.Label{"tag", "tagVal"})
	pusher := lokipush.NewLokiPusher(c)
	zap.NewDevelopmentConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		pusher,
		zapcore.DebugLevel,
	)
	logger := zap.New(core)
	logger = logger.With(zap.String("id", id))
	return logger, pusher
}

func LocalIP() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Name == "lo" {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "", nil
}
