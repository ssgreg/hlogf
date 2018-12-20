package main

import (
	"strings"

	"github.com/ssgreg/logf"
	"github.com/ssgreg/logftext"
)

func format(buf *logf.Buffer, eseq logftext.EscapeSequence, e *Entry, timeFormat string) {
	// Time.
	appendTime(buf, eseq, e.Time, timeFormat)

	// Level.
	buf.AppendByte(' ')
	appendLevel(buf, eseq, e.Level)

	// Logger name.
	if len(e.Name) != 0 {
		buf.AppendByte(' ')
		eseq.At(buf, logftext.EscBrightBlack, func() {
			buf.AppendBytes(e.Name[1 : len(e.Name)-1])
			buf.AppendByte(':')
		})
	}

	// Message.
	buf.AppendByte(' ')
	eseq.At(buf, logftext.EscBrightWhite, func() {
		unescapeString(buf, e.Msg[1:len(e.Msg)-1])
	})

	// Fields.
	for _, f := range e.Fields {
		buf.AppendByte(' ')
		eseq.At(buf, logftext.EscGreen, func() {
			buf.AppendString(strings.ToLower(string(f.Key)))
		})
		eseq.At(buf, logftext.EscBrightBlack, func() {
			buf.AppendByte('=')
		})
		unescapeString(buf, f.Value)
	}

	// Caller.
	if len(e.Caller) != 0 {
		buf.AppendByte(' ')
		eseq.At(buf, logftext.EscBrightBlack, func() {
			buf.AppendByte('@')
			buf.AppendBytes(e.Caller)
		})
	}

	buf.AppendByte('\n')
}

const (
	badTime       = "bad time"
	leftBrackets  = "<<"
	rightBrackets = ">>"
)

var (
	templateBadTime = ""
)

func formatTemplateBadTime(timeFormat string) {
	totalLen := len(badTime) + len(leftBrackets) + len(rightBrackets)
	if len(timeFormat) >= totalLen {
		spaceLen := (len(timeFormat) - totalLen) / 2
		templateBadTime = leftBrackets
		for i := 0; i < spaceLen; i++ {
			templateBadTime += " "
		}
		templateBadTime += badTime
		for i := 0; i < len(timeFormat)-totalLen-spaceLen; i++ {
			templateBadTime += " "
		}
		templateBadTime += rightBrackets
	} else {
		for i := 0; i < len(timeFormat); i++ {
			templateBadTime += "-"
		}
	}
}

func appendTime(buf *logf.Buffer, eseq logftext.EscapeSequence, ts []byte, timeFormat string) {
	eseq.At(buf, logftext.EscBrightBlack, func() {
		t, ok := encodeTime(ts)
		if !ok {
			if templateBadTime == "" {
				formatTemplateBadTime(timeFormat)
			}
			buf.AppendString(templateBadTime)

			return
		}
		buf.Data = t.AppendFormat(buf.Data, timeFormat)
	})
}

func appendLevel(buf *logf.Buffer, eseq logftext.EscapeSequence, lvl []byte) {
	if len(lvl) < 2 {
		lvl = []byte(`"unknown"`)
	}

	buf.AppendByte('|')

	switch strings.ToLower(string(lvl[1 : len(lvl)-1])) {
	case "debug":
		eseq.At(buf, logftext.EscMagenta, func() {
			buf.AppendString("DEBU")
		})
	case "info", "information":
		eseq.At(buf, logftext.EscCyan, func() {
			buf.AppendString("INFO")
		})
	case "warn", "warning":
		eseq.At2(buf, logftext.EscBrightYellow, logftext.EscReverse, func() {
			buf.AppendString("WARN")
		})
	case "err", "error", "fatal", "panic":
		eseq.At2(buf, logftext.EscBrightRed, logftext.EscReverse, func() {
			buf.AppendString("ERRO")
		})
	default:
		eseq.At(buf, logftext.EscBrightRed, func() {
			buf.AppendString("UNKN")
		})
	}

	buf.AppendByte('|')
}
