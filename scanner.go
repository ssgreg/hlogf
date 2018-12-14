package main

import (
	"bufio"
	"fmt"
	"io"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"
)

func scan(r io.Reader, w io.Writer, noColor bool, bufferSize uint) error {
	buf := logf.NewBuffer()
	scanBuf := make([]byte, bufferSize)
	eseq := logftext.EscapeSequence{NoColor: noColor}

	lastLineWasTooLong := false
	for {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(scanBuf, len(scanBuf))

		for scanner.Scan() {
			if lastLineWasTooLong {
				//

				lastLineWasTooLong = false
				fmt.Println("<line too long>")

				continue
			}

			e, ok := parse(scanner.Bytes())
			if !ok {
				w.Write(scanner.Bytes())
				w.Write([]byte{'\n'})

				continue
			}

			adoptEntry(&e)
			format(buf, eseq, &e)

			w.Write(buf.Bytes())
			buf.Reset()
		}

		switch scanner.Err() {
		case nil:
			return nil

		case bufio.ErrTooLong:
			// Data does not match to the buffer. As scanner drops the read
			// data there's nothing we can do about it except setting the flag
			// to drop the final (next) part of data.
			lastLineWasTooLong = true

		default:
			return scanner.Err()
		}
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
