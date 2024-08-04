package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/txtutil"
	"github.com/jinzhu/copier"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"github.com/qdm12/reprint"
)

// RecordConfig stores a DNS record.
// Valid types:
//
//	Official: (alphabetical)
//	  A
//	  AAAA
//	  ANAME  // Technically not an official rtype yet.
//	  CAA
//	  CNAME
//	  HTTPS
//	  LOC
//	  MX
//	  NAPTR
//	  NS
//	  PTR
//	  SOA
//	  SRV
//	  SSHFP
//	  SVCB
//	  TLSA
//	  TXT
//	Pseudo-Types: (alphabetical)
//	  ALIAS
//	  CF_REDIRECT
//	  CF_TEMP_REDIRECT
//	  CF_WORKER_ROUTE
//	  CLOUDFLAREAPI_SINGLE_REDIRECT
//	  CLOUDNS_WR
//	  FRAME
//	  IMPORT_TRANSFORM
//	  NAMESERVER
//	  NO_PURGE
//	  NS1_URLFWD
//	  PAGE_RULE
//	  PORKBUN_URLFWD
//	  PURGE
//	  URL
//	  URL301
//	  WORKER_ROUTE
//
// NOTE: All NEW record types should be prefixed with the provider name (Correct: CLOUDFLAREAPI_SINGLE_REDIRECT. Wrong: CF_REDIRECT)
//
// Notes about the fields:
//
// Name:
//
// This is the shortname i.e. the NameFQDN without the origin suffix. It should
// never have a trailing "." It should never be null. The apex (naked domain) is
// stored as "@". If the origin is "foo.com." and Name is "foo.com", this means
// the intended FQDN is "foo.com.foo.com." (which may look odd)
//
// NameFQDN:
//
// This is the FQDN version of Name. It should never have a trailing ".".
//
// NOTE: Eventually we will unexport Name/NameFQDN. Please start using
// the setters (SetLabel/SetLabelFromFQDN) and getters (GetLabel/GetLabelFQDN).
// as they will always work.
//
// target:
//
// This is the host or IP address of the record, with the other related
// parameters (weight, priority, etc.) stored in individual fields.
//
// NOTE: Eventually we will unexport Target. Please start using the
// setters (SetTarget*) and getters (GetTarget*) as they will always work.
//
// SubDomain:
//
// This is the subdomain path, if any, imported from the configuration. If
// present at the time of canonicalization it is inserted between the
// Name and origin when constructing a canonical (FQDN) target.
//
// Idioms:
//
//	rec.Label() == "@"   // Is this record at the apex?
type RecordConfig struct {
	Type      string            `json:"type"` // All caps rtype name.
	Name      string            `json:"name"` // The short name. See above.
	NameFQDN  string            `json:"-"`    // Must end with ".$origin". See above.
	SubDomain string            `json:"subdomain,omitempty"`
	target    string            // If a name, must end with "."
	TTL       uint32            `json:"ttl,omitempty"`
	Metadata  map[string]string `json:"meta,omitempty"`
	Original  interface{}       `json:"-"` // Store pointer to provider-specific record object. Used in diffing.

	// If you add a field to this struct, also add it to the list in the UnmarshalJSON function.
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
	DnskeyFlags      uint16            `json:"dnskeyflags,omitempty"`
	DnskeyProtocol   uint8             `json:"dnskeyprotocol,omitempty"`
	DnskeyAlgorithm  uint8             `json:"dnskeyalgorithm,omitempty"`
	DnskeyPublicKey  string            `json:"dnskeypublickey,omitempty"`
	LocVersion       uint8             `json:"locversion,omitempty"`
	LocSize          uint8             `json:"locsize,omitempty"`
	LocHorizPre      uint8             `json:"lochorizpre,omitempty"`
	LocVertPre       uint8             `json:"locvertpre,omitempty"`
	LocLatitude      uint32            `json:"loclatitude,omitempty"`
	LocLongitude     uint32            `json:"loclongitude,omitempty"`
	LocAltitude      uint32            `json:"localtitude,omitempty"`
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
	SvcPriority      uint16            `json:"svcpriority,omitempty"`
	SvcParams        string            `json:"svcparams,omitempty"`
	TlsaUsage        uint8             `json:"tlsausage,omitempty"`
	TlsaSelector     uint8             `json:"tlsaselector,omitempty"`
	TlsaMatchingType uint8             `json:"tlsamatchingtype,omitempty"`
	R53Alias         map[string]string `json:"r53_alias,omitempty"`
	AzureAlias       map[string]string `json:"azure_alias,omitempty"`
	UnknownTypeName  string            `json:"unknown_type_name,omitempty"`

	// Cloudflare-specific fields:
	// When these are used, .target is set to a human-readable version (only to be used for display purposes).
	CloudflareRedirect *CloudflareSingleRedirectConfig `json:"cloudflareapi_redirect,omitempty"`
}

