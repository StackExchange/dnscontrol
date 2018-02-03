package models

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

// RecordConfig stores a DNS record.
// Valid types:
//   Official:
//     A
//     AAAA
//     ANAME
//     CAA
//     CNAME
//     MX
//     NS
//     PTR
//     SRV
//     TLSA
//     TXT
//   Pseudo-Types:
//     ALIAs
//     CF_REDIRECT
//     CF_TEMP_REDIRECT
//     FRAME
//     IMPORT_TRANSFORM
//     NAMESERVER
//     NO_PURGE
//     PAGE_RULE
//     PURGE
//     URL
//     URL301
type RecordConfig struct {
	Type             string            `json:"type"`
	Name             string            `json:"name"`   // The short name. See below.
	Target           string            `json:"target"` // If a name, must end with "."
	TTL              uint32            `json:"ttl,omitempty"`
	Metadata         map[string]string `json:"meta,omitempty"`
	NameFQDN         string            `json:"-"` // Must end with ".$origin". See below.
	MxPreference     uint16            `json:"mxpreference,omitempty"`
	SrvPriority      uint16            `json:"srvpriority,omitempty"`
	SrvWeight        uint16            `json:"srvweight,omitempty"`
	SrvPort          uint16            `json:"srvport,omitempty"`
	CaaTag           string            `json:"caatag,omitempty"`
	CaaFlag          uint8             `json:"caaflag,omitempty"`
	TlsaUsage        uint8             `json:"tlsausage,omitempty"`
	TlsaSelector     uint8             `json:"tlsaselector,omitempty"`
	TlsaMatchingType uint8             `json:"tlsamatchingtype,omitempty"`
	TxtStrings       []string          `json:"txtstrings,omitempty"` // TxtStrings stores all strings (including the first). Target stores only the first one.
	R53Alias         map[string]string `json:"r53_alias,omitempty"`

	CombinedTarget bool `json:"-"`

	Original interface{} `json:"-"` // Store pointer to provider-specific record object. Used in diffing.
}

// Name:
//    This is the shortname i.e. the NameFQDN without the origin suffix.
//    It should never have a trailing "."
//    It should never be null. The apex (naked domain) is stored as "@".
//    If the origin is "foo.com." and Name is "foo.com", this literally means
//        the intended FQDN is "foo.com.foo.com." (which may look odd)
// NameFQDN:
//    This is the FQDN version of Name.
//    It should never have a trailiing ".".
//    NOTE: Eventually we will unexport Name/NameFQDN. Please start using
//      the setters (SetLabel/SetLabelFQDN) and getters (Label/LabelFQDN).
// Target:
//   If CombinedTarget, this is a string containing all the parameters of the record.
//     For example, an MX record would store `10 foo.example.tld` in .Target.
//   If !CombinedTarget, this is the host or IP address of the record, with
//     the other related paramters (weight, priority, etc.) stored in individual
//     fields.
//   NOTE: The idea that this record can mean completely different things based
//     on the value of CombinedTarget is considered a design mistake.
//     Eventually we will unexport Target. Please start using the
//     setters (SetTarget*) and getters (Target*) as they will always work.
//
// Idioms:
//  rec.Label() == "@"   // Is this record at the apex?

// Providers are responsible for validating or normalizing the data
// that goes into a RecordConfig. The easiest way to do this is to
// always use the getters/setters (Get*(), Target*(), Label*())

// SetLabel sets the .Name/.NameFQDN fields given a short name and origin.
// origin must not have a trailing dot: The entire code base
//   maintains dc.Name without the trailig dot. Finding a dot here means
//   something is very wrong.
// short must not have a training dot: That would mean you already have
//   a FQDN, and shouldn't be using SetLabel().  Maybe SetLabelFQDN()?
func (rc *RecordConfig) SetLabel(short, origin string) {
	if strings.HasSuffix(origin, ".") {
		panic(fmt.Errorf("origin (%s) is not supposed to end with a dot", origin))
	}
	if strings.HasSuffix(short, ".") {
		panic(fmt.Errorf("short (%s) is not supposed to end with a dot", origin))
		// NB(tlim): we should never get this panic.
	}

	short = strings.ToLower(short)
	origin = strings.ToLower(origin)
	if short == "" || short == "@" {
		rc.Name = "@"
		rc.NameFQDN = origin
	} else {
		rc.Name = short
		rc.NameFQDN = dnsutil.AddOrigin(short, origin)
	}
	rc.checkIntegrity(origin)
}

// SetLabelFQDN sets the .Name/.NameFQDN fields given a FQDN and origin.
// fqdn may have a trailing "." but it is not required.
// origin may not have a trailing dot.
func (rc *RecordConfig) SetLabelFQDN(fqdn, origin string) {

	if strings.HasSuffix(origin, ".") {
		panic(fmt.Errorf("origin (%s) is not supposed to end with a dot", origin))
	}
	if strings.HasSuffix(fqdn, "..") {
		panic(fmt.Errorf("fqdn (%s) is not supposed to end with double dots", origin))
	}

	if strings.HasSuffix(fqdn, ".") {
		// Trim off a trailing dot.
		fqdn = fqdn[:len(fqdn)-1]
	}

	fqdn = strings.ToLower(fqdn)
	origin = strings.ToLower(origin)
	rc.Name = dnsutil.TrimDomainName(fqdn, origin)
	rc.NameFQDN = fqdn
	rc.checkIntegrity(origin)
}

