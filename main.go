package main

import (
	"os"
	"os/signal"
)

func main() {
	signal.Ignore(os.Interrupt)

	scan(os.Stdin)
}
