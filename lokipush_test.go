package lokipush

import (
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func getProxyHttpClient() *http.Client {
	proxyURL, _ := url.Parse("http://localhost:8888") //for charles
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client := &http.Client{
		Transport: transport,
	}
	return client
}
func getConfig(pushType int) *Config {
	pushURL := "https://logs-prod-020.grafana.net/loki/api/v1/push"
	c := NewConfig(pushURL, "Promtail_test")
	c.PushType = pushType
	c = c.AddLabels(Label{"tag", "tagVal"})
	c.Client = getProxyHttpClient() // for Charles
	c.UserName = "1066737"
	c.Password = "glc_eyJvIjoiMTI4OTAyNyIsIm4iOiJzdGFjay0xMTA4OTU4LWhtLXJlYWQtcnciLCJrIjoiMjEwc3JFSTFsSjc0Tm1jVHl1YjkwOTVuIiwibSI6eyJyIjoicHJvZC1hcC1zb3V0aGVhc3QtMSJ9fQ=="
	return c
}
func TestHttp(t *testing.T) {
	c := getConfig(PUSH_TYPE_HTTP)
	pusher := NewLokiPusher(*c)
	defer pusher.Close()
	for i := 1; i < 5; i++ {
		pusher.Log("hello http")
		time.Sleep(1 * time.Second)
	}
}

func TestGrpc(t *testing.T) {
	c := getConfig(PUSH_TYPE_GRPC)
	pusher := NewLokiPusher(*c)
	defer pusher.Close()
	for i := 1; i < 5; i++ {
		pusher.Log("hello grpc")
		time.Sleep(1 * time.Second)
	}
}

func TestGzipHttp(t *testing.T) {
	c := getConfig(PUSH_TYPE_HTTP_GZIP)
	pusher := NewLokiPusher(*c)
	defer pusher.Close()
	for i := 0; i < 5; i++ {
		pusher.Log("hello gzip")
	}
}
func TestLog(t *testing.T) {
	c := getConfig(PUSH_TYPE_HTTP_GZIP)
	pusher := NewLokiPusher(*c)
	defer pusher.Close()
	log.Default().SetOutput(pusher)
	for i := 0; i < 5; i++ {
		log.Println("hello gzip form log")
	}
}

func TestSlog(t *testing.T) {
	c := getConfig(PUSH_TYPE_HTTP_GZIP)
	pusher := NewLokiPusher(*c)
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
	//{job="Promtail_test"} | json
}
func TestErrorNet(t *testing.T) {
	c := getConfig(PUSH_TYPE_GRPC)
	c.Password = ""
	c.StdOut = true
	pusher := NewLokiPusher(*c)
	defer pusher.Close()
	for i := 0; i < c.BatchMaxSize*5; i++ {
		pusher.Log("hello gzip" + strconv.Itoa(i))
		time.Sleep(time.Millisecond)
	}
}
