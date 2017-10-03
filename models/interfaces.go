package models

import "encoding/json"

//Registrar is an interface for a domain registrar. It can return a list of needed corrections to be applied in the future.
type Registrar interface {
	GetRegistrarCorrections(*DomainConfig) ([]*Correction, error)
}

//DNSServiceProvider is able to generate a set of corrections that need to be made to correct records for a domain
type DNSServiceProvider interface {
	GetNameservers(domain string) ([]*Nameserver, error)
	GetDomainCorrections(dc *DomainConfig) ([]*Correction, error)
}

//DomainCreator should be implemented by providers that have the ability to add domains to an account. the create-domains command
//can be run to ensure all domains are present before running preview/push
type DomainCreator interface {
	EnsureDomainExists(domain string) error
}

//RegistrarInitializer is a function to create a registrar. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type RegistrarInitializer func(map[string]string) (Registrar, error)

//DspInitializer is a function to create a DNS service provider. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type DspInitializer func(map[string]string, json.RawMessage) (DNSServiceProvider, error)
