package main

import (
	"flag"
	"os"
	"os/signal"
	"strings"

	"github.com/ssgreg/logftext"
)

func main() {
	coloredLogs := flag.String("color", "auto", `Show colored logs ("always"|"never"|"auto"). --color= is the same as --color=always.`)
	flag.Parse()

	signal.Ignore(os.Interrupt)
	scan(os.Stdin, handleColorOption(*coloredLogs))
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
