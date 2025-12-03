package providers

// This file implements the "new style" provider registration.
// The old-style required separate providers for registrar and DNS service provider.
// The new-style takes an RegisterOpts{} rather than having to update the function signature.

import (
	"log"
	"sort"
)

type RegisterOpts struct {
	// Basic info about the provider
	Name               string
	NameAliases        []string
	MaintainerGithubID string
	SupportLevel       string

	Initializer   func() (DNSServiceProvider, error)
	RecordAuditor func() (RecordAuditor, error)
	Func          any

	// Permitted fields in the creds.json file for this provider.`
	CredsFields []string

	// Permitted fields in the metadata provided to the initializer (REGISTER("name", metaMap)
	MetaFields []string

	// Capabilities: (default false)
	IsRegistrar          bool
	IsDnsServiceProvider bool
	//
	BrokeConcurrency bool
	DualHostSupport  bool

	// These files are generated:

	// DNS RecordTypes supported:
	SupportedRecordTypes  []string        // As a list
	IsRecordTypeSupported map[string]bool // As a map

}

// Info is a map of provider name to basic information about the provider.
var Info map[string]*RegisterOpts = map[string]*RegisterOpts{}

func Register(opts RegisterOpts) {
	popts := &opts

	// Register the names and aliases
	if _, exists := Info[opts.Name]; exists {
		log.Fatalf("Cannot register provider %q multiple times", opts.Name)
		Info[opts.Name] = popts
	}
	Info[opts.Name] = popts
	for _, alias := range opts.NameAliases {
		if _, exists := Info[alias]; exists {
			log.Fatalf("Cannot register provider alias %q multiple times", alias)
		}
		Info[alias] = popts
	}

	// Prep the record type support list
	sort.Strings(opts.SupportedRecordTypes)
	opts.IsRecordTypeSupported = map[string]bool{}
	for _, rt := range opts.SupportedRecordTypes {
		opts.IsRecordTypeSupported[rt] = true
	}

	// Copy data into the legacy system used for provider registration
	RegisterMaintainer(opts.Name, opts.MaintainerGithubID)

	features := convertFeatures(opts)
	providerName := opts.Name

	fns := DspFuncs{
		Initializer:   opts.Initializer,
		RecordAuditor: opts.RecordAuditor,
	}

	if opts.IsDnsServiceProvider {
		RegisterDomainServiceProviderType(providerName, fns, features)
		for _, rt := range opts.SupportedRecordTypes {
			RegisterCustomRecordType(rt, providerName, "")
		}
	}

	if opts.IsRegistrar {
		RegisterRegistrarType(providerName, ops.Initializer)
	}

}

/*
    client := providers.NewClient("credKey") // Get an API handle
  _, ok := client.(providers.Registrar) // Is this a registrar?
  _, ok := client.(providers.DNSServiceProvider) // Is this a DNS Service Provider?
  _, ok := client.(providers.ZoneLister) // Does "get-zones" work?
  _, ok := client.(providers.ZoneCreator) // Does "create-zone" work?
  rtypeMap := providers.GetSupportedRecordTypes("CLOUDFLARE") // List supported record types
  b := providers.IsRTypeSupported("CLOUDFLARE", "TXT") // Is TXT supported?
  b := providers.IsFeatureSupported("CLOUDFLARE", "feature_name") // Is feature_name supported?

Signature for the initializer (called any time an API client handle is needed):
  func newCloudflare(m map[string]string, metadata map[string]string) (providers.DNSServiceProvider, error) {}

*/
