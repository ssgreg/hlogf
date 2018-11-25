package main

import (
	"bufio"
	"io"
	"os"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"
)

func scan(r io.Reader, noColor bool) {
	buf := logf.NewBuffer()
	eseq := logftext.EscapeSequence{NoColor: noColor}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		e, ok := parse(scanner.Bytes())
		if !ok {
			os.Stdout.Write(scanner.Bytes())
			os.Stdout.Write([]byte{'\n'})

			continue
		}

		adoptEntry(&e)
		format(buf, eseq, &e)

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
	if len(data) < 2 {
		return t, false
	}
	if data[0] != '{' || data[len(data)-1] != '}' {
		return t, false
	}
	data = data[1 : len(data)-1]

	for idx := 0; idx < len(data); {
		key, length, ok := fetchKey(data[idx:])
		if !ok {
			return t, false
		}
		idx += length + 1

		val, length, ok := fetchValue(data[idx:])
		if !ok {
			return t, false
		}
		idx += length + 1

		switch string(key) {
		case "level", "LEVEL":
			if len(t.Level) == 0 {
				t.Level = val
			} else {
				if key[0] != '_' {
					t.Fields = append(t.Fields, Field{key, val})
				}
			}
		case "ts", "TS", "time", "TIME":
			t.Time = val
		case "_SOURCE_REALTIME_TIMESTAMP":
			t.RealtimeTimestamp = val
		case "__REALTIME_TIMESTAMP":
			t.SourceTimestamp = val
		case "msg", "MESSAGE":
			t.Msg = val
		case "logger", "LOGGER":
			t.Name = val
		case "caller", "CALLER":
			t.Caller = val
		case "PRIORITY":
			t.Priority = val
		case "SYSLOG_FACILITY", "SYSLOG_IDENTIFIER":
		default:
			if key[0] != '_' {
				t.Fields = append(t.Fields, Field{key, val})
			}
		}
	}

	return t, true
}
