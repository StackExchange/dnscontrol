package models

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

/* .Target is kind of a mess.
For simple rtypes it is the record's value. (i.e. for an A record
	it is the IP address).
For complex rtypes (like an MX record has a preference and a value)
	it might be a space-delimited string with all the parameters, or it
	might just be the hostname.

This was a bad design decision that I regret. Eventually we will eliminate this
field and replace it with setters/getters.  The setters/getters are below
so that it is easy to do things the right way in preparation.
*/

// TargetField returns the target. There may be other fields (for example
// an MX record also has a .MxPreference field.
func (rc *RecordConfig) TargetField() string {
	return rc.Target
}

// TargetSingle returns the target for types that have a single value target
// and panics for all others.
func (rc *RecordConfig) TargetSingle() string {
	if rc.Type == "MX" || rc.Type == "SRV" || rc.Type == "CAA" || rc.Type == "TLSA" || rc.Type == "TXT" {
		panic("TargetSingle called on a type with a multi-parameter rtype.")
	}
	return rc.Target
}

// TargetCombined returns a string with the various fields combined.
// For example, an MX record might output `10 mx10.example.tld`.
func (rc *RecordConfig) TargetCombined() string {
	return rc.content()
}

// TargetDebug returns a string with the various fields spelled out.
func (rc *RecordConfig) TargetDebug() string {
	return rc.String()
}

// TargetTXTQuoted returns the TXT target properly quoted.
func (rc *RecordConfig) TargetTXTQuoted() string {
	var res []string
	for _, t := range rc.TxtStrings {
		res = append(res, "\""+strings.Replace(t, "\"", "\\\"", -1)+"\"")
	}
	return strings.Join(res, " ")
}

// SetTarget sets the target, assuming that the rtype is appropriate.
func (rc *RecordConfig) SetTarget(target string) error {
	rc.Target = target
	return nil
}

// SetTargetIP sets the target to an IP, verifying this is an appropriate rtype.
func (rc *RecordConfig) SetTargetIP(target string) error {
	// TODO(tlim): Verify the rtype is appropriate for an IP.
	rc.Target = target
	return nil
}

// SetTargetFQDN sets the target to an IP, verifying this is an appropriate rtype.
func (rc *RecordConfig) SetTargetFQDN(target string) error {
	// TODO(tlim): Verify the rtype is appropriate for an hostname.
	rc.Target = target
	return nil
}

// Legacy Methods

func (rc *RecordConfig) String() (content string) {
	if rc.CombinedTarget {
		return rc.Target
	}

	content = fmt.Sprintf("%s %s %s %d", rc.Type, rc.NameFQDN, rc.Target, rc.TTL)
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "CNAME", "NS", "PTR", "TXT":
		// Nothing special.
	case "MX":
		content += fmt.Sprintf(" pref=%d", rc.MxPreference)
	case "SOA":
		content = fmt.Sprintf("%s %s %s %d", rc.Type, rc.Name, rc.Target, rc.TTL)
	case "SRV":
		content += fmt.Sprintf(" srvpriority=%d srvweight=%d srvport=%d", rc.SrvPriority, rc.SrvWeight, rc.SrvPort)
	case "TLSA":
		content += fmt.Sprintf(" tlsausage=%d tlsaselector=%d tlsamatchingtype=%d", rc.TlsaUsage, rc.TlsaSelector, rc.TlsaMatchingType)
	case "CAA":
		content += fmt.Sprintf(" caatag=%s caaflag=%d", rc.CaaTag, rc.CaaFlag)
	case "R53_ALIAS":
		content += fmt.Sprintf(" type=%s zone_id=%s", rc.R53Alias["type"], rc.R53Alias["zone_id"])
	default:
		msg := fmt.Sprintf("rc.String rtype %v unimplemented", rc.Type)
		panic(msg)
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	for k, v := range rc.Metadata {
		content += fmt.Sprintf(" %s=%s", k, v)
	}
	return content
}

// content combines Target and other fields into one string.
func (rc *RecordConfig) content() string {
	if rc.CombinedTarget {
		return rc.Target
	}

	// If this is a pseudo record, just return the target.
	if _, ok := dns.StringToType[rc.Type]; !ok {
		return rc.Target
	}

	// We cheat by converting to a dns.RR and use the String() function.
	// Sadly that function always includes a header, which we must strip out.
	// TODO(tlim): Request the dns project add a function that returns
	// the string without the header.
	rr := rc.ToRR()
	header := rr.Header().String()
	full := rr.String()
	if !strings.HasPrefix(full, header) {
		panic("dns.Hdr.String() not acting as we expect")
	}
	if rc.Type == "TXT" {
		fmt.Printf("DEBUG: txt stuff (%v) (%v)\n", full, header)
	}
	return full[len(header):]
}

// MergeToTarget combines "extra" fields into .Target, and zeros the merged fields.
func (rc *RecordConfig) MergeToTarget() {
	if rc.CombinedTarget {
		pm := strings.Join([]string{"MergeToTarget: Already collapsed: ", rc.Name, rc.Target}, " ")
		panic(pm)
	}

	// Merge "extra" fields into the Target.
	rc.Target = rc.TargetCombined()

	// Zap any fields that may have been merged.
	rc.MxPreference = 0
	rc.SrvPriority = 0
	rc.SrvWeight = 0
	rc.SrvPort = 0
	rc.CaaFlag = 0
	rc.CaaTag = ""
	rc.TlsaUsage = 0
	rc.TlsaMatchingType = 0
	rc.TlsaSelector = 0

	rc.CombinedTarget = true
}
