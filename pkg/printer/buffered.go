package printer

import (
	"bytes"
	"fmt"
	"sync"
)

type BufferedPrinter struct {
	Writer *bytes.Buffer
}

var mutex sync.RWMutex

func GetBufferedPrinter(domain string) Printer {
	mutex.RLock()
	if writer, ok := DomainWriter[domain]; ok {
		mutex.RUnlock()
		return writer
	}

	mutex.RUnlock()
	writer := &BufferedPrinter{
		Writer: &bytes.Buffer{},
	}
	mutex.Lock()
	DomainWriter[domain] = writer
	mutex.Unlock()
	return writer
}

// Debugf is called to print/format debug information.
func (b BufferedPrinter) Debugf(format string, args ...interface{}) {
	if Verbose {
		fmt.Fprintf(b.Writer, format, args...)
	}
}

// Infof is called to print/format debug information.
func (b BufferedPrinter) Infof(format string, args ...interface{}) {
	b.Printf(format, args...)
}

// Printf is called to print/format information.
func (b BufferedPrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(b.Writer, format, args...)
}

// Println is called to print/format information.
func (b BufferedPrinter) Println(lines ...string) {
	fmt.Fprintln(b.Writer, lines)
}

// Warnf is called to print/format a warning.
func (b BufferedPrinter) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(b.Writer, "WARNING: "+format, args...)
}

// Errorf is called to print/format an error.
func (b BufferedPrinter) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(b.Writer, "ERROR: "+format, args...)
}
