package spflib

import (
	"fmt"
	"strings"
)

// TXT outputs s as a TXT record.
func (s *SPFRecord) TXT() string {
	text := "v=spf1"
	for _, p := range s.Parts {
		text += " " + p.Text
	}
	return text
}

const maxLen = 255

// TXTSplit returns a set of txt records to use for SPF.
// pattern given is used to name all chained spf records.
// patern should include %d, which will be replaced by a counter.
// should result in fqdn after replacement
// returned map will have keys with fqdn of resulting records.
// root record will be under key "@"
func (s *SPFRecord) TXTSplit(pattern string) map[string]string {
	m := map[string]string{}
	s.split("@", pattern, 1, m)
	return m

}

func (s *SPFRecord) split(thisfqdn string, pattern string, nextIdx int, m map[string]string) {
	base := s.TXT()
	// simple case. it fits
	if len(base) <= maxLen {
		m[thisfqdn] = base
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
					m[thisfqdn] = base
					return
				}
			}
		}
		if over {
			newRec.Parts = append(newRec.Parts, part)
		}
	}
	m[thisfqdn] = thisText + tail
	newRec.split(nextFQDN, pattern, nextIdx+1, m)
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
			for _, childPart := range flattenedChild.Parts[:len(flattenedChild.Parts)-1] {
				newRec.Parts = append(newRec.Parts, childPart)
			}
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
