package main

import (
	"bufio"
	"io"
	"os"

	"github.com/ssgreg/logf"
)

func scan(r io.Reader) {
	buf := logf.NewBuffer()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		e, ok := parse(scanner.Bytes())
		if !ok {
			os.Stdout.Write(scanner.Bytes())
			os.Stdout.Write([]byte{'\n'})

			continue
		}

		adoptEntry(&e)
		format(buf, &e)

		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

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