// GetLabel returns the shortname of the label associated with this RecordConfig.
// It will never end with "."
// It does not need further shortening (i.e. if it returns "foo.com" and the
//   domain is "foo.com" then the FQDN is actually "foo.com.foo.com").
// It will never be "" (the apex is returned as "@").
func (rc *RecordConfig) GetLabel() string {
	return rc.Name
}

// GetLabelFQDN returns the FQDN of the label associated with this RecordConfig.
// It will not end with ".".
func (rc *RecordConfig) GetLabelFQDN() string {
	return rc.NameFQDN
}

// checkIntegrity verifies a RecordConfig is internally consistent or panics.
func (rc *RecordConfig) checkIntegrity(origin string) {
	if strings.HasSuffix(rc.Name, ".") {
		panic(fmt.Errorf("assertion failed: rc.Name should not end with dot (%s) (%s)", rc.Name, origin))
	}
	if strings.HasSuffix(rc.NameFQDN, ".") {
		panic(fmt.Errorf("assertion failed: rc.NameFQDN should not end with dot (%s) (%s)", rc.NameFQDN, origin))
	}
}

// ToRR converts a RecordConfig to a dns.RR.
func (rc *RecordConfig) ToRR() dns.RR {

	// Don't call this on fake types.
	rdtype, ok := dns.StringToType[rc.Type]
	if !ok {
		log.Fatalf("No such DNS type as (%#v)\n", rc.Type)
	}

	// Magicallly create an RR of the correct type.
	rr := dns.TypeToRR[rdtype]()

	// Fill in the header.
	rr.Header().Name = rc.NameFQDN + "."
	rr.Header().Rrtype = rdtype
	rr.Header().Class = dns.ClassINET
	rr.Header().Ttl = rc.TTL
	if rc.TTL == 0 {
		rr.Header().Ttl = DefaultTTL
	}

	// Fill in the data.
	switch rdtype { // #rtype_variations
	case dns.TypeA:
		rr.(*dns.A).A = net.ParseIP(rc.Target)
	case dns.TypeAAAA:
		rr.(*dns.AAAA).AAAA = net.ParseIP(rc.Target)
	case dns.TypeCNAME:
		rr.(*dns.CNAME).Target = rc.Target
	case dns.TypePTR:
		rr.(*dns.PTR).Ptr = rc.Target
	case dns.TypeMX:
		rr.(*dns.MX).Preference = rc.MxPreference
		rr.(*dns.MX).Mx = rc.Target
	case dns.TypeNS:
		rr.(*dns.NS).Ns = rc.Target
	case dns.TypeSOA:
		t := strings.Replace(rc.Target, `\ `, ` `, -1)
		parts := strings.Fields(t)
		rr.(*dns.SOA).Ns = parts[0]
		rr.(*dns.SOA).Mbox = parts[1]
		rr.(*dns.SOA).Serial = atou32(parts[2])
		rr.(*dns.SOA).Refresh = atou32(parts[3])
		rr.(*dns.SOA).Retry = atou32(parts[4])
		rr.(*dns.SOA).Expire = atou32(parts[5])
		rr.(*dns.SOA).Minttl = atou32(parts[6])
	case dns.TypeSRV:
		rr.(*dns.SRV).Priority = rc.SrvPriority
		rr.(*dns.SRV).Weight = rc.SrvWeight
		rr.(*dns.SRV).Port = rc.SrvPort
		rr.(*dns.SRV).Target = rc.Target
	case dns.TypeCAA:
		rr.(*dns.CAA).Flag = rc.CaaFlag
		rr.(*dns.CAA).Tag = rc.CaaTag
		rr.(*dns.CAA).Value = rc.Target
	case dns.TypeTLSA:
		rr.(*dns.TLSA).Usage = rc.TlsaUsage
		rr.(*dns.TLSA).MatchingType = rc.TlsaMatchingType
		rr.(*dns.TLSA).Selector = rc.TlsaSelector
		rr.(*dns.TLSA).Certificate = rc.Target
	case dns.TypeTXT:
		rr.(*dns.TXT).Txt = rc.TxtStrings
	default:
		panic(fmt.Sprintf("ToRR: Unimplemented rtype %v", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	return rr
}

// Records is a list of *RecordConfig.
type Records []*RecordConfig

// Grouped returns a map of keys to records.
func (r Records) Grouped() map[RecordKey]Records {
	groups := map[RecordKey]Records{}
	for _, rec := range r {
		groups[rec.Key()] = append(groups[rec.Key()], rec)
	}
	return groups
}
