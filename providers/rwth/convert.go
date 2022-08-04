package rwth

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/miekg/dns"
	"io"
	"strings"
)

// Print the generateZoneFileHelper
func (api *rwthProvider) printRecConfig(rr models.RecordConfig) string {
	// Similar to prettyzone
	// Fake types are commented out.
	prefix := ""
	_, ok := dns.StringToType[rr.Type]
	if !ok {
		prefix = ";"
	}

	// ttl
	ttl := ""
	if rr.TTL != 172800 && rr.TTL != 0 {
		ttl = fmt.Sprint(rr.TTL)
	}

	// type
	typeStr := rr.Type

	// the remaining line
	target := rr.GetTargetCombined()

	// comment
	comment := ";"

	return fmt.Sprintf("%s%s%s\n",
		prefix, prettyzone.FormatLine([]int{10, 5, 2, 5, 0}, []string{rr.NameFQDN, ttl, "IN", typeStr, target}), comment)
}

// NewRR returns custom dns.NewRR with RWTH default TTL
func NewRR(s string) (dns.RR, error) {
	if len(s) > 0 && s[len(s)-1] != '\n' { // We need a closing newline
		return ReadRR(strings.NewReader(s + "\n"))
	}
	return ReadRR(strings.NewReader(s))
}

func ReadRR(r io.Reader) (dns.RR, error) {
	zp := dns.NewZoneParser(r, ".", "")
	zp.SetDefaultTTL(172800)
	zp.SetIncludeAllowed(true)
	rr, _ := zp.Next()
	return rr, zp.Err()
}
