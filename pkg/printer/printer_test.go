package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDefaultPrinter checks that the DefaultPrinter properly controls output from the package-level
// Warnf/Printf/Debugf functions.
func TestDefaultPrinter(t *testing.T) {
	old := DefaultPrinter
	defer func() {
		DefaultPrinter = old
	}()

	output := &bytes.Buffer{}
	DefaultPrinter = &ConsolePrinter{
		Writer:  output,
		Verbose: true,
	}

	Warnf("warn\n")
	Printf("output\n")
	Debugf("debugging\n")
	assert.Equal(t, "WARNING: warn\noutput\ndebugging\n", output.String())
}

func TestVerbose(t *testing.T) {
	output := &bytes.Buffer{}
	p := ConsolePrinter{
		Writer:  output,
		Verbose: false,
	}

	// Test that verbose output is suppressed.
	p.Warnf("a dire warning!\n")
	p.Printf("output\n")
	p.Debugf("debugging\n")
	assert.Equal(t, "WARNING: a dire warning!\noutput\n", output.String())

	// Test that Verbose output can be dynamically enabled.
	p.Verbose = true
	p.Debugf("more debugging\n")
	assert.Equal(t, "WARNING: a dire warning!\noutput\nmore debugging\n", output.String())
}
