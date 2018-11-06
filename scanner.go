package main

import (
	"bufio"
	"io"
	"os"

	"github.com/ssgreg/logf"
)

func scan(r io.Reader) {
	buf := logf.NewBuffer()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		e, ok := parse(scanner.Bytes())
		if !ok {
			os.Stdout.Write(scanner.Bytes())
			os.Stdout.Write([]byte{'\n'})

			continue
		}

		adoptEntry(&e)
		format(buf, &e)

		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}
