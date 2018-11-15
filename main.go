package main

import (
	"os"
	"os/signal"

	"github.com/ssgreg/logftext"
)

func main() {
	signal.Ignore(os.Interrupt)
	ok := logftext.EnableSeqTTY(os.Stdout, true)
	noColor := !ok || logftext.CheckNoColor()

	scan(os.Stdin, noColor)
}
