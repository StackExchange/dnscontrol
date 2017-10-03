package providers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/models"
)

var registrarTypes = map[string]models.RegistrarInitializer{}

var dnsProviderTypes = map[string]models.DspInitializer{}

//RegisterRegistrarType adds a registrar type to the registry by providing a suitable initialization function.
func RegisterRegistrarType(name string, init models.RegistrarInitializer, pm ...ProviderMetadata) {
	if _, ok := registrarTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	registrarTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

//RegisterDomainServiceProviderType adds a dsp to the registry with the given initialization function.
func RegisterDomainServiceProviderType(name string, init models.DspInitializer, pm ...ProviderMetadata) {
	if _, ok := dnsProviderTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	dnsProviderTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

func createRegistrar(rType string, config map[string]string) (models.RegistrarDriver, error) {
	initer, ok := registrarTypes[rType]
	if !ok {
		return nil, fmt.Errorf("Registrar type %s not declared.", rType)
	}
	return initer(config)
}

func CreateDNSProvider(dType string, config map[string]string, meta json.RawMessage) (models.DNSServiceProviderDriver, error) {
	initer, ok := dnsProviderTypes[dType]
	if !ok {
		return nil, fmt.Errorf("DSP type %s not declared", dType)
	}
	return initer(config, meta)
}

//CreateRegistrars will load all registrars from the dns config, and create instances of the correct type using data from
//the provider config to load relevant keys and options.
func CreateRegistrars(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]models.RegistrarDriver, error) {
	regs := map[string]models.RegistrarDriver{}
	for _, reg := range d.Registrars {
		rawMsg, ok := providerConfigs[reg.Name]
		if !ok && reg.Type != "NONE" {
			return nil, fmt.Errorf("Registrar %s not listed in creds.json file.", reg.Name)
		}
		registrar, err := createRegistrar(reg.Type, rawMsg)
		if err != nil {
			return nil, err
		}
		regs[reg.Name] = registrar
	}
	return regs, nil
}

func CreateDsps(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]models.DNSServiceProviderDriver, error) {
	dsps := map[string]models.DNSServiceProviderDriver{}
	for _, dsp := range d.DNSProviders {
		vals := providerConfigs[dsp.Name]
		provider, err := CreateDNSProvider(dsp.Type, vals, dsp.Metadata)
		if err != nil {
			return nil, err
		}
		dsps[dsp.Name] = provider
	}
	return dsps, nil
}

// CreateProviders will initialize all dns providers and registrars, and store them in all domains as needed.
func CreateProviders(d *models.DNSConfig, providerConfigs map[string]map[string]string) error {
	dnsProvidersByName := map[string]models.DNSServiceProviderDriver{}
	registrarsByName := map[string]models.RegistrarDriver{}
	// create dns providers
	for _, dnsProvider := range d.DNSProviders {
		vals := providerConfigs[dnsProvider.Name]
		initer, ok := dnsProviderTypes[dnsProvider.Type]
		if !ok {
			return fmt.Errorf("DNS Provider type %s not declared", dnsProvider.Type)
		}
		if provider, err := initer(vals, dnsProvider.Metadata); err == nil {
			dnsProvidersByName[dnsProvider.Name] = provider
		} else {
			return err
		}
	}
	// create registrars
	for _, reg := range d.Registrars {
		vals := providerConfigs[reg.Name]
		initer, ok := registrarTypes[reg.Type]
		if !ok {
			return fmt.Errorf("Registrar type %s not declared", reg.Type)
		}
		if provider, err := initer(vals); err == nil {
			registrarsByName[reg.Name] = provider
		} else {
			return err
		}
	}
	return nil
}

// None is a basic provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.
type None struct{}

func (n None) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}

func (n None) GetNameservers(string) ([]*models.Nameserver, error) {
	return nil, nil
}

func (n None) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	return nil, nil
}

func init() {
	RegisterRegistrarType("NONE", func(map[string]string) (models.RegistrarDriver, error) {
		return None{}, nil
	})
}

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
