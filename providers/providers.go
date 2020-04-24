package providers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Registrar is an interface for a domain registrar. It can return a list of needed corrections to be applied in the future. Implement this only if the provider is a "registrar" (i.e. can update the NS records of the parent to a domain).
type Registrar interface {
	models.Registrar
}

// DNSServiceProvider is able to generate a set of corrections that need to be made to correct records for a domain. Implement this only if the provider is a DNS Service Provider (can update records in a DNS zone).
type DNSServiceProvider interface {
	models.DNSProvider
}

// DomainCreator should be implemented by providers that have the ability to add domains to an account. the create-domains command
// can be run to ensure all domains are present before running preview/push.  Implement this only if the provider supoprts the `dnscontrol create-domain` command.
type DomainCreator interface {
	EnsureDomainExists(domain string) error
}

// ZoneLister should be implemented by providers that have the
// ability to list the zones they manage. This facilitates using the
// "get-zones" command for "all" zones.
type ZoneLister interface {
	ListZones() ([]string, error)
}

// RegistrarInitializer is a function to create a registrar. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type RegistrarInitializer func(map[string]string) (Registrar, error)

// RegistrarTypes stores initializer for each registrar.
var RegistrarTypes = map[string]RegistrarInitializer{}

// DspInitializer is a function to create a DNS service provider. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type DspInitializer func(map[string]string, json.RawMessage) (DNSServiceProvider, error)

// DNSProviderTypes stores initializer for each DSP.
var DNSProviderTypes = map[string]DspInitializer{}

// RegisterRegistrarType adds a registrar type to the registry by providing a suitable initialization function.
func RegisterRegistrarType(name string, init RegistrarInitializer, pm ...ProviderMetadata) {
	if _, ok := RegistrarTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	RegistrarTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

// RegisterDomainServiceProviderType adds a dsp to the registry with the given initialization function.
func RegisterDomainServiceProviderType(name string, init DspInitializer, pm ...ProviderMetadata) {
	if _, ok := DNSProviderTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	DNSProviderTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

// CreateRegistrar initializes a registrar instance from given credentials.
func CreateRegistrar(rType string, config map[string]string) (Registrar, error) {
	initer, ok := RegistrarTypes[rType]
	if !ok {
		return nil, fmt.Errorf("registrar type %s not declared", rType)
	}
	return initer(config)
}

// CreateDNSProvider initializes a dns provider instance from given credentials.
func CreateDNSProvider(dType string, config map[string]string, meta json.RawMessage) (DNSServiceProvider, error) {
	initer, ok := DNSProviderTypes[dType]
	if !ok {
		return nil, fmt.Errorf("DSP type %s not declared", dType)
	}
	return initer(config, meta)
}

// None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.
type None struct{}

// GetRegistrarCorrections returns corrections to update registrars.
func (n None) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}

// GetNameservers returns the current nameservers for a domain.
func (n None) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n None) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns corrections to update a domain.
func (n None) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}

func init() {
	RegisterRegistrarType("NONE", func(map[string]string) (Registrar, error) {
		return None{}, nil
	})
}

// CustomRType stores an rtype that is only valid for this DSP.
type CustomRType struct {
	Name     string
	Provider string
	RealType string
}

// RegisterCustomRecordType registers a record type that is only valid for one provider.
// provider is the registered type of provider this is valid with
// name is the record type as it will appear in the js. (should be something like $PROVIDER_FOO)
// realType is the record type it will be replaced with after validation
func RegisterCustomRecordType(name, provider, realType string) {
	customRecordTypes[name] = &CustomRType{Name: name, Provider: provider, RealType: realType}
}

// GetCustomRecordType returns a registered custom record type, or nil if none
func GetCustomRecordType(rType string) *CustomRType {
	return customRecordTypes[rType]
}

var customRecordTypes = map[string]*CustomRType{}
