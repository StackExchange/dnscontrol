package printer

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ConsolePrinter is a handle for the console printer.
type ConsolePrinter struct {
	Reader *bufio.Reader
	Writer io.Writer
}

// PromptToRun prompts the user to see if they want to execute a correction.
func (c ConsolePrinter) PromptToRun() bool {
	fmt.Fprint(c.Writer, "Run? (Y/n): ")
	txt, err := c.Reader.ReadString('\n')
	run := true
	if err != nil {
		run = false
	}
	txt = strings.ToLower(strings.TrimSpace(txt))
	if txt != "y" {
		run = false
	}
	if !run {
		fmt.Fprintln(c.Writer, "Skipping")
	}
	return run
}

// Debugf is called to print/format debug information.
func (c ConsolePrinter) Debugf(format string, args ...interface{}) {
	if Verbose {
		fmt.Fprintf(c.Writer, format, args...)
	}
}

// Infof is called to print/format debug information.
func (c ConsolePrinter) Infof(format string, args ...interface{}) {
	c.Printf(format, args...)
}

// Printf is called to print/format information.
func (c ConsolePrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, format, args...)
}

// Println is called to print/format information.
func (c ConsolePrinter) Println(lines ...string) {
	fmt.Fprintln(c.Writer, lines)
}

// Warnf is called to print/format a warning.
func (c ConsolePrinter) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, "WARNING: "+format, args...)
}

// Errorf is called to print/format an error.
func (c ConsolePrinter) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, "ERROR: "+format, args...)
}
