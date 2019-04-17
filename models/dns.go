package models

import (
	"encoding/json"
	"strings"
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

// Nameserver describes a nameserver.
type Nameserver struct {
	Name string `json:"name"` // Normalized to a FQDN with NO trailing "."
}

func (n *Nameserver) String() string {
	return n.Name
}

// StringsToNameservers constructs a list of *Nameserver structs using a list of FQDNs.
func StringsToNameservers(nss []string) []*Nameserver {
	nservers := []*Nameserver{}
	for _, ns := range nss {
		nservers = append(nservers, &Nameserver{Name: ns})
	}
	return nservers
}

// Correction is anything that can be run. Implementation is up to the specific provider.
type Correction struct {
	F   func() error `json:"-"`
	Msg string
}

// DomainContainingFQDN finds the best domain from the dns config for the given record fqdn.
// It will chose the domain whose name is the longest suffix match for the fqdn.
func (config *DNSConfig) DomainContainingFQDN(fqdn string) *DomainConfig {
	fqdn = strings.TrimSuffix(fqdn, ".")
	longestLength := 0
	var d *DomainConfig
	for _, dom := range config.Domains {
		if (dom.Name == fqdn || strings.HasSuffix(fqdn, "."+dom.Name)) && len(dom.Name) > longestLength {
			longestLength = len(dom.Name)
			d = dom
		}
	}
	return d
}
