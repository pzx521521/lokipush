package lokipush

import (
	"net/http"
	"time"
)

const (
	PUSH_TYPE_GRPC = iota //default
	PUSH_TYPE_HTTP_GZIP
	PUSH_TYPE_HTTP
)

type Config struct {
	PushURL          string
	UserName         string
	Password         string
	BatchWait        time.Duration
	MaxRetryDuration time.Duration
	RetryCount       int
	BatchMaxSize     int
	Client           *http.Client
	StdOut           bool
	PushType         int
	labels           Labels
	labelStr         []byte
}

func NewConfig(pushURL, serviceName string) *Config {
	batchWait := 5 * time.Second
	conf := Config{
		PushURL:          pushURL,
		BatchWait:        batchWait,
		BatchMaxSize:     1000,
		MaxRetryDuration: time.Hour,
		RetryCount:       3,
		Client:           &http.Client{Timeout: batchWait},
		PushType:         PUSH_TYPE_GRPC,
	}
	return conf.AddLabels(Label{"job", serviceName})
}
func (c *Config) AddLabels(labels ...Label) *Config {
	c.labels = append(c.labels, labels...)
	if c.PushType == PUSH_TYPE_GRPC {
		c.labelStr = []byte(c.labels.String())
		return c
	}
	c.labelStr = []byte(c.labels.HttpString())
	return c
}
