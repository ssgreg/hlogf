package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/ssgreg/logftext"
)

const (
	// Default read buffer size, in units of KiB (1024 bytes).
	defaultBufferSize = uint(1024 * 10)
)

func main() {
	coloredLogs := flag.String("color", "auto", `Show colored logs ("always"|"never"|"auto"). --color= is the same as --color=always.`)
	bufferSize := flag.Uint("buffer-size", defaultBufferSize, `Set the read buffer size to buffer-size, in units of KiB (1024 bytes).`)
	flag.Parse()

	signal.Ignore(os.Interrupt)

	err := scan(os.Stdin, os.Stdout, handleColorOption(*coloredLogs), handleBufferSize(*bufferSize))
	if err != nil {
		fmt.Fprintf(os.Stderr, "hlogf failed: %s", err)
		os.Exit(1)
	}
}

// handleColorOption handles 'color' option. It returns true if colored
// output should be turned off.
func handleColorOption(coloredLogs string) bool {
	force := false

	switch strings.ToLower(coloredLogs) {
	case "never":
		return true

	case "always", "":
		force = true

		fallthrough

	default:
		ok := logftext.EnableSeqTTY(os.Stdout, true)
		return !force && (!ok || logftext.CheckNoColor())
	}
}

// handleBufferSize handles 'buffer-size' option. It returns buffer size
// in bytes.
func handleBufferSize(bufferSize uint) uint {
	return bufferSize * 1024
}
