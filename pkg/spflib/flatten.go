package spflib

import (
	"fmt"
	"strings"
)

// Chunks splits strings into arrays of a certain chunk size. We
// use this to split TXT records into 255 sized chunks for RFC 4408
// https://tools.ietf.org/html/rfc4408#section-3.1.3
// Borrowed from https://stackoverflow.com/a/61469854/11477663
func Chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string
	chunk := make([]rune, chunkSize)
	len := 0
	for _, r := range s {
		chunk[len] = r
		len++
		if len == chunkSize {
			chunks = append(chunks, string(chunk))
			len = 0
		}
	}
	if len > 0 {
		chunks = append(chunks, string(chunk[:len]))
	}
	return chunks
}

// TXT outputs s as a TXT record.
func (s *SPFRecord) TXT() string {
	text := "v=spf1"
	for _, p := range s.Parts {
		text += " " + p.Text
	}
	return text
}

// Maximum length of a single TXT string. Anything
// bigger than this will be split into multiple strings
// if the user has set a txtMaxSize length greater than 255
const txtStringLength = 255

// TXTSplit returns a set of txt records to use for SPF.
// pattern given is used to name all chained spf records.
// patern should include %d, which will be replaced by a counter.
// should result in fqdn after replacement
// returned map will have keys with fqdn of resulting records.
// root record will be under key "@"
// overhead specifies that the first split part should assume an
// overhead of that many bytes.  For example, if there are other txt
// records and you wish to reduce the first SPF record size to prevent
// DNS over TCP.
func (s *SPFRecord) TXTSplit(pattern string, overhead int, txtMaxSize int) map[string][]string {
	m := map[string][]string{}
	s.split("@", pattern, 1, m, overhead, txtMaxSize)
	return m

}

func (s *SPFRecord) split(thisfqdn string, pattern string, nextIdx int, m map[string][]string, overhead int, txtMaxSize int) {
	maxLen := txtMaxSize - overhead

	base := s.TXT()
	// simple case. it fits
	if len(base) <= maxLen {
		m[thisfqdn] = Chunks(base, txtStringLength)
		return
	}

	// we need to trim.
	// take parts while we fit
	nextFQDN := fmt.Sprintf(pattern, nextIdx)
	lastPart := s.Parts[len(s.Parts)-1]
	tail := " include:" + nextFQDN + " " + lastPart.Text
	thisText := "v=spf1"

	newRec := &SPFRecord{}
	over := false
	addedCount := 0
	for _, part := range s.Parts {
		if !over {
			if len(thisText)+1+len(part.Text)+len(tail) <= maxLen {
				thisText += " " + part.Text
				addedCount++
			} else {
				over = true
				if addedCount == 0 {
					// the first part is too big to include. We kinda have to give up here.
					m[thisfqdn] = []string{base}
					return
				}
			}
		}
		if over {
			newRec.Parts = append(newRec.Parts, part)
		}
	}

	m[thisfqdn] = Chunks(thisText+tail, txtStringLength)
	newRec.split(nextFQDN, pattern, nextIdx+1, m, 0, txtMaxSize)
}

// Flatten optimizes s.
func (s *SPFRecord) Flatten(spec string) *SPFRecord {
	newRec := &SPFRecord{}
	for _, p := range s.Parts {
		if p.IncludeRecord == nil {
			// non-includes copy straight over
			newRec.Parts = append(newRec.Parts, p)
		} else if !matchesFlatSpec(spec, p.IncludeDomain) {
			// includes that don't match get copied straight across
			newRec.Parts = append(newRec.Parts, p)
		} else {
			// flatten child recursively
			flattenedChild := p.IncludeRecord.Flatten(spec)
			// include their parts (skipping final all term)
			newRec.Parts = append(newRec.Parts, flattenedChild.Parts[:len(flattenedChild.Parts)-1]...)
		}
	}
	return newRec
}

func matchesFlatSpec(spec, fqdn string) bool {
	if spec == "*" {
		return true
	}
	for _, p := range strings.Split(spec, ",") {
		if p == fqdn {
			return true
		}
	}
	return false
}
