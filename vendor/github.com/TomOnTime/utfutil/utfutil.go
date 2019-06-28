// Package utfutil provides methods that make it easy to read data in an UTF-encoding agnostic.
package utfutil

// These functions autodetect UTF BOM and return UTF-8. If no
// BOM is found, a hint is provided as to which encoding to assume.
// You can use them as replacements for os.Open() and ioutil.ReadFile()
// when the encoding of the file is unknown.

// utfutil.OpenFile() is a replacement for os.Open().
// utfutil.ReadFile() is a replacement for ioutil.ReadFile().
// utfutil.NewScanner() takes a filename and returns a Scanner.
// utfutil.NewReader() rewraps an existing scanner to make it UTF-encoding agnostic.
// utfutil.BytesReader() takes a []byte and decodes it to UTF-8.

// When there is no BOM, it is impossible to guess correctly 100%
// of the time.  Therefore, the functions take a 2nd parameter of type
// "EncodingHint" where you specify the default encoding for BOM-less
// data.

// In the future we'd like to have a hint called AUTO that uses
// uchatdet (or a Go rewrite) to guess.

// Inspiration: I wrote this after spending half a day trying
// to figure out how to use unicode.BOMOverride.
// Hopefully this will save other golang newbies from the same.
// (golang.org/x/text/encoding/unicode)

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// EncodingHint indicates the file's encoding if there is no BOM.
type EncodingHint int

const (
	// UTF8 indicates the specified encoding.
	UTF8 EncodingHint = iota
	// UTF16LE indicates the specified encoding.
	UTF16LE
	// UTF16BE indicates the specified encoding.
	UTF16BE
	// WINDOWS indicates that the file came from a MS-Windows system
	WINDOWS = UTF16LE
	// POSIX indicates that the file came from Unix or Unix-like systems
	POSIX = UTF8
	// HTML5 indicates that the file came from the web
	HTML5 = UTF8
)

// UTFReadCloser describes the utfutil ReadCloser structure.
type UTFReadCloser interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// ReadCloser is a readcloser for the UTFUtil package.
type readCloser struct {
	file   *os.File
	reader io.Reader
}

// Read implements the standard Reader interface.
func (u readCloser) Read(p []byte) (n int, err error) {
	return u.reader.Read(p)
}

// Close implements the standard Closer interface.
func (u readCloser) Close() error {
	if u.file != nil {
		return u.file.Close()
	}
	return nil
}

// UTFScanCloser describes a new utfutil ScanCloser structure.
// It's similar to ReadCloser, but with a scanner instead of a reader.
type UTFScanCloser interface {
	Buffer(buf []byte, max int)
	Bytes() []byte
	Err() error
	Scan() bool
	Split(split bufio.SplitFunc)
	Text() string
	Close() error
}

type scanCloser struct {
	file    UTFReadCloser
	scanner *bufio.Scanner
}

// Buffer will run the Buffer function on the underlying bufio.Scanner.
func (sc scanCloser) Buffer(buf []byte, max int) {
	sc.scanner.Buffer(buf, max)
}

// Bytes will run the Bytes function on the underlying bufio.Scanner.
func (sc scanCloser) Bytes() []byte {
	return sc.scanner.Bytes()
}

// Err will run the Err function on the underlying bufio.Scanner.
func (sc scanCloser) Err() error {
	return sc.scanner.Err()
}

// Scan will run the Scan function on the underlying bufio.Scanner.
func (sc scanCloser) Scan() bool {
	return sc.scanner.Scan()
}

// Split will run the Split function on the underlying bufio.Scanner.
func (sc scanCloser) Split(split bufio.SplitFunc) {
	sc.scanner.Split(split)
}

// Text will return the text from the underlying bufio.Scanner.
func (sc scanCloser) Text() string {
	return sc.scanner.Text()
}

// Close will close the underlying file handle.
func (sc scanCloser) Close() error {
	return sc.file.Close()
}

// About utfutil.HTML5:
// This technique is recommended by the W3C for use in HTML 5:
// "For compatibility with deployed content, the byte order
// mark (also known as BOM) is considered more authoritative
// than anything else." http://www.w3.org/TR/encoding/#specification-hooks

// OpenFile is the equivalent of os.Open().
func OpenFile(name string, d EncodingHint) (UTFReadCloser, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	rc := readCloser{file: f}
	return NewReader(rc, d), nil
}

// ReadFile is the equivalent of ioutil.ReadFile()
func ReadFile(name string, d EncodingHint) ([]byte, error) {
	file, err := OpenFile(name, d)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// NewScanner is a convenience function that takes a filename and returns a scanner.
func NewScanner(name string, d EncodingHint) (UTFScanCloser, error) {
	f, err := OpenFile(name, d)
	if err != nil {
		return nil, err
	}

	return scanCloser{
		scanner: bufio.NewScanner(f),
		file:    f,
	}, nil
}

// NewReader wraps a Reader to decode Unicode to UTF-8 as it reads.
func NewReader(r io.Reader, d EncodingHint) UTFReadCloser {
	var decoder *encoding.Decoder
	switch d {
	case UTF8:
		// Make a transformer that assumes UTF-8 but abides by the BOM.
		decoder = unicode.UTF8.NewDecoder()
	case UTF16LE:
		// Make an tranformer that decodes MS-Windows (16LE) UTF files:
		winutf := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		// Make a transformer that is like winutf, but abides by BOM if found:
		decoder = winutf.NewDecoder()
	case UTF16BE:
		// Make an tranformer that decodes UTF-16BE files:
		utf16be := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
		// Make a transformer that is like utf16be, but abides by BOM if found:
		decoder = utf16be.NewDecoder()
	}

	// Make a Reader that uses utf16bom:
	if rc, ok := r.(readCloser); ok {
		rc.reader = transform.NewReader(rc.file, unicode.BOMOverride(decoder))
		return rc
	}

	return readCloser{
		reader: transform.NewReader(r, unicode.BOMOverride(decoder)),
	}
}

// BytesReader is a convenience function that takes a []byte and decodes them to UTF-8.
func BytesReader(b []byte, d EncodingHint) io.Reader {
	return NewReader(bytes.NewReader(b), d)
}
