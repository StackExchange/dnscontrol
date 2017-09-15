package printer

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
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

// Printer is a simple acstraction for printing data. Can be passed to providers to give simple output capabilities.
type Printer interface {
	Debugf(fmt string, args ...interface{})
	Warnf(fmt string, args ...interface{})
}

var reader = bufio.NewReader(os.Stdin)

type ConsolePrinter struct{}

func (c ConsolePrinter) StartDomain(domain string) {
	fmt.Printf("******************** Domain: %s\n", domain)
}

func (c ConsolePrinter) PrintCorrection(i int, correction *models.Correction) {
	fmt.Printf("#%d: %s\n", i+1, correction.Msg)
}

func (c ConsolePrinter) PromptToRun() bool {
	fmt.Print("Run? (Y/n): ")
	txt, err := reader.ReadString('\n')
	run := true
	if err != nil {
		run = false
	}
	txt = strings.ToLower(strings.TrimSpace(txt))
	if txt != "y" {
		run = false
	}
	if !run {
		fmt.Println("Skipping")
	}
	return run
}

func (c ConsolePrinter) EndCorrection(err error) {
	if err != nil {
		fmt.Println("FAILURE!", err)
	} else {
		fmt.Println("SUCCESS!")
	}
}

func (c ConsolePrinter) StartDNSProvider(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	fmt.Printf("----- DNS Provider: %s...%s", provider, lbl)
}

func (c ConsolePrinter) StartRegistrar(provider string, skip bool) {
	lbl := ""
	if skip {
		lbl = " (skipping)\n"
	}
	fmt.Printf("----- Registrar: %s...%s", provider, lbl)
}

func (c ConsolePrinter) EndProvider(numCorrections int, err error) {
	if err != nil {
		fmt.Println("ERROR")
		fmt.Printf("Error getting corrections: %s\n", err)
	} else {
		plural := "s"
		if numCorrections == 1 {
			plural = ""
		}
		fmt.Printf("%d correction%s\n", numCorrections, plural)
	}
}

func (c ConsolePrinter) Debugf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (c ConsolePrinter) Warnf(format string, args ...interface{}) {
	fmt.Printf("WARNING: "+format, args...)
}
