package models

// DNSProvider is an interface for DNS Provider plug-ins.
type DNSProvider interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetDomainCorrections(dc *DomainConfig) ([]*Correction, error)
}

// DNSProvider3 will replace DNSProvider in 3.0.
// If you want to future-proof your code, implement these
// functions and implement GetDomainCorrections() as in
// providers/gandi_v5/gandi_v5Provider.go
//type DNSProvider3 interface {
//	GetNameservers(domain string) ([]*Nameserver, error)
//	GetZoneRecords(domain string) (Records, error)
//	PrepFoundRecords(recs Records) Records
//	PrepDesiredRecords(dc *DomainConfig)
//	GenerateDomainCorrections(dc *DomainConfig, existing Records) ([]*Correction, error)
//}

// Registrar is an interface for Registrar plug-ins.
type Registrar interface {
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
