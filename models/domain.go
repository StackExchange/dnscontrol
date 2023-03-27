package models

import (
	"fmt"
	"strings"

	"github.com/qdm12/reprint"
	"golang.org/x/net/idna"
)

const (
	// DomainUniqueName is the full `example.com!tag` name`
	DomainUniqueName = "dnscontrol_uniquename"
	// DomainTag is the tag part of `example.com!tag` name
	DomainTag = "dnscontrol_tag"
)

// DomainConfig describes a DNS domain (technically a DNS zone).
type DomainConfig struct {
	Name             string         `json:"name"` // NO trailing "."
	RegistrarName    string         `json:"registrar"`
	DNSProviderNames map[string]int `json:"dnsProviders"`

	// Metadata[DomainUniqueName] // .Name + "!" + .Tag
	// Metadata[DomainTag] // split horizon tag
	Metadata    map[string]string `json:"meta,omitempty"`
	Records     Records           `json:"records"`
	Nameservers []*Nameserver     `json:"nameservers,omitempty"`

	EnsureAbsent Records `json:"recordsabsent,omitempty"` // ENSURE_ABSENT
	KeepUnknown  bool    `json:"keepunknown,omitempty"`   // NO_PURGE

	IgnoredNames    []*IgnoreName      `json:"ignored_names,omitempty"`
	IgnoredTargets  []*IgnoreTarget    `json:"ignored_targets,omitempty"`
	Unmanaged       []*UnmanagedConfig `json:"unmanaged,omitempty"`                      // UNMANAGED()
	UnmanagedUnsafe bool               `json:"unmanaged_disable_safety_check,omitempty"` // DISABLE_UNMANAGED_SAFETY_CHECK

	AutoDNSSEC string `json:"auto_dnssec,omitempty"` // "", "on", "off"
	//DNSSEC        bool              `json:"dnssec,omitempty"`

	// These fields contain instantiated provider instances once everything is linked up.
	// This linking is in two phases:
	// 1. Metadata (name/type) is available just from the dnsconfig. Validation can use that.
	// 2. Final driver instances are loaded after we load credentials. Any actual provider interaction requires that.
	RegistrarInstance    *RegistrarInstance     `json:"-"`
	DNSProviderInstances []*DNSProviderInstance `json:"-"`
}

// GetSplitHorizonNames returns the domain's name, uniquename, and tag.
func (dc *DomainConfig) GetSplitHorizonNames() (name, uniquename, tag string) {
	return dc.Name, dc.Metadata[DomainUniqueName], dc.Metadata[DomainTag]
}

// GetUniqueName returns the domain's uniquename.
func (dc *DomainConfig) GetUniqueName() (uniquename string) {
	return dc.Metadata[DomainUniqueName]
}

// UpdateSplitHorizonNames updates the split horizon fields
// (uniquename and tag) based on name.
func (dc *DomainConfig) UpdateSplitHorizonNames() {
	name, unique, tag := dc.GetSplitHorizonNames()

	if unique == "" {
		unique = name
	}

	if tag == "" {
		l := strings.SplitN(name, "!", 2)
		if len(l) == 2 {
			name = l[0]
			tag = l[1]
		}
	}

	dc.Name = name
	if dc.Metadata == nil {
		dc.Metadata = map[string]string{}
	}
	dc.Metadata[DomainUniqueName] = unique
	dc.Metadata[DomainTag] = tag
}

// Copy returns a deep copy of the DomainConfig.
func (dc *DomainConfig) Copy() (*DomainConfig, error) {
	newDc := &DomainConfig{}
	err := reprint.FromTo(dc, newDc) // Deep copy
	return newDc, err

	// NB(tlim): The old version of this copied the structure by gob-encoding
	// and decoding it. gob doesn't like the dc.RegisterInstance or
	// dc.DNSProviderInstances fields, so we saved a temporary copy of those,
	// nil'ed out the original, did the gob copy, and then manually copied those
	// fields using the temp variables we saved. It looked like:
	//reg, dnsps := dc.RegistrarInstance, dc.DNSProviderInstances
	//dc.RegistrarInstance, dc.DNSProviderInstances = nil, nil
	// (perform the copy)
	//dc.RegistrarInstance, dc.DNSProviderInstances = reg, dnsps
	//newDc.RegistrarInstance, newDc.DNSProviderInstances = reg, dnsps
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
		// Update the label:
		t, err := idna.ToASCII(rec.GetLabelFQDN())
		if err != nil {
			return err
		}
		rec.SetLabelFromFQDN(t, dc.Name)

		// Set the target:
		switch rec.Type { // #rtype_variations
		case "ALIAS", "MX", "NS", "CNAME", "PTR", "SRV", "URL", "URL301", "FRAME", "R53_ALIAS", "NS1_URLFWD", "AKAMAICDN", "CLOUDNS_WR":
			// These rtypes are hostnames, therefore need to be converted (unlike, for example, an AAAA record)
			t, err := idna.ToASCII(rec.GetTargetField())
			if err != nil {
				return err
			}
			rec.SetTarget(t)
		case "CF_REDIRECT", "CF_TEMP_REDIRECT", "CF_WORKER_ROUTE":
			rec.SetTarget(rec.GetTargetField())
		case "A", "AAAA", "CAA", "DS", "LOC", "NAPTR", "SOA", "SSHFP", "TXT", "TLSA", "AZURE_ALIAS", "OPENPGPKEY":
			// Nothing to do.
		default:
			return fmt.Errorf("Punycode rtype %v unimplemented", rec.Type)
		}
	}
	return nil
}
