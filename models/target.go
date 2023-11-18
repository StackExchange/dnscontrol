package models

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

/* .target is kind of a mess.
If an rType has more than one field, one field goes in .target and the remaining are stored in bespoke fields.
Not the best design, but we're stuck with it until we re-do RecordConfig, possibly using generics.
*/

// GetTargetField returns the target. There may be other fields (for example
// an MX record also has a .MxPreference field) they are not included.
func (rc *RecordConfig) GetTargetField() string {
	return rc.target
}

// GetTargetIP returns the net.IP stored in .target.
func (rc *RecordConfig) GetTargetIP() net.IP {
	if rc.Type != "A" && rc.Type != "AAAA" {
		panic(fmt.Errorf("GetTargetIP called on an inappropriate rtype (%s)", rc.Type))
	}
	return net.ParseIP(rc.target)
}

// GetTargetCombined returns a string with the various fields combined.
// For example, an MX record might output `10 mx10.example.tld`.
func (rc *RecordConfig) GetTargetCombined() string {
	// Pseudo records:
	if _, ok := dns.StringToType[rc.Type]; !ok {
		switch rc.Type { // #rtype_variations
		case "R53_ALIAS":
			// Differentiate between multiple R53_ALIASs on the same label.
			return fmt.Sprintf("%s atype=%s zone_id=%s", rc.target, rc.R53Alias["type"], rc.R53Alias["zone_id"])
		case "AZURE_ALIAS":
			// Differentiate between multiple AZURE_ALIASs on the same label.
			return fmt.Sprintf("%s atype=%s", rc.target, rc.AzureAlias["type"])
		default:
			// Just return the target.
			return rc.target
		}
	}

	switch rc.Type {
	case "TXT":
		return rc.zoneFileQuoted()
	case "SOA":
		return fmt.Sprintf("%s %v %d %d %d %d %d", rc.target, rc.SoaMbox, rc.SoaSerial, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
	}

	return rc.zoneFileQuoted()
}

// zoneFileQuoted returns the rData as would be quoted in a zonefile.
func (rc *RecordConfig) zoneFileQuoted() string {
	// We cheat by converting to a dns.RR and use the String() function.
	// This combines all the data for us, and even does proper quoting.
	// Sadly String() always includes a header, which we must strip out.
	// TODO(tlim): Request the dns project add a function that returns
	// the string without the header.
	if rc.Type == "NAPTR" && rc.GetTargetField() == "" {
		rc.SetTarget(".")
	}
	rr := rc.ToRR()
	header := rr.Header().String()
	full := rr.String()
	if !strings.HasPrefix(full, header) {
		panic("assertion failed. dns.Hdr.String() behavior has changed in an incompatible way")
	}
	return full[len(header):]
}

// GetTargetRFC1035Quoted returns the target as it would be in an
// RFC1035-style zonefile.
// Do not use this function if RecordConfig might be a pseudo-rtype
// such as R53_ALIAS.  Use GetTargetCombined() instead.
func (rc *RecordConfig) GetTargetRFC1035Quoted() string {
	return rc.zoneFileQuoted()
}

// GetTargetDebug returns a string with the various fields spelled out.
func (rc *RecordConfig) GetTargetDebug() string {
	target := rc.target
	if rc.Type == "TXT" {
		target = fmt.Sprintf("%q", target)
	}
	content := fmt.Sprintf("%s %s %s %d", rc.Type, rc.NameFQDN, target, rc.TTL)
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "AKAMAICDN", "CNAME", "DHCID", "NS", "PTR", "TXT":
		// Nothing special.
	case "AZURE_ALIAS":
		content += fmt.Sprintf(" type=%s", rc.AzureAlias["type"])
	case "CAA":
		content += fmt.Sprintf(" caatag=%s caaflag=%d", rc.CaaTag, rc.CaaFlag)
	case "DS":
		content += fmt.Sprintf(" ds_algorithm=%d ds_keytag=%d ds_digesttype=%d ds_digest=%s", rc.DsAlgorithm, rc.DsKeyTag, rc.DsDigestType, rc.DsDigest)
	case "MX":
		content += fmt.Sprintf(" pref=%d", rc.MxPreference)
	case "NAPTR":
		content += fmt.Sprintf(" naptrorder=%d naptrpreference=%d naptrflags=%s naptrservice=%s naptrregexp=%s", rc.NaptrOrder, rc.NaptrPreference, rc.NaptrFlags, rc.NaptrService, rc.NaptrRegexp)
	case "R53_ALIAS":
		content += fmt.Sprintf(" type=%s zone_id=%s", rc.R53Alias["type"], rc.R53Alias["zone_id"])
	case "SOA":
		content = fmt.Sprintf("%s ns=%v mbox=%v serial=%v refresh=%v retry=%v expire=%v minttl=%v", rc.Type, rc.target, rc.SoaMbox, rc.SoaSerial, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
	case "SRV":
		content += fmt.Sprintf(" srvpriority=%d srvweight=%d srvport=%d", rc.SrvPriority, rc.SrvWeight, rc.SrvPort)
	case "SSHFP":
		content += fmt.Sprintf(" sshfpalgorithm=%d sshfpfingerprint=%d", rc.SshfpAlgorithm, rc.SshfpFingerprint)
	case "TLSA":
		content += fmt.Sprintf(" tlsausage=%d tlsaselector=%d tlsamatchingtype=%d", rc.TlsaUsage, rc.TlsaSelector, rc.TlsaMatchingType)
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
	rc.target = target
	return nil
}

// SetTargetIP sets the target to an IP, verifying this is an appropriate rtype.
func (rc *RecordConfig) SetTargetIP(ip net.IP) error {
	// TODO(tlim): Verify the rtype is appropriate for an IP.
	rc.SetTarget(ip.String())
	return nil
}
