package fuzz

import (
	"strings"
	"testing"

	"github.com/DNSControl/dnscontrol/v5/pkg/txtutil"
)

// FuzzTXTQuoted fuzzes the RFC1035-quoted TXT record handling in
// pkg/txtutil: parsing quoted strings, splitting them into 255-octet
// chunks, and re-encoding them.  These functions process user input
// (TXT record contents in dnsconfig.js) and must never panic.
func FuzzTXTQuoted(f *testing.F) {
	f.Add("")
	f.Add("foo")
	f.Add(`"foo"`)
	f.Add(`"f\"oo"`)
	f.Add(`"f\\oo"`)
	f.Add(`"foo" "bar"`)
	f.Add(`"foo" bar`)
	f.Add(strings.Repeat("a", 255))
	f.Add(strings.Repeat("chunk ", 200))
	f.Add(`"` + strings.Repeat("b", 600) + `"`)
	f.Add("control \x00\x01 chars \xff\xfe")
	f.Add(`"unbalanced`)

	f.Fuzz(func(t *testing.T, s string) {
		// ParseQuoted must never panic.  Errors are fine: they become
		// a report to the user.
		_, _ = txtutil.ParseQuoted(s)

		// ToChunks must produce chunks of at most 255 octets which,
		// rejoined, are the original string.
		chunks := txtutil.ToChunks(s)
		joined := strings.Join(chunks, "")
		if joined != s {
			t.Fatalf("ToChunks(%q) rejoined to %q", s, joined)
		}
		for _, chunk := range chunks {
			if len(chunk) > 255 {
				t.Fatalf("ToChunks(%q) produced a %d-octet chunk: %q", s, len(chunk), chunk)
			}
		}

		// If the input parses, re-encoding it and parsing the result
		// must yield the same string (a full round trip).
		parsed, err := txtutil.ParseQuoted(s)
		if err != nil {
			return
		}
		encoded := txtutil.EncodeQuoted(parsed)
		reparsed, err := txtutil.ParseQuoted(encoded)
		if err != nil {
			t.Fatalf("ParseQuoted(EncodeQuoted(%q)) = %q: %v", parsed, encoded, err)
		}
		if reparsed != parsed {
			t.Fatalf("round trip of %q produced %q (encoded %q)", parsed, reparsed, encoded)
		}
	})
}
