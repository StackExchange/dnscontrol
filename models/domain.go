package models

import (
	"fmt"
	"strings"
	"sync"

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
	Metadata         map[string]string `json:"meta,omitempty"`
	Records          Records           `json:"records"`
	Nameservers      []*Nameserver     `json:"nameservers,omitempty"`
	NameserversMutex sync.Mutex        `json:"-"`

	EnsureAbsent Records `json:"recordsabsent,omitempty"` // ENSURE_ABSENT
	KeepUnknown  bool    `json:"keepunknown,omitempty"`   // NO_PURGE

	Unmanaged       []*UnmanagedConfig `json:"unmanaged,omitempty"`                      // IGNORE()
	UnmanagedUnsafe bool               `json:"unmanaged_disable_safety_check,omitempty"` // DISABLE_IGNORE_SAFETY_CHECK

	AutoDNSSEC string `json:"auto_dnssec,omitempty"` // "", "on", "off"
	//DNSSEC        bool              `json:"dnssec,omitempty"`

	// These fields contain instantiated provider instances once everything is linked up.
	// This linking is in two phases:
	// 1. Metadata (name/type) is available just from the dnsconfig. Validation can use that.
	// 2. Final driver instances are loaded after we load credentials. Any actual provider interaction requires that.
	RegistrarInstance    *RegistrarInstance     `json:"-"`
	DNSProviderInstances []*DNSProviderInstance `json:"-"`

	// Raw user-input from dnsconfig.js that will be processed into RecordConfigs later:
	RawRecords []RawRecordConfig `json:"rawrecords,omitempty"`

	// Pending work to do for each provider.  Provider may be a registrar or DSP.
	pendingCorrectionsMutex  sync.Mutex                 // Protect pendingCorrections*
	pendingCorrections       map[string]([]*Correction) // Work to be done for each provider
	pendingCorrectionsOrder  []string                   // Call the providers in this order
	pendingActualChangeCount map[string](int)           // Number of changes to report (cumulative)
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
		case "ALIAS", "MX", "NS", "CNAME", "DNAME", "PTR", "SRV", "URL", "URL301", "FRAME", "R53_ALIAS", "NS1_URLFWD", "AKAMAICDN", "CLOUDNS_WR", "PORKBUN_URLFWD":
			// These rtypes are hostnames, therefore need to be converted (unlike, for example, an AAAA record)
			t, err := idna.ToASCII(rec.GetTargetField())
			if err != nil {
				return err
			}
			rec.SetTarget(t)
		case "CLOUDFLAREAPI_SINGLE_REDIRECT", "CF_REDIRECT", "CF_TEMP_REDIRECT", "CF_WORKER_ROUTE":
			rec.SetTarget(rec.GetTargetField())
		case "A", "AAAA", "CAA", "DHCID", "DNSKEY", "DS", "HTTPS", "LOC", "NAPTR", "SOA", "SSHFP", "SVCB", "TXT", "TLSA", "AZURE_ALIAS":
			// Nothing to do.
		default:
			return fmt.Errorf("Punycode rtype %v unimplemented", rec.Type)
		}
	}
	return nil
}

// StoreCorrections accumulates corrections in a thread-safe way.
func (dc *DomainConfig) StoreCorrections(providerName string, corrections []*Correction) {
	dc.pendingCorrectionsMutex.Lock()
	defer dc.pendingCorrectionsMutex.Unlock()

	if dc.pendingCorrections == nil {
		// First time storing anything.
		dc.pendingCorrections = make(map[string]([]*Correction))
		dc.pendingCorrections[providerName] = corrections
		dc.pendingCorrectionsOrder = []string{providerName}
	} else if c, ok := dc.pendingCorrections[providerName]; !ok {
		// First time key used
		dc.pendingCorrections[providerName] = corrections
		dc.pendingCorrectionsOrder = []string{providerName}
	} else {
		// Add to existing.
		dc.pendingCorrections[providerName] = append(c, corrections...)
		dc.pendingCorrectionsOrder = append(dc.pendingCorrectionsOrder, providerName)
	}
}

// GetCorrections returns the accumulated corrections for providerName.
func (dc *DomainConfig) GetCorrections(providerName string) []*Correction {
	dc.pendingCorrectionsMutex.Lock()
	defer dc.pendingCorrectionsMutex.Unlock()

	if dc.pendingCorrections == nil {
		// First time storing anything.
		return nil
	}
	if c, ok := dc.pendingCorrections[providerName]; ok {
		return c
	}
	return nil
}

// IncrementChangeCount accumulates change count in a thread-safe way.
func (dc *DomainConfig) IncrementChangeCount(providerName string, delta int) {
	dc.pendingCorrectionsMutex.Lock()
	defer dc.pendingCorrectionsMutex.Unlock()

	if dc.pendingActualChangeCount == nil {
		// First time storing anything.
		dc.pendingActualChangeCount = make(map[string](int))
	}
	dc.pendingActualChangeCount[providerName] += delta
}

// GetChangeCount accumulates change count in a thread-safe way.
func (dc *DomainConfig) GetChangeCount(providerName string) int {
	dc.pendingCorrectionsMutex.Lock()
	defer dc.pendingCorrectionsMutex.Unlock()

	return dc.pendingActualChangeCount[providerName]
}
