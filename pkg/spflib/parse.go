package spflib

import (
	"fmt"
	"strings"

	"bytes"

	"io"
)

type SPFRecord struct {
	Parts []*SPFPart
}

func (s *SPFRecord) Lookups() int {
	count := 0
	for _, p := range s.Parts {
		if p.IsLookup {
			count++
		}
		if p.IncludeRecord != nil {
			count += p.IncludeRecord.Lookups()
		}
	}
	return count
}

type SPFPart struct {
	Text          string
	IsLookup      bool
	IncludeRecord *SPFRecord
	IncludeDomain string
}

var qualifiers = map[byte]bool{
	'?': true,
	'~': true,
	'-': true,
	'+': true,
}

func Parse(text string, dnsres Resolver) (*SPFRecord, error) {
	if !strings.HasPrefix(text, "v=spf1 ") {
		return nil, fmt.Errorf("Not an spf record")
	}
	parts := strings.Split(text, " ")
	rec := &SPFRecord{}
	for _, part := range parts[1:] {
		p := &SPFPart{Text: part}
		if qualifiers[part[0]] {
			part = part[1:]
		}
		rec.Parts = append(rec.Parts, p)
		if part == "all" {
			//all. nothing else matters.
			break
		} else if strings.HasPrefix(part, "a") || strings.HasPrefix(part, "mx") {
			p.IsLookup = true
		} else if strings.HasPrefix(part, "ip4:") || strings.HasPrefix(part, "ip6:") {
			//ip address, 0 lookups
			continue
		} else if strings.HasPrefix(part, "include:") {
			p.IsLookup = true
			p.IncludeDomain = strings.TrimPrefix(part, "include:")
			if dnsres != nil {
				subRecord, err := dnsres.GetSPF(p.IncludeDomain)
				if err != nil {
					return nil, err
				}
				p.IncludeRecord, err = Parse(subRecord, dnsres)
				if err != nil {
					return nil, fmt.Errorf("In included spf: %s", err)
				}
			}
		} else {
			return nil, fmt.Errorf("Unsupported spf part %s", part)
		}

	}
	return rec, nil
}

func dump(rec *SPFRecord, indent string, w io.Writer) {

	fmt.Fprintf(w, "%sTotal Lookups: %d\n", indent, rec.Lookups())
	fmt.Fprint(w, indent+"v=spf1")
	for _, p := range rec.Parts {
		fmt.Fprint(w, " "+p.Text)
	}
	fmt.Fprintln(w)
	indent += "\t"
	for _, p := range rec.Parts {
		if p.IsLookup {
			fmt.Fprintln(w, indent+p.Text)
		}
		if p.IncludeRecord != nil {
			dump(p.IncludeRecord, indent+"\t", w)
		}
	}
}

func (rec *SPFRecord) Print() string {
	w := &bytes.Buffer{}
	dump(rec, "", w)
	return w.String()
}
