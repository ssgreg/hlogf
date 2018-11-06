package main

import (
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/ssgreg/logf"
)

type Field struct {
	Key   []byte
	Value []byte
}

type Entry struct {
	Time              []byte
	SourceTimestamp   []byte
	RealtimeTimestamp []byte

	Level    []byte
	Msg      []byte
	Name     []byte
	Caller   []byte
	Priority []byte
	Fields   []Field
}

func parse(data []byte) (Entry, bool) {
	var t Entry

	idx := 0
	// TODO: check for curly brackets
	data = data[1 : len(data)-1]

	for idx < len(data) {
		v, length, ok := fetchKey(data[idx:])
		if !ok {
			return t, false
		}
		idx += length + 1

		v1, length, ok := fetchValue(data[idx:])
		if !ok {
			return t, false
		}
		idx += length + 1

		switch string(v) {
		case "level", "LEVEL":
			if len(t.Level) == 0 {
				t.Level = v1
			} else {
				if v[0] != '_' {
					t.Fields = append(t.Fields, Field{v, v1})
				}
			}
		case "ts", "TS":
			t.Time = v1
		case "_SOURCE_REALTIME_TIMESTAMP":
			t.RealtimeTimestamp = v1
		case "__REALTIME_TIMESTAMP":
			t.SourceTimestamp = v1
		case "msg", "MESSAGE":
			t.Msg = v1
		case "logger", "LOGGER":
			t.Name = v1
		case "caller", "CALLER":
			t.Caller = v1
		case "PRIORITY":
			t.Priority = v1
		case "SYSLOG_FACILITY", "SYSLOG_IDENTIFIER":
		default:
			if v[0] != '_' {
				t.Fields = append(t.Fields, Field{v, v1})
			}
		}
	}

	return t, true
}

func main() {
	signal.Ignore(os.Interrupt)

	scan(os.Stdin)
}

func appendTime(buf *logf.Buffer, ts []byte) {
	logf.AtEscapeSequence(buf, logf.EscBrightBlack, func() {
		t, ok := encodeTime(ts)
		if !ok {
			buf.AppendString("<<    no time    >>")
		}
		buf.Data = t.AppendFormat(buf.Data, time.StampMilli)
	})
}

func appendLevel(buf *logf.Buffer, lvl []byte) {
	if len(lvl) < 2 {
		lvl = []byte(`"unknown"`)
	}

	buf.AppendByte('|')

	switch strings.ToLower(string(lvl[1 : len(lvl)-1])) {
	case "debug":
		logf.AtEscapeSequence(buf, logf.EscMagenta, func() {
			buf.AppendString("DEBU")
		})
	case "info", "information":
		logf.AtEscapeSequence(buf, logf.EscCyan, func() {
			buf.AppendString("INFO")
		})
	case "warn", "warning":
		logf.AtEscapeSequence2(buf, logf.EscBrightYellow, logf.EscReverse, func() {
			buf.AppendString("WARN")
		})
	case "err", "error", "fatal", "panic":
		logf.AtEscapeSequence2(buf, logf.EscBrightRed, logf.EscReverse, func() {
			buf.AppendString("ERRO")
		})
	default:
		logf.AtEscapeSequence(buf, logf.EscBrightRed, func() {
			buf.AppendString("UNKN")
		})
	}

	buf.AppendByte('|')
}
