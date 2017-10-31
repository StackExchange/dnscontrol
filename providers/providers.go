package providers

import (
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

// CreateProviders will initialize all dns providers and registrars, and store them in all domains as needed.
func CreateProviders(cfg *models.DNSConfig, providerConfigs map[string]map[string]string) error {
	dnsProvidersByName := map[string]models.DNSServiceProviderDriver{}
	registrarsByName := map[string]models.RegistrarDriver{}
	// create dns providers
	for _, dnsProvider := range cfg.DNSProviders {
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
	for _, reg := range cfg.Registrars {
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
	// everything that does not explicitly include "_exclude_from_defaults: true" in creds.json is default
	isDefault := func(name string) bool {
		if vals, ok := providerConfigs[name]; ok {
			if exclude := vals["_exclude_from_defaults"]; exclude == "true" {
				return false
			}
		}
		return true
	}
	for _, d := range cfg.Domains {
		if reg, ok := registrarsByName[d.RegistrarName]; ok {
			d.Registrar = &registrar{
				RegistrarDriver: reg,
				name:            d.RegistrarName,
				isDefault:       isDefault(d.RegistrarName),
			}
		} else {
			return fmt.Errorf("Registrar '%s' for %s not defined", d.RegistrarName, d.Name)
		}
		for dspName, num := range d.DNSProviderNames {
			if prov, ok := dnsProvidersByName[dspName]; ok {
				d.DNSProviders = append(d.DNSProviders, &dnsProvider{
					DNSServiceProviderDriver: prov,
					name:           dspName,
					isDefault:      isDefault(dspName),
					numNameservers: num,
				})
			} else {
				return fmt.Errorf("DNS Provider '%s' for %s not defined", dspName, d.Name)
			}
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

type registrar struct {
	models.RegistrarDriver
	name      string
	isDefault bool
}

func (r *registrar) Name() string       { return r.name }
func (r *registrar) RunByDefault() bool { return r.isDefault }

type dnsProvider struct {
	models.DNSServiceProviderDriver
	name           string
	isDefault      bool
	numNameservers int
}

func (r *dnsProvider) Name() string                  { return r.name }
func (r *dnsProvider) RunByDefault() bool            { return r.isDefault }
func (r *dnsProvider) NumberOfNameserversToUse() int { return r.numNameservers }
