package spf

import (
	"fmt"
	"net"
	"strings"
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

func Parse(text string) (*SPFRecord, error) {
	if !strings.HasPrefix(text, "v=spf1 ") {
		return nil, fmt.Errorf("Not an spf record")
	}
	parts := strings.Split(text, " ")
	rec := &SPFRecord{}
	for _, part := range parts[1:] {
		p := &SPFPart{Text: part}
		rec.Parts = append(rec.Parts, p)
		if part == "~all" || part == "-all" || part == "?all" {
			//all. nothing else matters.
			break
		} else if strings.HasPrefix(part, "ip4:") || strings.HasPrefix(part, "ip6:") {
			//ip address, 0 lookups
			continue
		} else if strings.HasPrefix(part, "include:") {
			rec.Lookups++
			includeTarget := strings.TrimPrefix(part, "include:")
			subRecord, err := resolveSPF(includeTarget)
			if err != nil {
				return nil, err
			}
			p.IncludeRecord, err = Parse(subRecord)
			if err != nil {
				return nil, fmt.Errorf("In included spf: %s", err)
			}
			rec.Lookups += p.IncludeRecord.Lookups
		} else {
			return nil, fmt.Errorf("Unsupported spf part %s", part)
		}

	}
	return rec, nil
}

func resolveSPF(target string) (string, error) {
	recs, err := net.LookupTXT(target)
	if err != nil {
		return "", err
	}
	for _, r := range recs {
		if strings.HasPrefix(r, "v=spf1 ") {
			return r, nil
		}
	}
	return "", fmt.Errorf("No SPF records found for %s", target)
}
