package main

import (
	"strings"

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
