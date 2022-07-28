package printer

import (
	"bytes"
	"fmt"
	"sync"
)

var mutex sync.RWMutex

// GetBufferedPrinter return a BufferedPrinter for a domain
func GetBufferedPrinter(domain string) Printer {
	mutex.RLock()
	if writer, ok := DomainWriter[domain]; ok {
		mutex.RUnlock()
		return writer
	}

	mutex.RUnlock()
	writer := &BufferedPrinter{
		Buffer: bytes.Buffer{},
	}
	mutex.Lock()
	DomainWriter[domain] = writer
	mutex.Unlock()
	return writer
}

// BufferedPrinter buffers log messages until flushed
type BufferedPrinter struct {
	bytes.Buffer
}

// Debugf is called to print/format debug information.
func (b *BufferedPrinter) Debugf(format string, args ...interface{}) {
	if Verbose {
		fmt.Fprintf(b, format, args...)
	}
}

// Infof is called to print/format debug information.
func (b *BufferedPrinter) Infof(format string, args ...interface{}) {
	b.Printf(format, args...)
}

// Printf is called to print/format information.
func (b *BufferedPrinter) Printf(format string, args ...interface{}) {
	fmt.Fprintf(b, format, args...)
}

// Println is called to print/format information.
func (b *BufferedPrinter) Println(lines ...string) {
	fmt.Fprintln(b, lines)
}

// Warnf is called to print/format a warning.
func (b *BufferedPrinter) Warnf(format string, args ...interface{}) {
	fmt.Fprintf(b, "WARNING: "+format, args...)
}

// Errorf is called to print/format an error.
func (b *BufferedPrinter) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(b, "ERROR: "+format, args...)
}
