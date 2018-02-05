package models

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/transform"
	"github.com/miekg/dns"
	"github.com/miekg/dns/dnsutil"
	"golang.org/x/net/idna"
)

// DefaultTTL is applied to any DNS record without an explicit TTL.
const DefaultTTL = uint32(300)

// DNSConfig describes the desired DNS configuration, usually loaded from dnsconfig.js.
type DNSConfig struct {
	Registrars         []*RegistrarConfig            `json:"registrars"`
	DNSProviders       []*DNSProviderConfig          `json:"dns_providers"`
	Domains            []*DomainConfig               `json:"domains"`
	RegistrarsByName   map[string]*RegistrarConfig   `json:"-"`
	DNSProvidersByName map[string]*DNSProviderConfig `json:"-"`
}

// FindDomain returns the *DomainConfig for domain query in config.
func (config *DNSConfig) FindDomain(query string) *DomainConfig {
	for _, b := range config.Domains {
		if b.Name == query {
			return b
		}
	}
	return nil
}

// RegistrarConfig describes a registrar.
type RegistrarConfig struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"meta,omitempty"`
}

// DNSProviderConfig describes a DNS service provider.
type DNSProviderConfig struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"meta,omitempty"`
}

// RecordConfig stores a DNS record.
// Providers are responsible for validating or normalizing the data
// that goes into a RecordConfig.
// If you update Name, you have to update NameFQDN and vice-versa.
//
// Name:
//    This is the shortname i.e. the NameFQDN without the origin suffix.
//    It should never have a trailing "."
//    It should never be null. It should store It "@", not the apex domain, not null, etc.
//    It shouldn't end with the domain origin. If the origin is "foo.com." then
//       if Name == "foo.com" then that literally means "foo.com.foo.com." is
//       the intended FQDN.
// NameFQDN:
//    This is the FQDN version of Name.
//    It should never have a trailiing ".".
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

// FixNameFQDN sets the .NameFQDN field.
func (rc *RecordConfig) FixNameFQDN(origin string) {
	rc.NameFQDN = dnsutil.AddOrigin(rc.Name, origin)
}

// Content combines Target and other fields into one string.
func (rc *RecordConfig) Content() string {
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
	return full[len(header):]
}

// MergeToTarget combines "extra" fields into .Target, and zeros the merged fields.
func (rc *RecordConfig) MergeToTarget() {
	if rc.CombinedTarget {
		pm := strings.Join([]string{"MergeToTarget: Already collapsed: ", rc.Name, rc.Target}, " ")
		panic(pm)
	}

	// Merge "extra" fields into the Target.
	rc.Target = rc.Content()

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

func atou32(s string) uint32 {
	i64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(fmt.Sprintf("atou32 failed (%v) (err=%v", s, err))
	}
	return uint32(i64)
}

// Records is a list of *RecordConfig.
type Records []*RecordConfig

// GroupedByLabel returns a map of keys to records, and their original key order.
func (r Records) GroupedByLabel() ([]string, map[string]Records) {
	order := []string{}
	groups := map[string]Records{}
	for _, rec := range r {
		if _, found := groups[rec.Name]; !found {
			order = append(order, rec.Name)
		}
		groups[rec.Name] = append(groups[rec.Name], rec)
	}
	return order, groups
}

// Grouped returns a map of keys to records.
func (r Records) Grouped() map[RecordKey]Records {
	groups := map[RecordKey]Records{}
	for _, rec := range r {
		groups[rec.Key()] = append(groups[rec.Key()], rec)
	}
	return groups
}

// RecordKey represents a resource record in a format used by some systems.
type RecordKey struct {
	Name string
	Type string
}

// Key converts a RecordConfig into a RecordKey.
func (rc *RecordConfig) Key() RecordKey {
	return RecordKey{rc.Name, rc.Type}
}

// PostProcessRecords does any post-processing of the downloaded DNS records.
func PostProcessRecords(recs []*RecordConfig) {
	Downcase(recs)
	fixTxt(recs)
}

// Downcase converts all labels and targets to lowercase in a list of RecordConfig.
func Downcase(recs []*RecordConfig) {
	for _, r := range recs {
		r.Name = strings.ToLower(r.Name)
		r.NameFQDN = strings.ToLower(r.NameFQDN)
		switch r.Type {
		case "ANAME", "CNAME", "MX", "NS", "PTR":
			r.Target = strings.ToLower(r.Target)
		case "A", "AAAA", "ALIAS", "CAA", "IMPORT_TRANSFORM", "SRV", "TLSA", "TXT", "SOA", "CF_REDIRECT", "CF_TEMP_REDIRECT":
			// Do nothing.
		default:
			// TODO: we'd like to panic here, but custom record types complicate things.
		}
	}
	return
}

