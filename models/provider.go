package models

import "github.com/pkg/errors"

var ErrNotImplemented = errors.New("divide by zero")

// ZoneCorrector is an interface that defines a provider that can
// correct the DNS records in a Zone (i.e. domain).
type ZoneCorrector interface {
	//GetNameservers(domain string) ([]*Nameserver, error)
	GetZoneRecords(domain string) (Records, error)
	MakeZoneCorrections(*DomainConfig, Records) ([]*Correction, error)
}

// Popularizer is an interface that defines a provider that can
// assure that a zone exists at a provider. It does not
//type Popularizer interface {
//	GenPopCorrection(domain string) ([]*Correction, error)
//}

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
