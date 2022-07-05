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

// Debugf is called to print/format debug information.
func Debugf(fmt string, args ...interface{}) {
	DefaultPrinter.Debugf(fmt, args...)
}

// Infof is called to print/format information.
func Infof(fmt string, args ...interface{}) {
	DefaultPrinter.Infof(fmt, args...)
}

// Printf is called to print/format information.
func Printf(fmt string, args ...interface{}) {
	DefaultPrinter.Printf(fmt, args...)
}

// Println is called to print/format information.
func Println(lines ...string) {
	DefaultPrinter.Println(lines...)
}

// Warnf is called to print/format a warning.
func Warnf(fmt string, args ...interface{}) {
	DefaultPrinter.Warnf(fmt, args...)
}

// Errorf is called to print/format an error.
func Errorf(fmt string, args ...interface{}) {
	DefaultPrinter.Errorf(fmt, args...)
}

var (
	// DefaultPrinter is the default Printer, used by Debugf, Printf, and Warnf.
	DefaultPrinter = &ConsolePrinter{
		Reader:  bufio.NewReader(os.Stdin),
		Writer:  os.Stdout,
		Verbose: false,
	}

	DomainWriter = map[string]*BufferedPrinter{}
)
