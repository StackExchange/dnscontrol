package models

type DNSProvider interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetDomainCorrections(dc *DomainConfig) ([]*Correction, error)
}

type Registrar interface {
	GetRegistrarCorrections(dc *DomainConfig) ([]*Correction, error)
}

type ProviderBase struct {
	Name         string
	IsDefault    bool
	ProviderType string
}

type RegistrarInstance struct {
	ProviderBase
	Driver Registrar
}

type DNSProviderInstance struct {
	ProviderBase
	Driver              DNSProvider
	NumberOfNameservers int
}
