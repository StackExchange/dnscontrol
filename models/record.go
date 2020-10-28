package models

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
)

// RecordConfig stores a DNS record.
// Valid types:
//   Official:
//     A
//     AAAA
//     ANAME  // Technically not an official rtype yet.
//     CAA
//     CNAME
//     MX
//     NAPTR
//     NS
//     PTR
//     SRV
//     SSHFP
//     TLSA
//     TXT
//   Pseudo-Types:
//     ALIAS
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
//
// Notes about the fields:
//
// Name:
//    This is the shortname i.e. the NameFQDN without the origin suffix.
//    It should never have a trailing "."
//    It should never be null. The apex (naked domain) is stored as "@".
//    If the origin is "foo.com." and Name is "foo.com", this literally means
//        the intended FQDN is "foo.com.foo.com." (which may look odd)
// NameFQDN:
//    This is the FQDN version of Name.
//    It should never have a trailing ".".
//    NOTE: Eventually we will unexport Name/NameFQDN. Please start using
//      the setters (SetLabel/SetLabelFromFQDN) and getters (GetLabel/GetLabelFQDN).
//      as they will always work.
// Target:
//   This is the host or IP address of the record, with
//     the other related parameters (weight, priority, etc.) stored in individual
//     fields.
//   NOTE: Eventually we will unexport Target. Please start using the
//     setters (SetTarget*) and getters (GetTarget*) as they will always work.
// SubDomain:
//    This is the subdomain path, if any, imported from the configuration. If
//        present at the time of canonicalization it is inserted between the
//        Name and origin when constructing a canonical (FQDN) target.
//
// Idioms:
//  rec.Label() == "@"   // Is this record at the apex?
//
type RecordConfig struct {
	Type             string            `json:"type"` // All caps rtype name.
	Name             string            `json:"name"` // The short name. See above.
	SubDomain        string            `json:"subdomain,omitempty"`
	NameFQDN         string            `json:"-"`      // Must end with ".$origin". See above.
	Target           string            `json:"target"` // If a name, must end with "."
	TTL              uint32            `json:"ttl,omitempty"`
	Metadata         map[string]string `json:"meta,omitempty"`
	MxPreference     uint16            `json:"mxpreference,omitempty"`
	SrvPriority      uint16            `json:"srvpriority,omitempty"`
	SrvWeight        uint16            `json:"srvweight,omitempty"`
	SrvPort          uint16            `json:"srvport,omitempty"`
	CaaTag           string            `json:"caatag,omitempty"`
	CaaFlag          uint8             `json:"caaflag,omitempty"`
	DsKeyTag         uint16            `json:"dskeytag,omitempty"`
	DsAlgorithm      uint8             `json:"dsalgorithm,omitempty"`
	DsDigestType     uint8             `json:"dsdigesttype,omitempty"`
	DsDigest         string            `json:"dsdigest,omitempty"`
	NaptrOrder       uint16            `json:"naptrorder,omitempty"`
	NaptrPreference  uint16            `json:"naptrpreference,omitempty"`
	NaptrFlags       string            `json:"naptrflags,omitempty"`
	NaptrService     string            `json:"naptrservice,omitempty"`
	NaptrRegexp      string            `json:"naptrregexp,omitempty"`
	SshfpAlgorithm   uint8             `json:"sshfpalgorithm,omitempty"`
	SshfpFingerprint uint8             `json:"sshfpfingerprint,omitempty"`
	SoaMbox          string            `json:"soambox,omitempty"`
	SoaSerial        uint32            `json:"soaserial,omitempty"`
	SoaRefresh       uint32            `json:"soarefresh,omitempty"`
	SoaRetry         uint32            `json:"soaretry,omitempty"`
	SoaExpire        uint32            `json:"soaexpire,omitempty"`
	SoaMinttl        uint32            `json:"soaminttl,omitempty"`
	TlsaUsage        uint8             `json:"tlsausage,omitempty"`
	TlsaSelector     uint8             `json:"tlsaselector,omitempty"`
	TlsaMatchingType uint8             `json:"tlsamatchingtype,omitempty"`
	TxtStrings       []string          `json:"txtstrings,omitempty"` // TxtStrings stores all strings (including the first). Target stores only the first one.
	R53Alias         map[string]string `json:"r53_alias,omitempty"`
	AzureAlias       map[string]string `json:"azure_alias,omitempty"`

	Original interface{} `json:"-"` // Store pointer to provider-specific record object. Used in diffing.
}

// Copy returns a deep copy of a RecordConfig.
func (rc *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	err := copyObj(rc, newR)
	return newR, err
}

