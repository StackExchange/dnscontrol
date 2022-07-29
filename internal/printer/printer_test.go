package printer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerbose(t *testing.T) {
	Verbose = false
	output := &bytes.Buffer{}
	p := ConsolePrinter{
		Writer: output,
	}

	// Test that verbose output is suppressed.
	p.Warnf("a dire warning!\n")
	p.Printf("output\n")
	p.Debugf("debugging\n")
	assert.Equal(t, "WARNING: a dire warning!\noutput\n", output.String())

	// Test that Verbose output can be dynamically enabled.
	Verbose = true
	p.Debugf("more debugging\n")
	assert.Equal(t, "WARNING: a dire warning!\noutput\nmore debugging\n", output.String())
}
