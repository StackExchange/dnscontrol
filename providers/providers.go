package providers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/models"
)

//Registrar is an interface for a domain registrar. It can return a list of needed corrections to be applied in the future.
type Registrar interface {
	GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error)
}

//DNSServiceProvider is able to generate a set of corrections that need to be made to correct records for a domain
type DNSServiceProvider interface {
	GetNameservers(domain string) ([]*models.Nameserver, error)
	GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error)
}

//DomainCreator should be implemented by providers that have the ability to add domains to an account. the create-domains command
//can be run to ensure all domains are present before running preview/push
type DomainCreator interface {
	EnsureDomainExists(domain string) error
}

//RegistrarInitializer is a function to create a registrar. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type RegistrarInitializer func(map[string]string) (Registrar, error)

var registrarTypes = map[string]RegistrarInitializer{}

//DspInitializer is a function to create a registrar. Function will be passed the unprocessed json payload from the configuration file for the given provider.
type DspInitializer func(map[string]string, json.RawMessage) (DNSServiceProvider, error)

var dspTypes = map[string]DspInitializer{}

//RegisterRegistrarType adds a registrar type to the registry by providing a suitable initialization function.
func RegisterRegistrarType(name string, init RegistrarInitializer) {
	if _, ok := registrarTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	registrarTypes[name] = init
}

//RegisterDomainServiceProviderType adds a dsp to the registry with the given initialization function.
func RegisterDomainServiceProviderType(name string, init DspInitializer) {
	if _, ok := dspTypes[name]; ok {
		log.Fatalf("Cannot register registrar type %s multiple times", name)
	}
	dspTypes[name] = init
}

func createRegistrar(rType string, config map[string]string) (Registrar, error) {
	initer, ok := registrarTypes[rType]
	if !ok {
		return nil, fmt.Errorf("Registrar type %s not declared.", rType)
	}
	return initer(config)
}

func CreateDNSProvider(dType string, config map[string]string, meta json.RawMessage) (DNSServiceProvider, error) {
	initer, ok := dspTypes[dType]
	if !ok {
		return nil, fmt.Errorf("DSP type %s not declared", dType)
	}
	return initer(config, meta)
}

//CreateRegistrars will load all registrars from the dns config, and create instances of the correct type using data from
//the provider config to load relevant keys and options.
func CreateRegistrars(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]Registrar, error) {
	regs := map[string]Registrar{}
	for _, reg := range d.Registrars {
		rawMsg, ok := providerConfigs[reg.Name]
		if !ok && reg.Type != "NONE" {
			return nil, fmt.Errorf("Registrar %s not listed in -providers file.", reg.Name)
		}
		registrar, err := createRegistrar(reg.Type, rawMsg)
		if err != nil {
			return nil, err
		}
		regs[reg.Name] = registrar
	}
	return regs, nil
}

func CreateDsps(d *models.DNSConfig, providerConfigs map[string]map[string]string) (map[string]DNSServiceProvider, error) {
	dsps := map[string]DNSServiceProvider{}
	for _, dsp := range d.DNSProviders {
		//log.Printf("dsp.Name=%#v\n", dsp.Name)
		rawMsg, ok := providerConfigs[dsp.Name]
		if !ok {
			return nil, fmt.Errorf("DNSServiceProvider %s not listed in -providers file", dsp.Name)
		}
		provider, err := CreateDNSProvider(dsp.Type, rawMsg, dsp.Metadata)
		if err != nil {
			return nil, err
		}
		dsps[dsp.Name] = provider
	}
	return dsps, nil
}

// None is a basivc provider type that does absolutely nothing. Can be useful as a placeholder for third parties or unimplemented providers.
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
	RegisterRegistrarType("NONE", func(map[string]string) (Registrar, error) {
		return None{}, nil
	})
}