// fixTxt fixes TXT records generated by providers that do not understand CanUseTXTMulti.
func fixTxt(recs []*RecordConfig) {
	for _, r := range recs {
		if r.Type == "TXT" {
			if len(r.TxtStrings) == 0 {
				r.TxtStrings = []string{r.Target}
			}
		}
	}
}

// Nameserver describes a nameserver.
type Nameserver struct {
	Name   string `json:"name"` // Normalized to a FQDN with NO trailing "."
	Target string `json:"target"`
}

// StringsToNameservers constructs a list of *Nameserver structs using a list of FQDNs.
func StringsToNameservers(nss []string) []*Nameserver {
	nservers := []*Nameserver{}
	for _, ns := range nss {
		nservers = append(nservers, &Nameserver{Name: ns})
	}
	return nservers
}

// DomainConfig describes a DNS domain (tecnically a  DNS zone).
type DomainConfig struct {
	Name             string         `json:"name"` // NO trailing "."
	RegistrarName    string         `json:"registrar"`
	DNSProviderNames map[string]int `json:"dnsProviders"`

	Metadata      map[string]string `json:"meta,omitempty"`
	Records       Records           `json:"records"`
	Nameservers   []*Nameserver     `json:"nameservers,omitempty"`
	KeepUnknown   bool              `json:"keepunknown,omitempty"`
	IgnoredLabels []string          `json:"ignored_labels,omitempty"`

	// These fields contain instantiated provider instances once everything is linked up.
	// This linking is in two phases:
	// 1. Metadata (name/type) is availible just from the dnsconfig. Validation can use that.
	// 2. Final driver instances are loaded after we load credentials. Any actual provider interaction requires that.
	RegistrarInstance    *RegistrarInstance     `json:"-"`
	DNSProviderInstances []*DNSProviderInstance `json:"-"`
}

// Copy returns a deep copy of the DomainConfig.
func (dc *DomainConfig) Copy() (*DomainConfig, error) {
	newDc := &DomainConfig{}
	// provider instances are interfaces that gob hates if you don't register them.
	// and the specific types are not gob encodable since nothing is exported.
	// should find a better solution for this now.
	//
	// current strategy: remove everything, gob copy it. Then set both to stored copy.
	reg := dc.RegistrarInstance
	dnsps := dc.DNSProviderInstances
	dc.RegistrarInstance = nil
	dc.DNSProviderInstances = nil
	err := copyObj(dc, newDc)
	dc.RegistrarInstance = reg
	newDc.RegistrarInstance = reg
	dc.DNSProviderInstances = dnsps
	newDc.DNSProviderInstances = dnsps
	return newDc, err
}

// Copy returns a deep copy of a RecordConfig.
func (rc *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	err := copyObj(rc, newR)
	return newR, err
}

// Punycode will convert all records to punycode format.
// It will encode:
// - Name
// - NameFQDN
// - Target (CNAME and MX only)
func (dc *DomainConfig) Punycode() error {
	var err error
	for _, rec := range dc.Records {
		rec.Name, err = idna.ToASCII(rec.Name)
		if err != nil {
			return err
		}
		rec.NameFQDN, err = idna.ToASCII(rec.NameFQDN)
		if err != nil {
			return err
		}
		switch rec.Type { // #rtype_variations
		case "ALIAS", "MX", "NS", "CNAME", "PTR", "SRV", "URL", "URL301", "FRAME", "R53_ALIAS":
			rec.Target, err = idna.ToASCII(rec.Target)
			if err != nil {
				return err
			}
		case "A", "AAAA", "CAA", "TXT", "TLSA":
			// Nothing to do.
		default:
			msg := fmt.Sprintf("Punycode rtype %v unimplemented", rec.Type)
			panic(msg)
			// We panic so that we quickly find any switch statements
			// that have not been updated for a new RR type.
		}
	}
	return nil
}

// CombineMXs will merge the priority into the target field for all mx records.
// Useful for providers that desire them as one field.
func (dc *DomainConfig) CombineMXs() {
	for _, rec := range dc.Records {
		if rec.Type == "MX" {
			if rec.CombinedTarget {
				pm := strings.Join([]string{"CombineMXs: Already collapsed: ", rec.Name, rec.Target}, " ")
				panic(pm)
			}
			rec.Target = fmt.Sprintf("%d %s", rec.MxPreference, rec.Target)
			rec.MxPreference = 0
			rec.CombinedTarget = true
		}
	}
}

// SplitCombinedMxValue splits a combined MX preference and target into
// separate entities, i.e. splitting "10 aspmx2.googlemail.com."
// into "10" and "aspmx2.googlemail.com.".
func SplitCombinedMxValue(s string) (preference uint16, target string, err error) {
	parts := strings.Fields(s)

	if len(parts) != 2 {
		return 0, "", fmt.Errorf("MX value %#v contains too many fields", s)
	}

	n64, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return 0, "", fmt.Errorf("MX preference %#v does not fit into a uint16", parts[0])
	}
	return uint16(n64), parts[1], nil
}

