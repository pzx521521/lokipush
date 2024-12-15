package lokipush

import (
	"bytes"
	"strconv"
)

// use by https://github.com/grafana/loki/pkg/pattern/aggregation/push.go#L193
// use by https://github.com/grafana/loki/pkg/pattern/aggregation/push.go#L22
// copy from github.com/prometheus/prometheus/model/labels/labels_common.go
// add a costume func HttpString
type Label struct {
	Name, Value string
}
type Labels []Label

func (ls Labels) String() string {
	var bytea [1024]byte // On stack to avoid memory allocation while building the output.
	b := bytes.NewBuffer(bytea[:0])

	b.WriteByte('{')
	i := 0
	ls.Range(func(l Label) {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.WriteString(l.Name)
		b.WriteByte('=')
		b.Write(strconv.AppendQuote(b.AvailableBuffer(), l.Value))
		i++
	})
	b.WriteByte('}')
	return b.String()
}

func (ls Labels) HttpString() string {
	var bytea [1024]byte // On stack to avoid memory allocation while building the output.
	b := bytes.NewBuffer(bytea[:0])
	b.WriteByte('{')
	i := 0
	ls.Range(func(l Label) {
		if i > 0 {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		b.Write(strconv.AppendQuote(b.AvailableBuffer(), l.Name))
		b.WriteByte(':')
		b.Write(strconv.AppendQuote(b.AvailableBuffer(), l.Value))
		i++
	})
	b.WriteByte('}')
	return b.String()
}
func (ls Labels) Range(f func(l Label)) {
	for _, l := range ls {
		f(l)
	}
}
