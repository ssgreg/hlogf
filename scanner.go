package main

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"sync"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"
)

// Options holds scan options.
type Options struct {
	NoColor        bool
	BufferSize     uint
	NumberLines    bool
	StartingNumber int
	TimeFormat     string
}

func scan(r io.Reader, w io.Writer, opts Options) (int, error) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	ch := make(chan []byte, 1024)
	defer close(ch)

	wg.Add(1)
	go func() {
		defer wg.Done()

		bw := bufio.NewWriterSize(w, 1024*1024)
		defer bw.Flush()

		dirty := false
		for {
			if !dirty {
				data, ok := <-ch
				if !ok {
					return
				}
				bw.Write(data)
				dirty = true
			}
			select {
			case data, ok := <-ch:
				if !ok {
					return
				}
				bw.Write(data)
			default:
				bw.Flush()
				dirty = false
			}
		}
	}()

	scanBuf := make([]byte, opts.BufferSize)
	eseq := logftext.EscapeSequence{NoColor: opts.NoColor}

	// Buffer is enough for 10^32 lines.
	number := [40]byte{}
	numberStart := 8
	for i := range number {
		number[i] = ' '
	}

	lastLineWasTooLong := false
	for {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(scanBuf, len(scanBuf))

		for scanner.Scan() {
			if opts.NumberLines {
				onlyNumber := strconv.AppendInt(number[numberStart:numberStart:len(number)], int64(opts.StartingNumber), 10)
				window := ((len(onlyNumber)-1)/numberStart + 1) * numberStart
				padding := numberStart + len(onlyNumber) - window
				ch <- bytes.Repeat(number[padding:padding+window+1], 1)
			}
			opts.StartingNumber++

			if lastLineWasTooLong {
				lastLineWasTooLong = false
				ch <- []byte("<line too long>\n")
			} else {
				e, ok := parse(scanner.Bytes())
				if !ok {
					ch <- append(bytes.Repeat(scanner.Bytes(), 1), '\n')
				} else {
					adoptEntry(&e)
					buf := logf.NewBufferWithCapacity(logf.PageSize * 2)
					format(buf, eseq, &e, opts.TimeFormat)
					ch <- buf.Bytes()
				}
			}
		}

		switch scanner.Err() {
		case nil:
			return opts.StartingNumber, nil

		case bufio.ErrTooLong:
			// Data does not match to the buffer. As scanner drops the read
			// data there's nothing we can do about it except setting the flag
			// to drop the final (next) part of data.
			lastLineWasTooLong = true

		default:
			return opts.StartingNumber, scanner.Err()
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
