package main

import (
	"bufio"
	"io"
	"runtime"
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

type shot struct {
	exist  bool
	number int
	bytes  []byte
}

const (
	ringBufferCapacity     = 512
	writerChannelCapacity  = 128
	scannerChannelCapacity = 128
)

type ringBuffer struct {
	index int
	pos   int
	data  [ringBufferCapacity]shot
}

func (b *ringBuffer) put(s shot) bool {
	if s.number >= b.index+ringBufferCapacity {
		return false
	}

	pos := b.pos + (s.number - b.index)
	if pos >= ringBufferCapacity {
		pos -= ringBufferCapacity
	}
	b.data[pos] = s

	return true
}

func (b *ringBuffer) get() (shot, bool) {
	if b.data[b.pos].exist {
		b.data[b.pos].exist = false
		old := b.pos
		b.pos++
		b.index++
		if b.pos >= ringBufferCapacity {
			b.pos -= ringBufferCapacity
		}

		return b.data[old], true
	}

	return shot{}, false
}

func makeWorker(w io.Writer, opts Options) (chan shot, *sync.WaitGroup) {
	rb := &ringBuffer{}

	slowBuf := make(map[int]shot)

	ch := make(chan shot, writerChannelCapacity)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		bw := bufio.NewWriterSize(w, 1024*1024)
		defer bw.Flush()

		var data shot
		ok := false

		// Buffer is enough for 10^32 lines.
		number := [40]byte{}
		numberStart := 8
		for i := range number {
			number[i] = ' '
		}

		for {
			select {
			case data, ok = <-ch:
				if !ok {
					for _, v := range slowBuf {
						rb.put(v)
					}
					for {
						s, okg := rb.get()
						if !okg {
							break
						}
						if opts.NumberLines {
							onlyNumber := strconv.AppendInt(number[numberStart:numberStart:len(number)], int64(s.number), 10)
							window := ((len(onlyNumber)-1)/numberStart + 1) * numberStart
							padding := numberStart + len(onlyNumber) - window
							bw.Write(number[padding : padding+window+1])
						}

						bw.Write(s.bytes)
					}

					return
				}
			default:
				bw.Flush()
				data, ok = <-ch
			}
			if ok {
				if !rb.put(data) {
					for _, v := range slowBuf {
						if rb.put(v) {
							delete(slowBuf, v.number)
						}
					}
					slowBuf[data.number] = data
				}

				for {
					s, okg := rb.get()
					if !okg {
						break
					}

					if opts.NumberLines {
						onlyNumber := strconv.AppendInt(number[numberStart:numberStart:len(number)], int64(s.number), 10)
						window := ((len(onlyNumber)-1)/numberStart + 1) * numberStart
						padding := numberStart + len(onlyNumber) - window
						bw.Write(number[padding : padding+window+1])
					}

					bw.Write(s.bytes)
				}
			}
		}
	}()

	return ch, &wg
}

type scanEntry struct {
	number     int
	needFormat bool
	data       []byte
}

func makeFormatter(us chan scanEntry, ds chan shot, opts Options) *sync.WaitGroup {
	eseq := logftext.EscapeSequence{NoColor: opts.NoColor}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := 0

		for se := range us {
			c++
			s := shot{true, se.number - 1, se.data}
			if se.needFormat {
				e, ok := parse(se.data)
				if !ok {
					bt := make([]byte, len(se.data)+1)
					copy(bt, se.data)
					bt[len(bt)-1] = '\n'
					s.bytes = bt
				} else {
					buf := logf.NewBufferWithCapacity(1024)
					adoptEntry(&e)
					format(buf, eseq, &e, opts.TimeFormat)
					s.bytes = buf.Bytes()
				}
			}

			ds <- s
		}
	}()

	return &wg
}

func scan(r io.Reader, w io.Writer, opts Options) (int, error) {
	scanBuf := make([]byte, opts.BufferSize)

	usCh := make(chan scanEntry, scannerChannelCapacity)

	ch, wg := makeWorker(w, opts)
	defer wg.Wait()
	defer close(ch)

	wgs := make([]*sync.WaitGroup, runtime.NumCPU())
	for i := 0; i < len(wgs); i++ {
		wgs[i] = makeFormatter(usCh, ch, opts)
	}
	defer func() {
		close(usCh)

		for i := 0; i < len(wgs); i++ {
			wgs[i].Wait()
		}
	}()

	lastLineWasTooLong := false
	for {
		scanner := bufio.NewScanner(r)
		scanner.Buffer(scanBuf, len(scanBuf))

		for scanner.Scan() {
			se := scanEntry{number: opts.StartingNumber}
			opts.StartingNumber++

			if lastLineWasTooLong {
				lastLineWasTooLong = false
				se.data = []byte("<line too long>\n")
			} else {
				se.needFormat = true
				se.data = scanner.Bytes()
			}

			usCh <- se
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
