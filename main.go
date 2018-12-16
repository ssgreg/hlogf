package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ssgreg/logftext"
)

const (
	// Default read buffer size, in units of KiB (1024 bytes).
	defaultBufferSize = uint(1024 * 10)
)

func main() {
	cmd := newCommand()
	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "hlogf failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	coloredLogs := ""
	bufferSize := uint(0)

	cmd := &cobra.Command{
		Use:           "hlogf",
		Short:         "Makes json logs possible to read by humans. Supports systemd journal.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
	}

	flags := cmd.PersistentFlags()
	flags.StringVar(&coloredLogs, "color", "auto", `Show colored logs ("always"|"never"|"auto"). --color= is the same as --color=always.`)
	flags.UintVar(&bufferSize, "buffer-size", defaultBufferSize, `Set the read buffer size to buffer-size, in units of KiB (1024 bytes).`)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		signal.Ignore(os.Interrupt)

		out := os.Stdout
		handledNoColor := handleColorOption(coloredLogs)
		handledBufferSize := handleBufferSize(bufferSize)

		if len(args) == 0 {
			// No files were specified. Read stdin.
			return scan(os.Stdin, out, handledNoColor, handledBufferSize)
		}

		// Scan all specified files.
		for _, file := range args {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer func() {
				_ = f.Close()
			}()

			err = scan(f, out, handledNoColor, handledBufferSize)
			if err != nil {
				return err
			}
		}

		return nil
	}

	return cmd
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
