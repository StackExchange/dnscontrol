package printer

import (
	"bufio"
	"os"
)

// CLI is an abstraction around the CLI.
type CLI interface {
	Printer
	PromptToRun() bool
}

// Printer is a simple abstraction for printing data. Can be passed to providers to give simple output capabilities.
type Printer interface {
	Debugf(fmt string, args ...interface{})
	Infof(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
	Errorf(fmt string, args ...interface{})
	Printf(fmt string, args ...interface{})
	Println(lines ...string)
}

var (
	// DefaultPrinter is the default Printer, used by Debugf, Printf, and Warnf.
	DefaultPrinter = &ConsolePrinter{
		Reader: bufio.NewReader(os.Stdin),
		Writer: os.Stdout,
	}

	Verbose = false
)