// SetLabel sets the .Name/.NameFQDN fields given a short name and origin.
// origin must not have a trailing dot: The entire code base
//   maintains dc.Name without the trailig dot. Finding a dot here means
//   something is very wrong.
// short must not have a training dot: That would mean you have
//   a FQDN, and shouldn't be using SetLabel().  Maybe SetLabelFromFQDN()?
func (rc *RecordConfig) SetLabel(short, origin string) {

	// Assertions that make sure the function is being used correctly:
	if strings.HasSuffix(origin, ".") {
		panic(fmt.Errorf("origin (%s) is not supposed to end with a dot", origin))
	}
	if strings.HasSuffix(short, ".") {
		panic(fmt.Errorf("short (%s) is not supposed to end with a dot", origin))
	}

	// TODO(tlim): We should add more validation here or in a separate validation
	// module.  We might want to check things like (\w+\.)+

	short = strings.ToLower(short)
	origin = strings.ToLower(origin)
	if short == "" || short == "@" {
		rc.Name = "@"
		rc.NameFQDN = origin
	} else {
		rc.Name = short
		rc.NameFQDN = dnsutil.AddOrigin(short, origin)
	}
}

// UnsafeSetLabelNull sets the label to "". Normally the FQDN is denoted by .Name being
// "@" however this can be used to violate that assertion. It should only be used
// on copies of a RecordConfig that is being used for non-standard things like
// Marshalling yaml.
func (rc *RecordConfig) UnsafeSetLabelNull() {
	rc.Name = ""
}

