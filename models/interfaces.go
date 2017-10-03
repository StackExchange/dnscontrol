package models

import "encoding/json"

//RegistrarDriver is an interface for a domain registrar. It can return a list of needed corrections to be applied in the future.
type RegistrarDriver interface {
	GetRegistrarCorrections(*DomainConfig) ([]*Correction, error)
}

// Provider is a base interface for basic provider information
type Provider interface {
	Name() string
	RunByDefault() bool
}

// Registrar is a RegistrarDriver with its' associated instance information
type Registrar interface {
	RegistrarDriver
	Provider
}

//DNSServiceProviderDriver is able to generate a set of corrections that need to be made to correct records for a domain
type DNSServiceProviderDriver interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetDomainCorrections(dc *DomainConfig) ([]*Correction, error)
}

// DNSProvider is a DNSServiceProviderDriver with its' associated instance information
type DNSProvider interface {
	DNSServiceProviderDriver
	Provider
	NumberOfNameserversToUse() int
}

//DomainCreator should be implemented by providers that have the ability to add domains to an account. the create-domains command
//can be run to ensure all domains are present before running preview/push
type DomainCreator interface {
	EnsureDomainExists(domain string) error
}

//RegistrarInitializer is a function to create a registrar. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type RegistrarInitializer func(map[string]string) (RegistrarDriver, error)

//DspInitializer is a function to create a DNS service provider. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type DspInitializer func(map[string]string, json.RawMessage) (DNSServiceProviderDriver, error)
