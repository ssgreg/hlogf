package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ssgreg/logftext"
)

const (
	// Default read buffer size, in units of KiB (1024 bytes).
	defaultBufferSize = uint(1024 * 10)

	// Default time format.
	defaultTimeFormat = time.StampMilli
)

var (
	// Version is specified automatically by goreleaser.
	version = "dev"
)

func main() {
	cmd := newRootCommand()
	cmd.SetHelpTemplate(helpTemplate)

	err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "hlogf failed: %s\n", err.Error())
		os.Exit(1)
	}
}

type rootOptions struct {
	coloredLogs string
	bufferSize  uint
	numberLines bool
	timeFormat  string
	files       []string
}

func newRootCommand() *cobra.Command {
	var opts rootOptions

	cmd := &cobra.Command{
		Use:           "hlogf [OPTIONS] [file ...]",
		Short:         description,
		Example:       example,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ArbitraryArgs,
		Version:       version,
	}

	// TODO: override default values with environment variables.
	// TODO: query by field set (dont forget about excludes, see grep for details).
	// TODO: custom field name mapping.
	// TODO: customize skipping systemd fields.

	flags := cmd.PersistentFlags()
	flags.StringVarP(&opts.timeFormat, "time-format", "t", defaultTimeFormat, `Set format for 'time' field using golang time format. e.g. "2006-01-02T15:04:05.999999999Z07:00"`)
	flags.StringVar(&opts.coloredLogs, "color", "auto", `Show colored logs ("always"|"never"|"auto"). --color= is the same as --color=always.`)
	flags.UintVar(&opts.bufferSize, "buffer-size", defaultBufferSize, `Set the read buffer size to buffer-size, in units of KiB (1024 bytes).`)
	flags.BoolVarP(&opts.numberLines, "number", "n", false, `Number the output lines, starting at 1.`)
	flags.BoolP("version", "v", false, "Print version information and exit.")
	flags.BoolP("help", "h", false, "Print this help and exit.")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts.files = args

		return runRoot(opts)
	}

	return cmd
}

func runRoot(opts rootOptions) error {
	signal.Ignore(os.Interrupt)

	out := os.Stdout
	scanOpts := Options{
		NoColor:        handleColorOption(opts.coloredLogs),
		BufferSize:     handleBufferSize(opts.bufferSize),
		NumberLines:    opts.numberLines,
		StartingNumber: 1,
		TimeFormat:     opts.timeFormat,
	}

	handleReader := func(r io.Reader) error {
		var err error
		scanOpts.StartingNumber, err = scan(r, out, scanOpts)
		if err != nil {
			return err
		}

		return nil
	}

	handleFile := func(name string) error {
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		defer func() {
			_ = f.Close()
		}()

		return handleReader(f)
	}

	if len(opts.files) == 0 {
		// No files were specified. Read stdin.
		return handleReader(os.Stdin)
	}

	// Scan all specified files.
	for _, file := range opts.files {
		var handle func() error
		switch file {
		case "-":
			handle = func() error {
				return handleReader(os.Stdin)
			}
		default:
			handle = func() error {
				return handleFile(file)
			}
		}

		err := handle()
		if err != nil {
			return err
		}
	}

	return nil
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

const (
	description = `
Makes json logs possible to read by humans. Supports systemd journal.

The hlogf reads and parses files sequentally, writing the colored logs to the standard output.
The 'file' operands are processed in command-line order. If 'file' is a single dash '-' or
absent, hlogf reads from the standard input.`

	example = `
  The command:

  	hlogf file1
  	
  will parse the content of file1 and print parsed result to the standard output.
  
  The command:
  
  	hlogf file1 file2 > file3

  will sequentially parse the content of file1 and file2 and print parsed result to the file3,
  truncating file3 if it already exists.`

	helpTemplate = `Usage: {{.Use}}
{{.Short}}

Examples:{{.Example}}

Options:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`
)
