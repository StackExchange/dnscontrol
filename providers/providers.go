package providers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/models"
)

var RegistrarTypes = map[string]models.RegistrarInitializer{}

var DNSProviderTypes = map[string]models.DspInitializer{}

//RegisterRegistrarType adds a registrar type to the registry by providing a suitable initialization function.
func RegisterRegistrarType(name string, init models.RegistrarInitializer, pm ...ProviderMetadata) {
	if _, ok := RegistrarTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	RegistrarTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

//RegisterDomainServiceProviderType adds a dsp to the registry with the given initialization function.
func RegisterDomainServiceProviderType(name string, init models.DspInitializer, pm ...ProviderMetadata) {
	if _, ok := DNSProviderTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	DNSProviderTypes[name] = init
	unwrapProviderCapabilities(name, pm)
}

func createRegistrar(rType string, config map[string]string) (models.Registrar, error) {
	initer, ok := RegistrarTypes[rType]
	if !ok {
		return nil, fmt.Errorf("Registrar type %s not declared.", rType)
	}
	return initer(config)
}

func CreateDNSProvider(dType string, config map[string]string, meta json.RawMessage) (models.DNSServiceProvider, error) {
	initer, ok := DNSProviderTypes[dType]
	if !ok {
		return nil, fmt.Errorf("DSP type %s not declared", dType)
	}
	return initer(config, meta)
}

//CreateRegistrars will load all registrars from the dns config, and create instances of the correct type using data from
//the provider config to load relevant keys and options.
func CreateRegistrars(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]models.Registrar, error) {
	regs := map[string]models.Registrar{}
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

func CreateDsps(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]models.DNSServiceProvider, error) {
	dsps := map[string]models.DNSServiceProvider{}
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
	RegisterRegistrarType("NONE", func(map[string]string) (models.Registrar, error) {
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
