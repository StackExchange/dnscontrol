package spflib

import (
	"fmt"
)

func (s *SPFRecord) TXT() string {
	text := "v=spf1"
	for _, p := range s.Parts {
		text += " " + p.Text
	}
	return text
}

const maxLen = 255

//TXTSplit returns a set of txt records to use for SPF.
//pattern given is used to name all chained spf records.
//patern should include %d, which will be replaced by a counter.
//should result in fqdn after replacement
//returned map will have keys with fqdn of resulting records.
//root record will be under key "@"
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
	lastPart := s.Parts[len(s.Parts)-1] // all part TODO: verify that this is an ALL clause
	tail := " include:" + nextFQDN + " " + lastPart.Text
	thisText := "v=spf1"
	newRec := &SPFRecord{}
	over := false
	for _, part := range s.Parts {
		if !over {
			if len(thisText)+1+len(part.Text)+len(tail) <= maxLen {
				thisText += " " + part.Text
			} else {
				over = true
			}
		}
		if over {
			newRec.Parts = append(newRec.Parts, part)
		}
	}
	m[thisfqdn] = thisText + tail
	newRec.split(nextFQDN, pattern, nextIdx+1, m)
}