// SetLabelFromFQDN sets the .Name/.NameFQDN fields given a FQDN and origin.
// fqdn may have a trailing "." but it is not required.
// origin may not have a trailing dot.
func (rc *RecordConfig) SetLabelFromFQDN(fqdn, origin string) {

	// Assertions that make sure the function is being used correctly:
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

// ToDiffable returns a string that is comparable by a differ.
// extraMaps: a list of maps that should be included in the comparison.
func (rc *RecordConfig) ToDiffable(extraMaps ...map[string]string) string {
	content := fmt.Sprintf("%v ttl=%d", rc.GetTargetCombined(), rc.TTL)
	if rc.Type == "SOA" {
		content = fmt.Sprintf("%s %v %d %d %d %d ttl=%d", rc.Target, rc.SoaMbox, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl, rc.TTL)
		// SoaSerial is not used in comparison
	}
	for _, valueMap := range extraMaps {
		// sort the extra values map keys to perform a deterministic
		// comparison since Golang maps iteration order is not guaranteed

		// FIXME(tlim) The keys of each map is sorted per-map, not across
		// all maps. This may be intentional since we'd have no way to
		// deal with duplicates.

		keys := make([]string, 0)
		for k := range valueMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := valueMap[k]
			content += fmt.Sprintf(" %s=%s", k, v)
		}
	}
	return content
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
		rr.(*dns.A).A = rc.GetTargetIP()
	case dns.TypeAAAA:
		rr.(*dns.AAAA).AAAA = rc.GetTargetIP()
	case dns.TypeCNAME:
		rr.(*dns.CNAME).Target = rc.GetTargetField()
	case dns.TypeDS:
		rr.(*dns.DS).Algorithm = rc.DsAlgorithm
		rr.(*dns.DS).DigestType = rc.DsDigestType
		rr.(*dns.DS).Digest = rc.DsDigest
		rr.(*dns.DS).KeyTag = rc.DsKeyTag
	case dns.TypePTR:
		rr.(*dns.PTR).Ptr = rc.GetTargetField()
	case dns.TypeNAPTR:
		rr.(*dns.NAPTR).Order = rc.NaptrOrder
		rr.(*dns.NAPTR).Preference = rc.NaptrPreference
		rr.(*dns.NAPTR).Flags = rc.NaptrFlags
		rr.(*dns.NAPTR).Service = rc.NaptrService
		rr.(*dns.NAPTR).Regexp = rc.NaptrRegexp
		rr.(*dns.NAPTR).Replacement = rc.GetTargetField()
	case dns.TypeMX:
		rr.(*dns.MX).Preference = rc.MxPreference
		rr.(*dns.MX).Mx = rc.GetTargetField()
	case dns.TypeNS:
		rr.(*dns.NS).Ns = rc.GetTargetField()
	case dns.TypeSOA:
		rr.(*dns.SOA).Ns = rc.GetTargetField()
		rr.(*dns.SOA).Mbox = rc.SoaMbox
		rr.(*dns.SOA).Serial = rc.SoaSerial
		rr.(*dns.SOA).Refresh = rc.SoaRefresh
		rr.(*dns.SOA).Retry = rc.SoaRetry
		rr.(*dns.SOA).Expire = rc.SoaExpire
		rr.(*dns.SOA).Minttl = rc.SoaMinttl
	case dns.TypeSRV:
		rr.(*dns.SRV).Priority = rc.SrvPriority
		rr.(*dns.SRV).Weight = rc.SrvWeight
		rr.(*dns.SRV).Port = rc.SrvPort
		rr.(*dns.SRV).Target = rc.GetTargetField()
	case dns.TypeSSHFP:
		rr.(*dns.SSHFP).Algorithm = rc.SshfpAlgorithm
		rr.(*dns.SSHFP).Type = rc.SshfpFingerprint
		rr.(*dns.SSHFP).FingerPrint = rc.GetTargetField()
	case dns.TypeCAA:
		rr.(*dns.CAA).Flag = rc.CaaFlag
		rr.(*dns.CAA).Tag = rc.CaaTag
		rr.(*dns.CAA).Value = rc.GetTargetField()
	case dns.TypeTLSA:
		rr.(*dns.TLSA).Usage = rc.TlsaUsage
		rr.(*dns.TLSA).MatchingType = rc.TlsaMatchingType
		rr.(*dns.TLSA).Selector = rc.TlsaSelector
		rr.(*dns.TLSA).Certificate = rc.GetTargetField()
	case dns.TypeTXT:
		rr.(*dns.TXT).Txt = rc.TxtStrings
	default:
		panic(fmt.Sprintf("ToRR: Unimplemented rtype %v", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	return rr
}

// RecordKey represents a resource record in a format used by some systems.
type RecordKey struct {
	NameFQDN string
	Type     string
}

// Key converts a RecordConfig into a RecordKey.
func (rc *RecordConfig) Key() RecordKey {
	t := rc.Type
	if rc.R53Alias != nil {
		if v, ok := rc.R53Alias["type"]; ok {
			// Route53 aliases append their alias type, so that records for the same
			// label with different alias types are considered separate.
			t = fmt.Sprintf("%s_%s", t, v)
		}
	} else if rc.AzureAlias != nil {
		if v, ok := rc.AzureAlias["type"]; ok {
			// Azure aliases append their alias type, so that records for the same
			// label with different alias types are considered separate.
			t = fmt.Sprintf("%s_%s", t, v)
		}
	}
	return RecordKey{rc.NameFQDN, t}
}

// Records is a list of *RecordConfig.
type Records []*RecordConfig

// HasRecordTypeName returns True if there is a record with this rtype and name.
func (recs Records) HasRecordTypeName(rtype, name string) bool {
	for _, r := range recs {
		if r.Type == rtype && r.Name == name {
			return true
		}
	}
	return false
}

// FQDNMap returns a map of all LabelFQDNs. Useful for making a
// truthtable of labels that exist in Records.
func (recs Records) FQDNMap() (m map[string]bool) {
	m = map[string]bool{}
	for _, rec := range recs {
		m[rec.GetLabelFQDN()] = true
	}
	return m
}

// GroupedByKey returns a map of keys to records.
func (recs Records) GroupedByKey() map[RecordKey]Records {
	groups := map[RecordKey]Records{}
	for _, rec := range recs {
		groups[rec.Key()] = append(groups[rec.Key()], rec)
	}
	return groups
}

// GroupedByLabel returns a map of keys to records, and their original key order.
func (recs Records) GroupedByLabel() ([]string, map[string]Records) {
	order := []string{}
	groups := map[string]Records{}
	for _, rec := range recs {
		if _, found := groups[rec.Name]; !found {
			order = append(order, rec.Name)
		}
		groups[rec.Name] = append(groups[rec.Name], rec)
	}
	return order, groups
}

// GroupedByFQDN returns a map of keys to records, grouped by FQDN.
func (recs Records) GroupedByFQDN() ([]string, map[string]Records) {
	order := []string{}
	groups := map[string]Records{}
	for _, rec := range recs {
		namefqdn := rec.GetLabelFQDN()
		if _, found := groups[namefqdn]; !found {
			order = append(order, namefqdn)
		}
		groups[namefqdn] = append(groups[namefqdn], rec)
	}
	return order, groups
}

// PostProcessRecords does any post-processing of the downloaded DNS records.
func PostProcessRecords(recs []*RecordConfig) {
	downcase(recs)
}

// Downcase converts all labels and targets to lowercase in a list of RecordConfig.
func downcase(recs []*RecordConfig) {
	for _, r := range recs {
		r.Name = strings.ToLower(r.Name)
		r.NameFQDN = strings.ToLower(r.NameFQDN)
		switch r.Type { // #rtype_variations
		case "ANAME", "CNAME", "DS", "MX", "NS", "PTR", "NAPTR", "SRV":
			// These record types have a target that is case insensitive, so we downcase it.
			r.Target = strings.ToLower(r.Target)
		case "A", "AAAA", "ALIAS", "CAA", "IMPORT_TRANSFORM", "TLSA", "TXT", "SSHFP", "CF_REDIRECT", "CF_TEMP_REDIRECT":
			// These record types have a target that is case sensitive, or is an IP address. We leave them alone.
			// Do nothing.
		case "SOA":
			if r.Target != "DEFAULT_NOT_SET." {
				r.Target = strings.ToLower(r.Target) // .Target stores the Ns
			}
			if r.SoaMbox != "DEFAULT_NOT_SET." {
				r.SoaMbox = strings.ToLower(r.SoaMbox)
			}
		default:
			// TODO: we'd like to panic here, but custom record types complicate things.
		}
	}
}
