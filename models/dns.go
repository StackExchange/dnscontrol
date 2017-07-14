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

	"github.com/StackExchange/dnscontrol/pkg/transform"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
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
// RR:
//    This stores extra fields such as the MX record's Preference
//    and (in the future) the SRV record's Priority, Weight and Port.
//    Note that certain fields are duplicates (e.g. RecordConfig.Type
//    and RecordConfig.Hdr.Type). When this is the case, the
//    RecordConfig field is used and the other is ignored. There is no
//    attempt to keep them in sync.
//    TODO(tlim): Eventually we should refactor so that RR.Hdr fields
//    are used and can be removed from RecordConfig. Hdr.TTL is probably
//    the easiest and best place to start.  The problem with making this
//    change is that every place where RecordConfig{} is used as a constructor,
//    will need to be replaced by a custom constructor.
type RecordConfig struct {
	Type     string            `json:"type"`
	Name     string            `json:"name"`   // The short name. See below.
	Target   string            `json:"target"` // If a name, must end with "."
	TTL      uint32            `json:"ttl,omitempty"`
	Metadata map[string]string `json:"meta,omitempty"`
	NameFQDN string            `json:"-"` // Must end with ".$origin". See below.
	RR       dns.RR            `json:"-,omitempty"`

	Original interface{} `json:"-"` // Store pointer to provider-specific record object. Used in diffing.
}

func (r *RecordConfig) String() string {
	content := fmt.Sprintf("%s %s %s %d", r.Type, r.NameFQDN, r.Target, r.TTL)
	if r.Type == "MX" {
		content += fmt.Sprintf(" priority=%d", r.RR.(*dns.MX).Preference)
	}
	for k, v := range r.Metadata {
		content += fmt.Sprintf(" %s=%s", k, v)
	}
	return content
}

// MarshalJSON is an dns.RR-aware JSON marshaller.
func (r *RecordConfig) MarshalJSON() ([]byte, error) {
	type Alias RecordConfig

	var pref uint16

	switch r.Type {
	case "A", "AAAA", "ALIAS", "CF_REDIRECT", "CF_TEMP_REDIRECT", "CNAME", "IMPORT_TRANSFORM":
	case "MX":
		pref = r.RR.(*dns.MX).Preference
	case "NS":
	default:
		return nil, errors.Errorf("MarshalJSON unimplemented for type (%v)", r.Type)
	}

	return json.Marshal(&struct {
		Priority uint16 `json:"priority,omitempty"`
		*Alias
	}{
		Priority: pref,
		Alias:    (*Alias)(r),
	})
}

/// Convert RecordConfig -> dns.RR.
func (r *RecordConfig) ToRR() dns.RR {

	// Note: The label is a FQDN ending in a ".".  It will not put "@" in the Name field.

	// NB(tlim): An alternative way to do this would be
	// to create the rr via: rr := TypeToRR[x]()
	// then set the parameters. A benchmark may find that
	// faster. This was faster to implement.

	rdtype, ok := dns.StringToType[r.Type]
	if !ok {
		log.Fatalf("No such DNS type as (%#v)\n", r.Type)
	}

	hdr := dns.RR_Header{
		Name:   r.NameFQDN + ".",
		Rrtype: rdtype,
		Class:  dns.ClassINET,
		Ttl:    r.TTL,
	}

	// Handle some special cases:
	switch rdtype {
	case dns.TypeMX:
		// Has a Priority field.
		return &dns.MX{Hdr: hdr, Preference: r.RR.(*dns.MX).Preference, Mx: r.Target}
	case dns.TypeTXT:
		// Assure no problems due to quoting/unquoting:
		return &dns.TXT{Hdr: hdr, Txt: []string{r.Target}}
	default:
	}

	var ttl string
	if r.TTL == 0 {
		ttl = strconv.FormatUint(uint64(DefaultTTL), 10)
	} else {
		ttl = strconv.FormatUint(uint64(r.TTL), 10)
	}

	s := fmt.Sprintf("%s %s IN %s %s", r.NameFQDN, ttl, r.Type, r.Target)
	rc, err := dns.NewRR(s)
	if err != nil {
		log.Fatalf("NewRR rejected RecordConfig: %#v (t=%#v)\n%v\n", s, r.Target, err)
	}
	return rc
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
	Records      []*RecordConfig   `json:"records"`
	Nameservers  []*Nameserver     `json:"nameservers,omitempty"`
	KeepUnknown  bool              `json:"keepunknown"`
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
		if rec.Type == "MX" || rec.Type == "CNAME" {
			rec.Target, err = idna.ToASCII(rec.Target)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CombineMXs will merge the priority into the target field for all mx records.
// Useful for providers that desire them as one field.
func (dc *DomainConfig) CombineMXs() {
	for _, rec := range dc.Records {
		if rec.Type == "MX" {
			rec.Target = fmt.Sprintf("%d %s", rec.RR.(*dns.MX).Preference, rec.Target)
			rec.RR.(*dns.MX).Preference = 0
		}
	}
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
