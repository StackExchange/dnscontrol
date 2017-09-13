package spflib

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/dnsresolver"
)

type SPFRecord struct {
	Lookups int
	Parts   []*SPFPart
}

type SPFPart struct {
	Text          string
	Lookups       int
	IncludeRecord *SPFRecord
}

func Lookup(target string, dnsres dnsresolver.DnsResolver) (string, error) {
	txts, err := dnsres.GetTxt(target)
	if err != nil {
		return "", err
	}
	var result []string
	for _, txt := range txts {
		if strings.HasPrefix(txt, "v=spf1 ") {
			result = append(result, txt)
		}
	}
	if len(result) == 0 {
		return "", fmt.Errorf("%s has no spf TXT records", target)
	}
	if len(result) != 1 {
		return "", fmt.Errorf("%s has multiple spf TXT records", target)
	}
	return result[0], nil
}

var qualifiers = map[byte]bool{
	'?': true,
	'~': true,
	'-': true,
	'+': true,
}

func Parse(text string, dnsres dnsresolver.DnsResolver) (*SPFRecord, error) {
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
			rec.Lookups++
			p.Lookups = 1
		} else if strings.HasPrefix(part, "ip4:") || strings.HasPrefix(part, "ip6:") {
			//ip address, 0 lookups
			continue
		} else if strings.HasPrefix(part, "include:") {
			rec.Lookups++
			includeTarget := strings.TrimPrefix(part, "include:")
			subRecord, err := Lookup(includeTarget, dnsres)
			if err != nil {
				return nil, err
			}
			p.IncludeRecord, err = Parse(subRecord, dnsres)
			if err != nil {
				return nil, fmt.Errorf("In included spf: %s", err)
			}
			rec.Lookups += p.IncludeRecord.Lookups
			p.Lookups = p.IncludeRecord.Lookups + 1
		} else {
			return nil, fmt.Errorf("Unsupported spf part %s", part)
		}

	}
	return rec, nil
}

// DumpSPF outputs an SPFRecord and related data for debugging purposes.
func DumpSPF(rec *SPFRecord, indent string) {
	fmt.Printf("%sTotal Lookups: %d\n", indent, rec.Lookups)
	fmt.Print(indent + "v=spf1")
	for _, p := range rec.Parts {
		fmt.Print(" " + p.Text)
	}
	fmt.Println()
	indent += "\t"
	for _, p := range rec.Parts {
		if p.Lookups > 0 {
			fmt.Println(indent + p.Text)
		}
		if p.IncludeRecord != nil {
			DumpSPF(p.IncludeRecord, indent+"\t")
		}
	}
}
