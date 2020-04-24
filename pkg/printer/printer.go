package printer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// CLI is an abstraction around the CLI.
type CLI interface {
	Printer
	StartDomain(domain string)
	StartDNSProvider(name string, skip bool)
	EndProvider(numCorrections int, err error)
	StartRegistrar(name string, skip bool)

	PrintCorrection(n int, c *models.Correction)
	EndCorrection(err error)
	PromptToRun() bool
}

// Printer is a simple abstraction for printing data. Can be passed to providers to give simple output capabilities.
type Printer interface {
	Debugf(fmt string, args ...interface{})
	Printf(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
}

// Debugf is called to print/format debug information.
func Debugf(fmt string, args ...interface{}) {
	DefaultPrinter.Debugf(fmt, args...)
}

// Printf is called to print/format information.
func Printf(fmt string, args ...interface{}) {
	DefaultPrinter.Printf(fmt, args...)
}

// Warnf is called to print/format a warning.
func Warnf(fmt string, args ...interface{}) {
	DefaultPrinter.Warnf(fmt, args...)
}

var (
	// DefaultPrinter is the default Printer, used by Debugf, Printf, and Warnf.
	DefaultPrinter = &ConsolePrinter{
		Reader:  bufio.NewReader(os.Stdin),
		Writer:  os.Stdout,
		Verbose: false,
	}
)

// ConsolePrinter is a handle for the console printer.
type ConsolePrinter struct {
	Reader *bufio.Reader
	Writer io.Writer

	Verbose bool
}

// StartDomain is called at the start of each domain.
func (c ConsolePrinter) StartDomain(domain string) {
	fmt.Fprintf(c.Writer, "******************** Domain: %s\n", domain)
}

// PrintCorrection is called to print/format each correction.
func (c ConsolePrinter) PrintCorrection(i int, correction *models.Correction) {
	fmt.Fprintf(c.Writer, "#%d: %s\n", i+1, correction.Msg)
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

// EndCorrection is called at the end of each correction.
func (c ConsolePrinter) EndCorrection(err error) {
	if err != nil {
		fmt.Fprintln(c.Writer, "FAILURE!", err)
	} else {
		fmt.Fprintln(c.Writer, "SUCCESS!")
	}
}

// StartDNSProvider is called at the start of each new provider.
func (c ConsolePrinter) StartDNSProvider(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	fmt.Fprintf(c.Writer, "----- DNS Provider: %s...%s", provider, lbl)
}

// StartRegistrar is called at the start of each new registrar.
func (c ConsolePrinter) StartRegistrar(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	fmt.Fprintf(c.Writer, "----- Registrar: %s...%s", provider, lbl)
}

// EndProvider is called at the end of each provider.
func (c ConsolePrinter) EndProvider(numCorrections int, err error) {
	if err != nil {
		fmt.Fprintln(c.Writer, "ERROR")
		fmt.Fprintf(c.Writer, "Error getting corrections: %s\n", err)
	} else {
		plural := "s"
		if numCorrections == 1 {
			plural = ""
		}
		fmt.Fprintf(c.Writer, "%d correction%s\n", numCorrections, plural)
	}
}

// Debugf is called to print/format debug information.
func (c ConsolePrinter) Debugf(format string, args ...interface{}) {
	if c.Verbose {
		fmt.Fprintf(c.Writer, format, args...)
	}
}

// Printf is called to print/format information.
func (c ConsolePrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, format, args...)
}

// Warnf is called to print/format a warning.
func (c ConsolePrinter) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(c.Writer, "WARNING: "+format, args...)
}