// CloudflareSingleRedirectConfig contains info about a Cloudflare Single Redirect.
//
//	When these are used, .target is set to a human-readable version (only to be used for display purposes).
type CloudflareSingleRedirectConfig struct {
	//
	Code uint16 `json:"code,omitempty"` // 301 or 302
	// PR == PageRule
	PRWhen     string `json:"pr_when,omitempty"`
	PRThen     string `json:"pr_then,omitempty"`
	PRPriority int    `json:"pr_priority,omitempty"` // Really an identifier for the rule.
	PRDisplay  string `json:"pr_display,omitempty"`  // How is this displayed to the user (SetTarget) for CF_REDIRECT/CF_TEMP_REDIRECT
	//
	// SR == SingleRedirect
	SRName           string `json:"sr_name,omitempty"` // How is this displayed to the user
	SRWhen           string `json:"sr_when,omitempty"`
	SRThen           string `json:"sr_then,omitempty"`
	SRRRulesetID     string `json:"sr_rulesetid,omitempty"`
	SRRRulesetRuleID string `json:"sr_rulesetruleid,omitempty"`
	SRDisplay        string `json:"sr_display,omitempty"` // How is this displayed to the user (SetTarget) for CF_SINGLE_REDIRECT
}

// MarshalJSON marshals RecordConfig.
func (rc *RecordConfig) MarshalJSON() ([]byte, error) {
	recj := &struct {
		RecordConfig
		Target string `json:"target"`
	}{
		RecordConfig: *rc,
		Target:       rc.GetTargetField(),
	}
	j, err := json.Marshal(*recj)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// UnmarshalJSON unmarshals RecordConfig.
func (rc *RecordConfig) UnmarshalJSON(b []byte) error {
	recj := &struct {
		Target string `json:"target"`

		Type      string            `json:"type"` // All caps rtype name.
		Name      string            `json:"name"` // The short name. See above.
		SubDomain string            `json:"subdomain,omitempty"`
		NameFQDN  string            `json:"-"` // Must end with ".$origin". See above.
		target    string            // If a name, must end with "."
		TTL       uint32            `json:"ttl,omitempty"`
		Metadata  map[string]string `json:"meta,omitempty"`
		Original  interface{}       `json:"-"` // Store pointer to provider-specific record object. Used in diffing.
		Args      []any             `json:"args,omitempty"`

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
		DnskeyFlags      uint16            `json:"dnskeyflags,omitempty"`
		DnskeyProtocol   uint8             `json:"dnskeyprotocol,omitempty"`
		DnskeyAlgorithm  uint8             `json:"dnskeyalgorithm,omitempty"`
		DnskeyPublicKey  string            `json:"dnskeypublickey,omitempty"`
		LocVersion       uint8             `json:"locversion,omitempty"`
		LocSize          uint8             `json:"locsize,omitempty"`
		LocHorizPre      uint8             `json:"lochorizpre,omitempty"`
		LocVertPre       uint8             `json:"locvertpre,omitempty"`
		LocLatitude      int               `json:"loclatitude,omitempty"`
		LocLongitude     int               `json:"loclongitude,omitempty"`
		LocAltitude      uint32            `json:"localtitude,omitempty"`
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
		SvcPriority      uint16            `json:"svcpriority,omitempty"`
		SvcParams        string            `json:"svcparams,omitempty"`
		TlsaUsage        uint8             `json:"tlsausage,omitempty"`
		TlsaSelector     uint8             `json:"tlsaselector,omitempty"`
		TlsaMatchingType uint8             `json:"tlsamatchingtype,omitempty"`
		R53Alias         map[string]string `json:"r53_alias,omitempty"`
		AzureAlias       map[string]string `json:"azure_alias,omitempty"`
		UnknownTypeName  string            `json:"unknown_type_name,omitempty"`

		EnsureAbsent bool `json:"ensure_absent,omitempty"` // Override NO_PURGE and delete this record

		// NB(tlim): If anyone can figure out how to do this without listing all
		// the fields, please let us know!
	}{}
	if err := json.Unmarshal(b, &recj); err != nil {
		return err
	}

	// Copy the exported fields.
	copier.CopyWithOption(&rc, &recj, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	// Set each unexported field.
	rc.SetTarget(recj.Target)

	// Some sanity checks:
	if recj.Type != rc.Type {
		panic("DEBUG: TYPE NOT COPIED\n")
	}
	if recj.Type == "" {
		panic("DEBUG: TYPE BLANK\n")
	}
	if recj.Name != rc.Name {
		panic("DEBUG: NAME NOT COPIED\n")
	}

	return nil
}

// Copy returns a deep copy of a RecordConfig.
func (rc *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	// Copy the exported fields.
	err := reprint.FromTo(rc, newR) // Deep copy
	// Set each unexported field.
	newR.target = rc.target
	return newR, err
}

// SetLabel sets the .Name/.NameFQDN fields given a short name and origin.
// origin must not have a trailing dot: The entire code base maintains dc.Name
// without the trailig dot. Finding a dot here means something is very wrong.
//
// short must not have a training dot: That would mean you have a FQDN, and
// shouldn't be using SetLabel().  Maybe SetLabelFromFQDN()?
func (rc *RecordConfig) SetLabel(short, origin string) {

	// Assertions that make sure the function is being used correctly:
	if strings.HasSuffix(origin, ".") {
		panic(fmt.Errorf("origin (%s) is not supposed to end with a dot", origin))
	}
	if strings.HasSuffix(short, ".") {
		if short != "**current-domain**" {
			panic(fmt.Errorf("short (%s) is not supposed to end with a dot", origin))
		}
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

	// Trim off a trailing dot.
	fqdn = strings.TrimSuffix(fqdn, ".")

	fqdn = strings.ToLower(fqdn)
	origin = strings.ToLower(origin)
	rc.Name = dnsutil.TrimDomainName(fqdn, origin)
	rc.NameFQDN = fqdn
}

// GetLabel returns the shortname of the label associated with this RecordConfig.
// It will never end with ".". It does not need further shortening (i.e. if it
// returns "foo.com" and the domain is "foo.com" then the FQDN is actually
// "foo.com.foo.com"). It will never be "" (the apex is returned as "@").
func (rc *RecordConfig) GetLabel() string {
	return rc.Name
}

// GetLabelFQDN returns the FQDN of the label associated with this RecordConfig.
// It will not end with ".".
func (rc *RecordConfig) GetLabelFQDN() string {
	return rc.NameFQDN
}

// ToComparableNoTTL returns a comparison string. If you need to compare two
// RecordConfigs, you can simply compare the string returned by this function.
// The comparison includes all fields except TTL and any provider-specific
// metafields.  Provider-specific metafields like CF_PROXY are not the same as
// pseudo-records like ANAME or R53_ALIAS
func (rc *RecordConfig) ToComparableNoTTL() string {
	switch rc.Type {
	case "SOA":
		return fmt.Sprintf("%s %v %d %d %d %d", rc.target, rc.SoaMbox, rc.SoaRefresh, rc.SoaRetry, rc.SoaExpire, rc.SoaMinttl)
		// SoaSerial is not included because it isn't used in comparisons.
	case "TXT":
		//fmt.Fprintf(os.Stdout, "DEBUG: ToComNoTTL raw txts=%s q=%q\n", rc.target, rc.target)
		r := txtutil.EncodeQuoted(rc.target)
		//fmt.Fprintf(os.Stdout, "DEBUG: ToComNoTTL cmp txts=%s q=%q\n", r, r)
		return r
	case "UNKNOWN":
		return fmt.Sprintf("rtype=%s rdata=%s", rc.UnknownTypeName, rc.target)
	}
	return rc.GetTargetCombined()
}

// ToRR converts a RecordConfig to a dns.RR.
func (rc *RecordConfig) ToRR() dns.RR {

	// Don't call this on fake types.
	rdtype, ok := dns.StringToType[rc.Type]
	if !ok {
		log.Fatalf("No such DNS type as (%#v)\n", rc.Type)
	}

	// Magically create an RR of the correct type.
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
	case dns.TypeCAA:
		rr.(*dns.CAA).Flag = rc.CaaFlag
		rr.(*dns.CAA).Tag = rc.CaaTag
		rr.(*dns.CAA).Value = rc.GetTargetField()
	case dns.TypeCNAME:
		rr.(*dns.CNAME).Target = rc.GetTargetField()
	case dns.TypeDHCID:
		rr.(*dns.DHCID).Digest = rc.GetTargetField()
	case dns.TypeDNAME:
		rr.(*dns.DNAME).Target = rc.GetTargetField()
	case dns.TypeDS:
		rr.(*dns.DS).Algorithm = rc.DsAlgorithm
		rr.(*dns.DS).DigestType = rc.DsDigestType
		rr.(*dns.DS).Digest = rc.DsDigest
		rr.(*dns.DS).KeyTag = rc.DsKeyTag
	case dns.TypeDNSKEY:
		rr.(*dns.DNSKEY).Flags = rc.DnskeyFlags
		rr.(*dns.DNSKEY).Protocol = rc.DnskeyProtocol
		rr.(*dns.DNSKEY).Algorithm = rc.DnskeyAlgorithm
		rr.(*dns.DNSKEY).PublicKey = rc.DnskeyPublicKey
	case dns.TypeHTTPS:
		rr.(*dns.HTTPS).Priority = rc.SvcPriority
		rr.(*dns.HTTPS).Target = rc.GetTargetField()
		rr.(*dns.HTTPS).Value = rc.GetSVCBValue()
	case dns.TypeLOC:
		// fmt.Printf("ToRR long: %d, lat:%d, sz: %d, hz:%d, vt:%d\n", rc.LocLongitude, rc.LocLatitude, rc.LocSize, rc.LocHorizPre, rc.LocVertPre)
		// fmt.Printf("ToRR rc: %+v\n", *rc)
		rr.(*dns.LOC).Version = rc.LocVersion
		rr.(*dns.LOC).Longitude = rc.LocLongitude
		rr.(*dns.LOC).Latitude = rc.LocLatitude
		rr.(*dns.LOC).Altitude = rc.LocAltitude
		rr.(*dns.LOC).Size = rc.LocSize
		rr.(*dns.LOC).HorizPre = rc.LocHorizPre
		rr.(*dns.LOC).VertPre = rc.LocVertPre
	case dns.TypeMX:
		rr.(*dns.MX).Preference = rc.MxPreference
		rr.(*dns.MX).Mx = rc.GetTargetField()
	case dns.TypeNAPTR:
		rr.(*dns.NAPTR).Order = rc.NaptrOrder
		rr.(*dns.NAPTR).Preference = rc.NaptrPreference
		rr.(*dns.NAPTR).Flags = rc.NaptrFlags
		rr.(*dns.NAPTR).Service = rc.NaptrService
		rr.(*dns.NAPTR).Regexp = rc.NaptrRegexp
		rr.(*dns.NAPTR).Replacement = rc.GetTargetField()
	case dns.TypeNS:
		rr.(*dns.NS).Ns = rc.GetTargetField()
	case dns.TypePTR:
		rr.(*dns.PTR).Ptr = rc.GetTargetField()
	case dns.TypeSOA:
		rr.(*dns.SOA).Ns = rc.GetTargetField()
		rr.(*dns.SOA).Mbox = rc.SoaMbox
		rr.(*dns.SOA).Serial = rc.SoaSerial
		rr.(*dns.SOA).Refresh = rc.SoaRefresh
		rr.(*dns.SOA).Retry = rc.SoaRetry
		rr.(*dns.SOA).Expire = rc.SoaExpire
		rr.(*dns.SOA).Minttl = rc.SoaMinttl
	case dns.TypeSPF:
		rr.(*dns.SPF).Txt = rc.GetTargetTXTSegmented()
	case dns.TypeSRV:
		rr.(*dns.SRV).Priority = rc.SrvPriority
		rr.(*dns.SRV).Weight = rc.SrvWeight
		rr.(*dns.SRV).Port = rc.SrvPort
		rr.(*dns.SRV).Target = rc.GetTargetField()
	case dns.TypeSSHFP:
		rr.(*dns.SSHFP).Algorithm = rc.SshfpAlgorithm
		rr.(*dns.SSHFP).Type = rc.SshfpFingerprint
		rr.(*dns.SSHFP).FingerPrint = rc.GetTargetField()
	case dns.TypeSVCB:
		rr.(*dns.SVCB).Priority = rc.SvcPriority
		rr.(*dns.SVCB).Target = rc.GetTargetField()
		rr.(*dns.SVCB).Value = rc.GetSVCBValue()
	case dns.TypeTLSA:
		rr.(*dns.TLSA).Usage = rc.TlsaUsage
		rr.(*dns.TLSA).MatchingType = rc.TlsaMatchingType
		rr.(*dns.TLSA).Selector = rc.TlsaSelector
		rr.(*dns.TLSA).Certificate = rc.GetTargetField()
	case dns.TypeTXT:
		rr.(*dns.TXT).Txt = rc.GetTargetTXTSegmented()
	default:
		panic(fmt.Sprintf("ToRR: Unimplemented rtype %v", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}

	return rr
}

// GetDependencies returns the FQDNs on which this record dependents
func (rc *RecordConfig) GetDependencies() []string {

	switch rc.Type {
	// #rtype_variations
	case "NS", "SRV", "CNAME", "DNAME", "MX", "ALIAS", "AZURE_ALIAS", "R53_ALIAS":
		return []string{
			rc.target,
		}
	}

	return []string{}
}

// RecordKey represents a resource record in a format used by some systems.
type RecordKey struct {
	NameFQDN string
	Type     string
}

func (rk *RecordKey) String() string {
	return rk.NameFQDN + ":" + rk.Type
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

// GetSVCBValue returns the SVCB Key/Values as a list of Key/Values.
func (rc *RecordConfig) GetSVCBValue() []dns.SVCBKeyValue {
	record, err := dns.NewRR(fmt.Sprintf("%s %s %d %s %s", rc.NameFQDN, rc.Type, rc.SvcPriority, rc.target, rc.SvcParams))
	if err != nil {
		log.Fatalf("could not parse SVCB record: %s", err)
	}
	switch r := record.(type) {
	case *dns.HTTPS:
		return r.Value
	case *dns.SVCB:
		return r.Value
	}
	return nil
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

// GetByType returns the records that match rtype typeName.
func (recs Records) GetByType(typeName string) Records {
	results := Records{}
	for _, rec := range recs {
		if rec.Type == typeName {
			results = append(results, rec)
		}
	}
	return results
}

// GroupedByKey returns a map of keys to records.
func (recs Records) GroupedByKey() map[RecordKey]Records {
	groups := map[RecordKey]Records{}
	for _, rec := range recs {
		groups[rec.Key()] = append(groups[rec.Key()], rec)
	}
	return groups
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

// GetAllDependencies concatinates all dependencies of all records
func (recs Records) GetAllDependencies() []string {
	var dependencies []string
	for _, rec := range recs {
		dependencies = append(dependencies, rec.GetDependencies()...)
	}

	return dependencies
}

// PostProcessRecords does any post-processing of the downloaded DNS records.
// Deprecated. zonerecords.CorrectZoneRecords() calls Downcase directly.
func PostProcessRecords(recs []*RecordConfig) {
	Downcase(recs)
}

// Downcase converts all labels and targets to lowercase in a list of RecordConfig.
func Downcase(recs []*RecordConfig) {
	for _, r := range recs {
		r.Name = strings.ToLower(r.Name)
		r.NameFQDN = strings.ToLower(r.NameFQDN)
		switch r.Type { // #rtype_variations
		case "AKAMAICDN", "ALIAS", "AAAA", "ANAME", "CNAME", "DNAME", "DS", "DNSKEY", "MX", "NS", "NAPTR", "PTR", "SRV", "TLSA":
			// Target is case insensitive. Downcase it.
			r.target = strings.ToLower(r.target)
			// BUGFIX(tlim): isn't ALIAS in the wrong case statement?
		case "A", "CAA", "CLOUDFLAREAPI_SINGLE_REDIRECT", "CF_REDIRECT", "CF_TEMP_REDIRECT", "CF_WORKER_ROUTE", "DHCID", "IMPORT_TRANSFORM", "LOC", "SSHFP", "TXT":
			// Do nothing. (IP address or case sensitive target)
		case "SOA":
			if r.target != "DEFAULT_NOT_SET." {
				r.target = strings.ToLower(r.target) // .target stores the Ns
			}
			if r.SoaMbox != "DEFAULT_NOT_SET." {
				r.SoaMbox = strings.ToLower(r.SoaMbox)
			}
		default:
			// TODO: we'd like to panic here, but custom record types complicate things.
		}
	}
}

// CanonicalizeTargets turns Targets into FQDNs
func CanonicalizeTargets(recs []*RecordConfig, origin string) {
	originFQDN := origin + "."

	for _, r := range recs {
		switch r.Type { // #rtype_variations
		case "ALIAS", "ANAME", "CNAME", "DNAME", "DS", "DNSKEY", "MX", "NS", "NAPTR", "PTR", "SRV":
			// Target is a hostname that might be a shortname. Turn it into a FQDN.
			r.target = dnsutil.AddOrigin(r.target, originFQDN)
		case "A", "AKAMAICDN", "CAA", "DHCID", "CLOUDFLAREAPI_SINGLE_REDIRECT", "CF_REDIRECT", "CF_TEMP_REDIRECT", "CF_WORKER_ROUTE", "HTTPS", "IMPORT_TRANSFORM", "LOC", "SSHFP", "SVCB", "TLSA", "TXT":
			// Do nothing.
		case "SOA":
			if r.target != "DEFAULT_NOT_SET." {
				r.target = dnsutil.AddOrigin(r.target, originFQDN) // .target stores the Ns
			}
			if r.SoaMbox != "DEFAULT_NOT_SET." {
				r.SoaMbox = dnsutil.AddOrigin(r.SoaMbox, originFQDN)
			}
		default:
			// TODO: we'd like to panic here, but custom record types complicate things.
		}
	}
}
