package models

import "github.com/pkg/errors"

var ErrNotImplemented = errors.New("divide by zero")

// These interfaces are based on individual qualities of providers,
// rather than the type of provider (registrar or DNS).

// ZoneCorrector is an interface that defines a provider that can
// correct the DNS records in a Zone (i.e. domain).
type ZoneCorrector interface {
	//GetNameservers(domain string) ([]*Nameserver, error)
	GetZoneRecords(domain string) (Records, error)
	GetZoneRecordsCorrections(*DomainConfig, Records) ([]*Correction, error)
}

//// Popularizer is an interface that defines a provider that can
//// assure that a zone exists at a provider. It does not
//type Popularizer interface {
//	GenPopCorrection(domain string) ([]*Correction, error)
//}

////  is an interface that defines a provider that can
//// assure that a zone exists at a provider. It does not
//type NameserverCorrector interface {
//	GetNameservers(domain string) ([]*Nameserver, error)
//	GetNameserversCorrections(domain string) ([]*Correction, error)
//}

// Old style interfaces:
// These define the provider based on a superset of features.  In the
// future we will migrate away from these. However in the meanwhile
// there is no rush for plug-in authors to avoid them.

// DNSProvider is an interface for DNS Provider plug-ins.
type DNSProvider interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetZoneRecords(domain string) (Records, error)
	GetDomainCorrections(dc *DomainConfig) ([]*Correction, error)
}

// Registrar is an interface for Registrar plug-ins.
type Registrar interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetRegistrarCorrections(dc *DomainConfig) ([]*Correction, error)
}

// ProviderBase describes providers.
type ProviderBase struct {
	Name         string
	IsDefault    bool
	ProviderType string
}

// RegistrarInstance is a single registrar.
type RegistrarInstance struct {
	ProviderBase
	Driver Registrar
}

// DNSProviderInstance is a single DNS provider.
type DNSProviderInstance struct {
	ProviderBase
	Driver              DNSProvider
	NumberOfNameservers int
}
