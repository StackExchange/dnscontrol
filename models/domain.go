package models

import (
	"fmt"

	"golang.org/x/net/idna"
)

// DomainConfig describes a DNS domain (tecnically a  DNS zone).
type DomainConfig struct {
	Name             string         `json:"name"` // NO trailing "."
	RegistrarName    string         `json:"registrar"`
	DNSProviderNames map[string]int `json:"dnsProviders"`

	Metadata       map[string]string `json:"meta,omitempty"`
	Records        Records           `json:"records"`
	Nameservers    []*Nameserver     `json:"nameservers,omitempty"`
	KeepUnknown    bool              `json:"keepunknown,omitempty"`
	IgnoredNames   []string          `json:"ignored_names,omitempty"`
	IgnoredTargets []*IgnoreTarget   `json:"ignored_targets,omitempty"`
	AutoDNSSEC     string            `json:"auto_dnssec,omitempty"` // "", "on", "off"
	//DNSSEC        bool              `json:"dnssec,omitempty"`

	// These fields contain instantiated provider instances once everything is linked up.
	// This linking is in two phases:
	// 1. Metadata (name/type) is available just from the dnsconfig. Validation can use that.
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

// Punycode will convert all records to punycode format.
// It will encode:
// - Name
// - NameFQDN
// - Target (CNAME and MX only)
func (dc *DomainConfig) Punycode() error {
	for _, rec := range dc.Records {
		t, err := idna.ToASCII(rec.GetLabelFQDN())
		if err != nil {
			return err
		}
		rec.SetLabelFromFQDN(t, dc.Name)
		switch rec.Type { // #rtype_variations
		case "ALIAS", "MX", "NS", "CNAME", "PTR", "SRV", "URL", "URL301", "FRAME", "R53_ALIAS":
			// These rtypes are hostnames, therefore need to be converted (unlike, for example, an AAAA record)
			t, err := idna.ToASCII(rec.GetTargetField())
			rec.SetTarget(t)
			if err != nil {
				return err
			}
		case "A", "AAAA", "CAA", "DS", "NAPTR", "SOA", "SSHFP", "TXT", "TLSA", "AZURE_ALIAS":
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
