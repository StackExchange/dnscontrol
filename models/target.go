package models

import (
	"fmt"
	"net"
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

// GetTargetField returns the target. There may be other fields (for example
// an MX record also has a .MxPreference field.
func (rc *RecordConfig) GetTargetField() string {
	if rc.Type == "TXT" {
		return strings.Join(rc.TxtStrings, "")
	}
	return rc.Target
}

// // GetTargetSingle returns the target for types that have a single value target
// // and panics for all others.
// func (rc *RecordConfig) GetTargetSingle() string {
// 	if rc.Type == "MX" || rc.Type == "SRV" || rc.Type == "CAA" || rc.Type == "TLSA" || rc.Type == "TXT" {
// 		panic("TargetSingle called on a type with a multi-parameter rtype.")
// 	}
// 	return rc.Target
// }

// GetTargetIP returns the net.IP stored in Target.
func (rc *RecordConfig) GetTargetIP() net.IP {
	if rc.Type != "A" && rc.Type != "AAAA" {
		panic(fmt.Errorf("GetTargetIP called on an inappropriate rtype (%s)", rc.Type))
	}
	return net.ParseIP(rc.Target)
}

// GetTargetCombined returns a string with the various fields combined.
// For example, an MX record might output `10 mx10.example.tld`.
func (rc *RecordConfig) GetTargetCombined() string {
	// Pseudo records:
	if _, ok := dns.StringToType[rc.Type]; !ok {
		switch rc.Type { // #rtype_variations
		case "R53_ALIAS":
			// Differentiate between multiple R53_ALIASs on the same label.
			return fmt.Sprintf("%s atype=%s zone_id=%s", rc.Target, rc.R53Alias["type"], rc.R53Alias["zone_id"])
		case "AZURE_ALIAS":
			// Differentiate between multiple AZURE_ALIASs on the same label.
			return fmt.Sprintf("%s atype=%s", rc.Target, rc.AzureAlias["type"])
		case "SOA":
			return fmt.Sprintf("%s %v %d %d %d %d %d", rc.Target, rc.SoaMbox, rc.SoaSerial, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
		default:
			// Just return the target.
			return rc.Target
		}
	}

	// We cheat by converting to a dns.RR and use the String() function.
	// This combines all the data for us, and even does proper quoting.
	// Sadly String() always includes a header, which we must strip out.
	// TODO(tlim): Request the dns project add a function that returns
	// the string without the header.
	rr := rc.ToRR()
	header := rr.Header().String()
	full := rr.String()
	if !strings.HasPrefix(full, header) {
		panic("assertion failed. dns.Hdr.String() behavior has changed in an incompatible way")
	}
	return full[len(header):]
}

// GetTargetSortable returns a string that is sortable.
func (rc *RecordConfig) GetTargetSortable() string {
	return rc.GetTargetDebug()
}

// GetTargetDebug returns a string with the various fields spelled out.
func (rc *RecordConfig) GetTargetDebug() string {
	content := fmt.Sprintf("%s %s %s %d", rc.Type, rc.NameFQDN, rc.Target, rc.TTL)
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "CNAME", "NS", "PTR", "TXT":
		// Nothing special.
	case "DS":
		content += fmt.Sprintf(" ds_algorithm=%d ds_keytag=%d ds_digesttype=%d ds_digest=%s", rc.DsAlgorithm, rc.DsKeyTag, rc.DsDigestType, rc.DsDigest)
	case "NAPTR":
		content += fmt.Sprintf(" naptrorder=%d naptrpreference=%d naptrflags=%s naptrservice=%s naptrregexp=%s", rc.NaptrOrder, rc.NaptrPreference, rc.NaptrFlags, rc.NaptrService, rc.NaptrRegexp)
	case "MX":
		content += fmt.Sprintf(" pref=%d", rc.MxPreference)
	case "SOA":
		content = fmt.Sprintf("%s ns=%v mbox=%v serial=%v refresh=%v retry=%v expire=%v minttl=%v", rc.Type, rc.Target, rc.SoaMbox, rc.SoaSerial, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
	case "SRV":
		content += fmt.Sprintf(" srvpriority=%d srvweight=%d srvport=%d", rc.SrvPriority, rc.SrvWeight, rc.SrvPort)
	case "SSHFP":
		content += fmt.Sprintf(" sshfpalgorithm=%d sshfpfingerprint=%d", rc.SshfpAlgorithm, rc.SshfpFingerprint)
	case "TLSA":
		content += fmt.Sprintf(" tlsausage=%d tlsaselector=%d tlsamatchingtype=%d", rc.TlsaUsage, rc.TlsaSelector, rc.TlsaMatchingType)
	case "CAA":
		content += fmt.Sprintf(" caatag=%s caaflag=%d", rc.CaaTag, rc.CaaFlag)
	case "R53_ALIAS":
		content += fmt.Sprintf(" type=%s zone_id=%s", rc.R53Alias["type"], rc.R53Alias["zone_id"])
	case "AZURE_ALIAS":
		content += fmt.Sprintf(" type=%s", rc.AzureAlias["type"])
	default:
		panic(fmt.Errorf("rc.String rtype %v unimplemented", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	for k, v := range rc.Metadata {
		content += fmt.Sprintf(" %s=%s", k, v)
	}
	return content
}

// SetTarget sets the target, assuming that the rtype is appropriate.
func (rc *RecordConfig) SetTarget(target string) error {
	rc.Target = target
	return nil
}

// SetTargetIP sets the target to an IP, verifying this is an appropriate rtype.
func (rc *RecordConfig) SetTargetIP(ip net.IP) error {
	// TODO(tlim): Verify the rtype is appropriate for an IP.
	rc.SetTarget(ip.String())
	return nil
}

// // SetTargetFQDN sets the target to a string, verifying this is an appropriate rtype.
// func (rc *RecordConfig) SetTargetFQDN(target string) error {
// 	// TODO(tlim): Verify the rtype is appropriate for an hostname.
// 	rc.Target = target
// 	return nil
// }
