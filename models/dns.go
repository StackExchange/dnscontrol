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
	"golang.org/x/net/idna"
)

const DefaultTTL = uint32(300)

type DNSConfig struct {
	Registrars   []*RegistrarConfig   `json:"registrars"`
	DNSProviders []*DNSProviderConfig `json:"dns_providers"`
	Domains      []*DomainConfig      `json:"domains"`
}

func (config *DNSConfig) FindDomain(query string) *DomainConfig {
	for _, b := range config.Domains {
		if b.Name == query {
			return b
		}
	}
	return nil
}

type RegistrarConfig struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Metadata json.RawMessage `json:"meta,omitempty"`
}

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
type RecordConfig struct {
	Type             string            `json:"type"`
	Name             string            `json:"name"`   // The short name. See below.
	Target           string            `json:"target"` // If a name, must end with "."
	TTL              uint32            `json:"ttl,omitempty"`
	Metadata         map[string]string `json:"meta,omitempty"`
	NameFQDN         string            `json:"-"`                      // Must end with ".$origin". See below.
	MxPreference     uint16            `json:"mxpreference,omitempty"` // FIXME(tlim): Rename to MxPreference
	SrvPriority      uint16            `json:"srvpriority,omitempty"`
	SrvWeight        uint16            `json:"srvweight,omitempty"`
	SrvPort          uint16            `json:"srvport,omitempty"`
	CaaTag           string            `json:"caatag,omitempty"`
	CaaFlag          uint8             `json:"caaflag,omitempty"`
	TlsaUsage        uint8             `json:"tlsausage,omitempty"`
	TlsaSelector     uint8             `json:"tlsaselector,omitempty"`
	TlsaMatchingType uint8             `json:"tlsamatchingtype,omitempty"`

	CombinedTarget bool `json:"-"`

	Original interface{} `json:"-"` // Store pointer to provider-specific record object. Used in diffing.
}

func (r *RecordConfig) String() (content string) {
	if r.CombinedTarget {
		return r.Target
	}

	content = fmt.Sprintf("%s %s %s %d", r.Type, r.NameFQDN, r.Target, r.TTL)
	switch r.Type { // #rtype_variations
	case "A", "AAAA", "CNAME", "PTR", "TXT":
		// Nothing special.
	case "MX":
		content += fmt.Sprintf(" priority=%d", r.MxPreference)
	case "SOA":
		content = fmt.Sprintf("%s %s %s %d", r.Type, r.Name, r.Target, r.TTL)
	case "CAA":
		content += fmt.Sprintf(" caatag=%s caaflag=%d", r.CaaTag, r.CaaFlag)
	default:
		msg := fmt.Sprintf("rc.String rtype %v unimplemented", r.Type)
		panic(msg)
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	for k, v := range r.Metadata {
		content += fmt.Sprintf(" %s=%s", k, v)
	}
	return content
}

// Content combines Target and other fields into one string.
func (r *RecordConfig) Content() string {
	if r.CombinedTarget {
		return r.Target
	}

	// If this is a pseudo record, just return the target.
	if _, ok := dns.StringToType[r.Type]; !ok {
		return r.Target
	}

	// We cheat by converting to a dns.RR and use the String() function.
	// Sadly that function always includes a header, which we must strip out.
	// TODO(tlim): Request the dns project add a function that returns
	// the string without the header.
	rr := r.ToRR()
	header := rr.Header().String()
	full := rr.String()
	if !strings.HasPrefix(full, header) {
		panic("dns.Hdr.String() not acting as we expect")
	}
	return full[len(header):]
}

// MergeToTarget combines "extra" fields into .Target, and zeros the merged fields.
func (r *RecordConfig) MergeToTarget() {
	if r.CombinedTarget {
		pm := strings.Join([]string{"MergeToTarget: Already collapsed: ", r.Name, r.Target}, " ")
		panic(pm)
	}

	// Merge "extra" fields into the Target.
	r.Target = r.Content()

	// Zap any fields that may have been merged.
	r.MxPreference = 0
	r.SrvPriority = 0
	r.SrvWeight = 0
	r.SrvPort = 0
	r.CaaFlag = 0
	r.CaaTag = ""
	r.TlsaUsage = 0
	r.TlsaMatchingType = 0
	r.TlsaSelector = 0

	r.CombinedTarget = true
}

/// Convert RecordConfig -> dns.RR.
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
		rr.(*dns.TXT).Txt = []string{rc.Target}
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

type Records []*RecordConfig

func (r Records) Grouped() map[RecordKey]Records {
	groups := map[RecordKey]Records{}
	for _, rec := range r {
		groups[rec.Key()] = append(groups[rec.Key()], rec)
	}
	return groups
}

type RecordKey struct {
	Name string
	Type string
}

func (r *RecordConfig) Key() RecordKey {
	return RecordKey{r.Name, r.Type}
}

type Nameserver struct {
	Name   string `json:"name"` // Normalized to a FQDN with NO trailing "."
	Target string `json:"target"`
}

func StringsToNameservers(nss []string) []*Nameserver {
	nservers := []*Nameserver{}
	for _, ns := range nss {
		nservers = append(nservers, &Nameserver{Name: ns})
	}
	return nservers
}

type DomainConfig struct {
	Name         string            `json:"name"` // NO trailing "."
	Registrar    string            `json:"registrar"`
	DNSProviders map[string]int    `json:"dnsProviders"`
	Metadata     map[string]string `json:"meta,omitempty"`
	Records      Records           `json:"records"`
	Nameservers  []*Nameserver     `json:"nameservers,omitempty"`
	KeepUnknown  bool              `json:"keepunknown,omitempty"`
}

func (dc *DomainConfig) Copy() (*DomainConfig, error) {
	newDc := &DomainConfig{}
	err := copyObj(dc, newDc)
	return newDc, err
}

func (r *RecordConfig) Copy() (*RecordConfig, error) {
	newR := &RecordConfig{}
	err := copyObj(r, newR)
	return newR, err
}

//Punycode will convert all records to punycode format.
//It will encode:
//- Name
//- NameFQDN
//- Target (CNAME and MX only)
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
		case "ALIAS", "MX", "NS", "CNAME", "PTR", "SRV", "URL", "URL301", "FRAME":
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

//SplitCombinedSrvValue splits a combined SRV priority, weight, port and target into
//separate entities, some DNS providers want "5" "10" 15" and "foo.com.",
//while other providers want "5 10 15 foo.com.".
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

func copyObj(input interface{}, output interface{}) error {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	dec := gob.NewDecoder(buf)
	if err := enc.Encode(input); err != nil {
		return err
	}
	if err := dec.Decode(output); err != nil {
		return err
	}
	return nil
}

func (dc *DomainConfig) HasRecordTypeName(rtype, name string) bool {
	for _, r := range dc.Records {
		if r.Type == rtype && r.Name == name {
			return true
		}
	}
	return false
}

func (dc *DomainConfig) Filter(f func(r *RecordConfig) bool) {
	recs := []*RecordConfig{}
	for _, r := range dc.Records {
		if f(r) {
			recs = append(recs, r)
		}
	}
	dc.Records = recs
}

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
		return nil, fmt.Errorf("Cannot convert type %s to ip.", reflect.TypeOf(i))
	}
}

//Correction is anything that can be run. Implementation is up to the specific provider.
type Correction struct {
	F   func() error `json:"-"`
	Msg string
}
