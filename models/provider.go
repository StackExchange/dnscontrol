package models

import "github.com/StackExchange/dnscontrol/v3/internal/dnscontrol"

// DNSProvider is an interface for DNS Provider plug-ins.
type DNSProvider interface {
	GetNameservers(ctx dnscontrol.Context, domain string) ([]*Nameserver, error)
	GetZoneRecords(ctx dnscontrol.Context, domain string) (Records, error)
	GetDomainCorrections(ctx dnscontrol.Context, dc *DomainConfig) ([]*Correction, error)
}

// Registrar is an interface for Registrar plug-ins.
type Registrar interface {
	GetRegistrarCorrections(ctx dnscontrol.Context, dc *DomainConfig) ([]*Correction, error)
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
