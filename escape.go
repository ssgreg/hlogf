package main

import (
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/ssgreg/logf"
)

func unescapeString(buf *logf.Buffer, data []byte) {
	p := 0
	i := 0
	for i < len(data) {
		switch data[i] {
		case '\\':
			buf.AppendBytes(data[p:i])
			i += handleEscapeSequence(buf, data[i:])
			p = i
		default:
			i++
		}
	}

	buf.AppendBytes(data[p:i])
}

// handleEscapeSequence handles a single escape sequence.
func handleEscapeSequence(buf *logf.Buffer, data []byte) int {
	if len(data) < 2 {
		return len(data)
	}

	c := data[1]
	switch c {
	case '"', '/', '\\':
		buf.AppendByte(c)
	case 'b':
		buf.AppendByte('\b')
	case 'f':
		buf.AppendByte('\f')
	case 'n':
		buf.AppendByte('\n')
	case 'r':
		buf.AppendByte('\r')
	case 't':
		buf.AppendByte('\t')
	case 'u':
		length := 6
		if len(data) < length {
			return len(data)
		}
		r1 := decodeU4(data)
		if r1 < 0 {
			return length
		}

		if utf16.IsSurrogate(r1) {
			if len(data[length:]) < length {
				return len(data)
			}
			r2 := decodeU4(data[length:])
			if d := utf16.DecodeRune(r1, r2); d != unicode.ReplacementChar {
				length += 6
				r1 = d
			} else {
				r1 = unicode.ReplacementChar
			}
		}

		var d [4]byte
		s := utf8.EncodeRune(d[:], r1)
		buf.AppendBytes(d[:s])

		return length
	}

	return 2
}

// decodeU4 decodes \uXXXX bytes string to a rune.
func decodeU4(s []byte) rune {
	if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
		return -1
	}

	var r rune
	for i := 2; i < len(s) && i < 6; i++ {
		var v byte
		c := s[i]
		switch {
		case '0' <= c && c <= '9':
			v = c - '0'
		case 'a' <= c && c <= 'z':
			v = c - 'a' + 10
		case 'A' <= c && c <= 'Z':
			v = c - 'A' + 10
		default:
			return -1
		}

		r <<= 4
		r |= rune(v)
	}

	return r
}
