// https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs
// https://github.com/grafana/loki/pkg/pattern/aggregation/push.go#L193
package lokipush

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/push"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

type LokiPusher struct {
	log        *slog.Logger
	config     *Config
	file       *os.File
	fileOnce   sync.Once
	quit       chan struct{}
	quitOnce   sync.Once //防止多次close
	bf         *Backoff  //在发生错误的情况下 防止数据堆积
	running    sync.WaitGroup
	entries    chan *push.Entry
	batchEntry []push.Entry
}

func NewLokiPusher(conf Config) *LokiPusher {
	bf := New(context.Background(), BackOffConfig{
		MinBackoff: conf.BatchWait,
		MaxBackoff: conf.MaxRetryDuration, //失败之后尝试按时间递增尝试3次,全部失败最少10分钟尝试一下
		MaxRetries: conf.RetryCount,
	})
	lp := LokiPusher{
		log:     slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		config:  &conf,
		bf:      bf,
		quit:    make(chan struct{}),
		entries: make(chan *push.Entry, 1), //防止post 时间过长时  chan阻塞
	}
	lp.running.Add(1)
	go lp.run()
	return &lp
}

func (lp *LokiPusher) Write(p []byte) (n int, err error) {
	lp.Log(string(p))
	return len(p), nil
}
func (lp *LokiPusher) Log(logLine string) {
	if lp.config.StdOut {
		os.Stdout.WriteString(logLine)
	}
	lp.entries <- &push.Entry{
		Timestamp: time.Now(),
		Line:      logLine,
	}
}

// for zap
func (lp *LokiPusher) Sync() error {
	return lp.send()
}

func (lp *LokiPusher) Close() {
	lp.quitOnce.Do(func() {
		close(lp.quit)
	})
	lp.running.Wait()
}

func (lp *LokiPusher) getPushReqBuffer() ([]byte, error) {
	switch lp.config.PushType {
	case PUSH_TYPE_GRPC:
		return lp.createProtobufRequest()
	case PUSH_TYPE_HTTP:
		return lp.createJSONRequest()
	case PUSH_TYPE_HTTP_GZIP:
		return lp.createGzipJSONRequest()
	default:
		return nil, fmt.Errorf("unsupported push type: %v", lp.config.PushType)
	}
}

func (lp *LokiPusher) createProtobufRequest() ([]byte, error) {
	pushReq := push.PushRequest{
		Streams: []push.Stream{{
			Labels:  string(lp.config.labelStr),
			Entries: lp.batchEntry,
		}},
	}
	data, err := proto.Marshal(&pushReq)
	if err != nil {
		return nil, err
	}
	return snappy.Encode(nil, data), nil
}

func (lp *LokiPusher) createJSONRequest() ([]byte, error) {
	pushReq := lokiHttpPushRequest{
		Streams: []stream{{
			Stream: streamString(lp.config.labelStr),
			Values: streamValues(lp.batchEntry),
		}},
	}
	return json.Marshal(pushReq)
}

func (lp *LokiPusher) createGzipJSONRequest() ([]byte, error) {
	pushReq := lokiHttpPushRequest{
		Streams: []stream{{
			Stream: streamString(lp.config.labelStr),
			Values: streamValues(lp.batchEntry),
		}},
	}
	buf := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(buf)
	if err := json.NewEncoder(gz).Encode(pushReq); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (lp *LokiPusher) initLogFile() {
	var err error
	lp.file, err = os.OpenFile("unupload.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o666)
	if err != nil {
		lp.log.Warn("open unupload.log failed", slog.Any("error", err))
	}
}
func (lp *LokiPusher) saveToFile() {
	lp.fileOnce.Do(lp.initLogFile)
	if lp.file == nil {
		return
	}
	writer := bufio.NewWriter(lp.file)
	defer writer.Flush()
	for _, entry := range lp.batchEntry {
		_, err := writer.WriteString(entry.Line)
		if err != nil {
			lp.log.Warn("file.Write err", slog.Any("error", err))
			return
		}
		writer.WriteRune('\n')
	}
}
func (lp *LokiPusher) sendAndResetBatch(t *time.Ticker) {
	if !lp.bf.Ongoing() {
		if t != nil && len(lp.batchEntry) < 2*lp.config.BatchMaxSize {
			return
		}
		lp.log.Warn("batch size is too large when Backoff Retry Err,save to file and resetting batch to 0",
			slog.Int("batch_size", len(lp.batchEntry)))
		lp.saveToFile()
		lp.batchEntry = lp.batchEntry[:0]
		return
	}
	err := lp.send()
	if err != nil {
		d := lp.bf.NextDelay()
		lp.log.Error("failed to send logs:",
			slog.Int("NumRetries", lp.bf.NumRetries()),
			slog.Duration("NextDelay", d),
			slog.Any("error", err))
		if t != nil {
			t.Reset(d)
		}

	} else {
		lp.bf.Reset()
		lp.batchEntry = lp.batchEntry[:0]
	}
}
func (lp *LokiPusher) run() {
	maxWait := time.NewTicker(lp.config.BatchWait)
	defer func() {
		maxWait.Stop()
		if len(lp.batchEntry) > 0 {
			lp.sendAndResetBatch(nil)
		}
		if lp.file != nil {
			lp.file.Close()
		}
		lp.running.Done()
	}()
	for {
		select {
		case <-lp.quit:
			return
		case entry := <-lp.entries:
			lp.batchEntry = append(lp.batchEntry, *entry)
			if len(lp.batchEntry) >= lp.config.BatchMaxSize {
				lp.sendAndResetBatch(maxWait)
			}
		case <-maxWait.C:
			if len(lp.batchEntry) > 0 {
				lp.sendAndResetBatch(maxWait)
			}
		}
	}
}
func (lp *LokiPusher) send() error {
	if len(lp.batchEntry) == 0 {
		return nil
	}
	buf, err := lp.getPushReqBuffer()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, lp.config.PushURL, bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	switch lp.config.PushType {
	case PUSH_TYPE_GRPC:
		req.Header.Set("Content-Type", "application/x-protobuf")
	case PUSH_TYPE_HTTP:
		req.Header.Set("Content-Type", "application/json")
	case PUSH_TYPE_HTTP_GZIP:
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
	}
	if lp.config.UserName != "" && lp.config.Password != "" {
		req.SetBasicAuth(lp.config.UserName, lp.config.Password)
	}

	resp, err := lp.config.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	//https://grafana.com/docs/loki/latest/reference/loki-http-api/#ingest-logs
	//If the configured status code is 200, no error message will be returned.
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected response code from Loki: %d, message: %s", resp.StatusCode, body)
	}
	return nil
}