// CombineSRVs will merge the priority, weight, and port into the target field for all srv records.
// Useful for providers that desire them as one field.
func (dc *DomainConfig) CombineSRVs() {
	for _, rec := range dc.Records {
		if rec.Type == "SRV" {
			if rec.CombinedTarget {
				pm := strings.Join([]string{"CombineSRVs: Already collapsed: ", rec.Name, rec.Target}, " ")
				panic(pm)
			}
			rec.Target = fmt.Sprintf("%d %d %d %s", rec.SrvPriority, rec.SrvWeight, rec.SrvPort, rec.Target)
			rec.CombinedTarget = true
		}
	}
}

// SplitCombinedSrvValue splits a combined SRV priority, weight, port and target into
// separate entities, some DNS providers want "5" "10" 15" and "foo.com.",
// while other providers want "5 10 15 foo.com.".
func SplitCombinedSrvValue(s string) (priority, weight, port uint16, target string, err error) {
	parts := strings.Fields(s)

	if len(parts) != 4 {
		return 0, 0, 0, "", fmt.Errorf("SRV value %#v contains too many fields", s)
	}

	priorityconv, err := strconv.ParseInt(parts[0], 10, 16)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("Priority %#v does not fit into a uint16", parts[0])
	}
	weightconv, err := strconv.ParseInt(parts[1], 10, 16)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("Weight %#v does not fit into a uint16", parts[0])
	}
	portconv, err := strconv.ParseInt(parts[2], 10, 16)
	if err != nil {
		return 0, 0, 0, "", fmt.Errorf("Port %#v does not fit into a uint16", parts[0])
	}
	return uint16(priorityconv), uint16(weightconv), uint16(portconv), parts[3], nil
}

// CombineCAAs will merge the tags and flags into the target field for all CAA records.
// Useful for providers that desire them as one field.
func (dc *DomainConfig) CombineCAAs() {
	for _, rec := range dc.Records {
		if rec.Type == "CAA" {
			if rec.CombinedTarget {
				pm := strings.Join([]string{"CombineCAAs: Already collapsed: ", rec.Name, rec.Target}, " ")
				panic(pm)
			}
			rec.Target = rec.Content()
			rec.CombinedTarget = true
		}
	}
}

// SplitCombinedCaaValue parses a string listing the parts of a CAA record into its components.
func SplitCombinedCaaValue(s string) (tag string, flag uint8, value string, err error) {

	splitData := strings.SplitN(s, " ", 3)
	if len(splitData) != 3 {
		err = fmt.Errorf("Unexpected data for CAA record returned by Vultr")
		return
	}

	lflag, err := strconv.ParseUint(splitData[0], 10, 8)
	if err != nil {
		return
	}
	flag = uint8(lflag)

	tag = splitData[1]

	value = splitData[2]
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		value = value[1 : len(value)-1]
	}
	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
		value = value[1 : len(value)-1]
	}
	return
}

func copyObj(input interface{}, output interface{}) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)
	if err := enc.Encode(input); err != nil {
		return err
	}
	return dec.Decode(output)
}

// HasRecordTypeName returns True if there is a record with this rtype and name.
func (dc *DomainConfig) HasRecordTypeName(rtype, name string) bool {
	for _, r := range dc.Records {
		if r.Type == rtype && r.Name == name {
			return true
		}
	}
	return false
}

// Filter removes all records that don't match the filter f.
func (dc *DomainConfig) Filter(f func(r *RecordConfig) bool) {
	recs := []*RecordConfig{}
	for _, r := range dc.Records {
		if f(r) {
			recs = append(recs, r)
		}
	}
	dc.Records = recs
}

// InterfaceToIP returns an IP address when given a 32-bit value or a string. That is,
// dnsconfig.js output may represent IP addresses as either  a string ("1.2.3.4")
// or as an numeric value (the integer representation of the 32-bit value). This function
// converts either to a net.IP.
func InterfaceToIP(i interface{}) (net.IP, error) {
	switch v := i.(type) {
	case float64:
		u := uint32(v)
		return transform.UintToIP(u), nil
	case string:
		if ip := net.ParseIP(v); ip != nil {
			return ip, nil
		}
		return nil, fmt.Errorf("%s is not a valid ip address", v)
	default:
		return nil, fmt.Errorf("cannot convert type %s to ip", reflect.TypeOf(i))
	}
}

// Correction is anything that can be run. Implementation is up to the specific provider.
type Correction struct {
	F   func() error `json:"-"`
	Msg string
}

// MustStringToTTL converts a string to a uinet32 TTL or panics.
func MustStringToTTL(s string) uint32 {
	t, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		fmt.Printf("DEBUG:  ttl string = (%s)\n", s)
		panic(err)
	}
	return uint32(t)
}
