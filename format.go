package main

import (
	"strings"
	"time"

	"github.com/ssgreg/logf"
)

func format(buf *logf.Buffer, e *Entry) {
	// Time.
	appendTime(buf, e.Time)

	// Level.
	buf.AppendByte(' ')
	appendLevel(buf, e.Level)

	// Logger name.
	if len(e.Name) != 0 {
		buf.AppendByte(' ')
		logf.AtEscapeSequence(buf, logf.EscBrightBlack, func() {
			buf.AppendBytes(e.Name[1 : len(e.Name)-1])
			buf.AppendByte(':')
		})
	}

	// Message.
	buf.AppendByte(' ')
	logf.AtEscapeSequence(buf, logf.EscBrightWhite, func() {
		unescapeString(buf, e.Msg[1:len(e.Msg)-1])
	})

	// Fields.
	for _, f := range e.Fields {
		buf.AppendByte(' ')
		logf.AtEscapeSequence(buf, logf.EscGreen, func() {
			buf.AppendString(strings.ToLower(string(f.Key)))
		})
		logf.AtEscapeSequence(buf, logf.EscBrightBlack, func() {
			buf.AppendByte('=')
		})
		buf.AppendBytes(f.Value)
	}

	// Caller.
	if len(e.Caller) != 0 {
		buf.AppendByte(' ')
		logf.AtEscapeSequence(buf, logf.EscBrightBlack, func() {
			buf.AppendByte('@')
			buf.AppendBytes(e.Caller)
		})
	}

	buf.AppendByte('\n')
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
