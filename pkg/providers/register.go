package providers

import (
	"log"
	"maps"
	"slices"
	"sort"
	"strings"
)

type ProviderHandle = interface{}

type RegisterOpts = struct {
	// Information about the provider.
	Name               string
	NameAliases        []string
	MaintainerGithubID string
	SupportLevel       SupportLevel
	ProviderHandle     ProviderHandle

	// Legacy functions: (eventually these will be methods of the ProviderHandle)
	RegistrarInitializer          RegistrarInitializer
	DNSServiceProviderInitializer DspInitializer
	RecordAuditor                 RecordAuditor

	// CredsFields is what fields to expect in creds.json for this provider.
	CredsFields []string
	// foo             Pave "foo" as a string
	// foo:string      Pave "foo" as a string
	// foo:bool        Pave "foo" as a bool
	// Access the values as Provider.Creds["foo"] (.Creds is a map[string]any)

	// MetadataFields is what fields to expect in metadata for this provider.
	// REGISTRAR("credkey", { metafield: "metavalue" }),
	// or
	// NewDnsProvider("credkey", { metafield: "metavalue" }, ...
	MetadataFields []string
	// foo             Pave "foo" as a string
	// foo:string      Pave "foo" as a string
	// foo:bool        Pave "foo" as a bool
	// Access these as Provider.Metadata["foo"] (.Metadata is a map[string]any)

	// DNS RecordTypes supported:
	RecordTypes []string
	// TYPE                 TYPE is fully supported
	// TYPE:note:foo        TYPE is supported, but document "foo" as a note.
	// TYPE:unimplemented   TYPE is supported by the provider, but DNSControl doesn't have code to support it.
	// TYPE:unimplemented:foo  TYPE is supported by the provider, but DNSControl doesn't have code to support it. Document "foo" as a note instead of the default message.
	// If this is left empty, all record types are assumed to be supported. (useful for NONE, BIND and a future "BACKUP" provider).
	// Access this info as:
	//     Provider.IsRecordTypeSupported["TYPE"] returns DocumentationNote{}
	//     Provider.IsRecordTypeSupported["TYPE"].HasFeature is true if TYPE is supported.
	//     Provider.SupportedRTypes returns []string of supported types (unimplemented and unsupported types are excluded)

	Features DocumentationNotes // As input at Register() time.  Read from HasFeature instead.

	//////////////////////////////////////////////////////////////////////////
	// These are computed during registration; do not set them.

	// Map of supported record types.
	IsRecordTypeSupported map[string]DocumentationNote
	// Sorted list of supported record types (excludes unimplemented and unsupported types).
	SupportedRecordTypes []string // Being on this map means it is supported.

	// Features supported by this provider.
	HasFeature map[Capability]DocumentationNote // Computed map of features supported by this provider.
	// NB(tlim): Keep this has "pure" features, not whether or not a record type
	// is supported. That's what IsRecTypeSupported is for.

	// Maps field names to their types.
	CredsSchema map[string]FieldType
	// Maps field names to their types.
	MetadataSchema map[string]FieldType
	// Maps record types to their documentation notes. .IsRecordTypeSupported["A"].HasFeature tells if "A" is supported.

}

// Info stores registration information for all providers.  Access providers.Info["CLOUDFLARE"] to get the registration info for Cloudflare.
var Info = map[string]*RegisterOpts{}

func Register(opts RegisterOpts) {

	// Register the options under all the names/aliases:
	for _, providerName := range append([]string{opts.Name}, opts.NameAliases...) {
		if _, ok := Info[providerName]; ok {
			log.Fatalf("Cannot register provider %q multiple times", providerName)
		}
		Info[providerName] = &opts
	}

	opts.CredsSchema = parseFieldTypeSpec(opts.CredsFields)
	opts.MetadataSchema = parseFieldTypeSpec(opts.MetadataFields)
	opts.IsRecordTypeSupported, opts.SupportedRecordTypes = parseRecordTypes(opts.RecordTypes)

	// Populate the Features map:

	// Features we determine automatically based on implemented interfaces.
	handle := opts.ProviderHandle
	opts.HasFeature = map[Capability]DocumentationNote{}
	if _, ok := handle.(ZoneLister); ok {
		opts.HasFeature[CanGetZones] = DocumentationNote{HasFeature: true}
	}
	if _, ok := handle.(Registrar); ok {
		opts.HasFeature[IsRegistrar] = DocumentationNote{HasFeature: true}
	}
	if _, ok := handle.(ZoneCreator); ok {
		opts.HasFeature[DocCreateDomains] = DocumentationNote{HasFeature: true}
	}
	if _, ok := handle.(DNSServiceProvider); ok {
		opts.HasFeature[IsDnsServiceProvider] = DocumentationNote{HasFeature: true}
	}
	if opts.SupportLevel == SupportLevelOfficial {
		opts.HasFeature[DocOfficiallySupported] = DocumentationNote{HasFeature: true}
	}
	//fmt.Printf("DEBUG: opts.HasFeature[DocCreateDomains] = %+v\n", opts.HasFeature[DocCreateDomains])
	//fmt.Printf("DEBUG: opts.HasFeature[CanGetZones] = %+v\n", opts.HasFeature[CanGetZones])
	//fmt.Printf("DEBUG: opts.HasFeature[IsDnsServiceProvider] = %+v\n", opts.HasFeature[IsDnsServiceProvider])

	//
	// Populate legacy fields for backward compatibility.
	//

	// Warn if the developer has specified any legacy features that are now automatic.
	checkForLegacyFeatures(opts.Features)

	// Populate .Features if needed.
	opts.Features = createFeaturesForRecordTypes(opts.Features, opts.IsRecordTypeSupported)
	opts.Features = createFeaturesForOther(opts.Features, opts.HasFeature)

	// Am I a registrar?
	if opts.RegistrarInitializer != nil || opts.HasFeature[IsRegistrar].HasFeature {
		RegisterRegistrarType(opts.Name, opts.RegistrarInitializer)
		//RegistrarTypes[opts.Name] = opts.RegistrarInitializer
	}

	// Am I a DNS service provider?
	if opts.DNSServiceProviderInitializer != nil || opts.HasFeature[IsDnsServiceProvider].HasFeature {
		RegisterDomainServiceProviderType(opts.Name, DspFuncs{
			Initializer:   opts.DNSServiceProviderInitializer,
			RecordAuditor: opts.RecordAuditor,
		}, opts.Features)
		//DNSProviderTypes[opts.Name] = DspFuncs{Initializer: opts.DNSServiceProviderInitializer, RecordAuditor: opts.RecordAuditor}
		//fmt.Printf("DEBUG: Registered DSP %q\n", opts.Name)
		//for i, j := range opts.Features {
		//	fmt.Printf("DEBUG:    Feature: %q: %+v\n", i, j)
		//}
		//fmt.Printf("DEBUG: end\n")
	}
	//fmt.Printf("DEBUG: Provider %q registered with features:\n", opts.Name)
	//fmt.Printf("DEBUG:       recTypes = %v\n", opts.IsRecordTypeSupported)
	//fmt.Printf("DEBUG:       others   = %v\n", opts.Features)
	unwrapProviderCapabilities(opts.Name, []ProviderMetadata{
		opts.Features, // The non-recordtype features.
		//slices.Collect(maps.Values(opts.IsRecordTypeSupported)), // The Record Types supported.
		//opts.IsRecordTypeSupported, // The Record Types supported.
		convertToDocumentationNotes(opts.IsRecordTypeSupported), // The Record Types supported.
	})

	RegisterMaintainer(opts.Name, opts.MaintainerGithubID)

	//fmt.Printf("DEBUG: providerCapabilities[%q] %v\n", opts.Name, providerCapabilities[opts.Name])
	//fmt.Printf("DEBUG: providerCapabilities[%q][CanUseSMIME] %v\n", opts.Name, providerCapabilities[opts.Name][CanUseSMIMEA])
	//fmt.Printf("DEBUG: Notes[%q] %v\n", opts.Name, Notes[opts.Name])
	//fmt.Printf("DEBUG: Notes[%q][CanUseSMIME] %v\n", opts.Name, Notes[opts.Name][CanUseSMIMEA])

}

var ValidTypes = []string{}
var IsValidType = map[string]struct{}{}

// PostInitAllProviders performs any initialization that has to be done after
// all providers have registered themselves. For example, collecting all the
// record types supported by any provider.
func PostInitAllProviders() {
	typeMap := map[string]DocumentationNote{}

	// 	DNSProviderTypes[name] = fns
	// RegisterTypes[name]

	// Gather the names of all record types supported by any provider.
	for _, opts := range Info {
		for _, rtypeName := range opts.SupportedRecordTypes {
			typeMap[rtypeName] = DocumentationNote{HasFeature: true}
		}
	}
	typeList := slices.Sorted(maps.Keys(typeMap))

	// Save the data for global access:
	ValidTypes = typeList
	for _, rt := range typeList {
		IsValidType[rt] = struct{}{}
	}

	// Any providers that support all record types? Populate their record type support info.
	//fmt.Printf("DEBUG: Provider Info %v\n", Info)
	for providerName, opts := range Info {
		//fmt.Printf("DEBUG: Provider %q\n", providerName)
		// Find any provider that has an empty .SupportedRecordTypes list; that means it supports all types.
		if len(opts.RecordTypes) == 0 {
			//fmt.Printf("DEBUG:  EMPTY %q\n", providerName)
			// Populate the record-type support info:
			opts.SupportedRecordTypes = typeList
			opts.IsRecordTypeSupported = typeMap
			opts.RecordTypes = typeList // Pretend the user specified all types.
			// Repopulate the features for the provider:
			unwrapProviderCapabilities(providerName, []ProviderMetadata{
				opts.Features, // The non-recordtype features.
				convertToDocumentationNotes(opts.IsRecordTypeSupported), // The Record Types supported.
			})
		}
	}

	// DEBUG
	//fmt.Printf("DEBUG: ValidTypes: %#v\n", ValidTypes)
	//fmt.Printf("DEBUG: IsValidType: %#v\n", IsValidType)
}

func convertToDocumentationNotes(in map[string]DocumentationNote) DocumentationNotes {
	ret := DocumentationNotes{}
	for k, v := range in {
		ret[nameToCap[k]] = &v
	}
	return ret
}

// parseFieldTypeSpec parses field type specifications into a map of field names to types. Field types supported are:
//
//	typename           (Defaults to string)
//	typename:string
//	typename:bool
func parseFieldTypeSpec(fields []string) map[string]FieldType {
	m := map[string]FieldType{}
	for i, f := range fields {
		fl := strings.SplitN(f, ":", 2)
		name := fl[0]
		if _, exists := m[name]; exists {
			log.Fatalf("item %d is invalid: Duplicate field name %q", i, name)
		}
		if len(fl) == 1 {
			m[name] = FieldTypeString
		} else {
			switch fl[1] {
			case "string":
				m[name] = FieldTypeString
			case "bool":
				m[name] = FieldTypeBool
			default:
				log.Fatalf("item %d is invalid: Unknown field type %q in field %q", i, fl[1], f)
			}
		}
	}
	return m
}

// parseRecordTypeSpec parses record type specifications into a map of record
// types to DocumentationNote, and a sorted slice of supported record types.
// Each element of rtypes can be in one of the following forms:
//
// TYPE                       TYPE is fully supported
// TYPE:note:mynote           TYPE is supported, but document "mynote" as a note. All fields after "note:" are part of the note.
// TYPE:unimplemented         TYPE is supported by the provider, but DNSControl doesn't have code to support it.
// TYPE:unimplemented:mynote  TYPE is supported by the provider, but DNSControl doesn't have code to support it, with a note for the documentation.
//
// If no note is provided for an unimplemented type, a default note is used.
// If the note starts with a URL, it will not be part of the note but will be rendered as a link in the documentation.
// The end of the URL is determined by the first whitespace character after "http://" or "https://".
func parseRecordTypes(rtypes []string) (map[string]DocumentationNote, []string) {
	m := map[string]DocumentationNote{}
	supported := []string{}
	for i, item := range rtypes {
		item = strings.TrimSpace(item)
		if item == "" {
			log.Fatalf("item %d is invalid: Empty field in schema: %#v", i, rtypes)
		}

		name := ""
		verb := ""
		note := ""
		url := ""
		unimplemented := false

		parts := strings.SplitN(item, ":", 3)
		switch len(parts) {
		case 0:
			log.Fatalf("item %d is invalid: Empty field in schema: %#v", i, rtypes)
		case 3:
			note = strings.TrimSpace(parts[2])
			fallthrough
		case 2:
			verb = strings.TrimSpace(parts[1])
			fallthrough
		case 1:
			name = parts[0]
		}
		note = strings.TrimSpace(note)

		if name != strings.ToUpper(name) {
			log.Fatalf("item %d is invalid: Record type %q must be uppercase", i, name)
		}

		if _, exists := m[name]; exists {
			log.Fatalf("item %d is invalid: Duplicate record type %q", i, name)
		}

		switch verb {
		case "":
			// fully supported
		case "note":
			// supported with note
		case "unimplemented":
			unimplemented = true
		default:
			log.Fatalf("item %d is invalid: Invalid specification %q", i, item)
		}
		note, url = extractURLFromNote(note)
		if unimplemented && note == "" {
			note = "This record type is supported by the provider, but DNSControl does not yet have code to manage it."
		}

		m[name] = DocumentationNote{
			HasFeature:    true,
			Comment:       note,
			Unimplemented: unimplemented,
			Link:          url,
		}
		if !unimplemented {
			supported = append(supported, name)
		}
	}
	sort.Strings(supported)
	return m, supported
}

func extractURLFromNote(note string) (string, string) {
	note = strings.TrimSpace(note)
	if strings.HasPrefix(note, "http://") || strings.HasPrefix(note, "https://") {
		before, after, found := strings.Cut(note, " ")
		if found {
			return strings.TrimSpace(after), before
		}
		return before, before
	}
	return note, ""
}

func checkForLegacyFeatures(features DocumentationNotes) {
	for cap := range features {
		switch cap {
		case CanGetZones, DocCreateDomains, DocOfficiallySupported,
			IsRegistrar, IsDnsServiceProvider:
			log.Printf("Warning: Capability %s is now set automatically based on implemented interfaces; please remove it from the Features list.", cap)
		case CanUseAKAMAICDN, CanUseAKAMAITLC, CanUseAlias, CanUseAzureAlias,
			CanUseCAA, CanUseDHCID, CanUseDNAME, CanUseDS, CanUseHTTPS, CanUseLOC,
			CanUseNAPTR, CanUsePTR, CanUseRoute53Alias, CanUseRP, CanUseSMIMEA,
			CanUseSOA, CanUseSRV, CanUseSSHFP, CanUseSVCB, CanUseTLSA, CanUseDNSKEY,
			CanUseOPENPGPKEY:
			log.Printf("Warning: Capability %s is now set automatically. Please remove it from the Features list.", cap)
		case CanAutoDNSSEC, CanConcur, CanOnlyDiff1Features, CanUseDSForChildren, DocDualHost:
			// User has set this manually, as required. No action needed.
		default:
			// Unknown feature! That's a bug.
			log.Fatalf("Warning: Capability %s is not implemented in checkForLegacyFeatures.  This should not happen.", cap)
		}
	}
}

// nameToCap maps record type names to their corresponding Capability constants.
// This map should be treated as read-only after initialization.
var nameToCap = map[string]Capability{
	"AKAMAICDN":         CanUseAKAMAICDN,
	"AKAMAITLC":         CanUseAKAMAITLC,
	"ALIAS":             CanUseAlias,
	"AZURE_ALIAS_A":     CanUseAzureAlias,
	"AZURE_ALIAS_AAAA":  CanUseAzureAlias,
	"AZURE_ALIAS_CNAME": CanUseAzureAlias,
	"CAA":               CanUseCAA,
	"DHCID":             CanUseDHCID,
	"DNAME":             CanUseDNAME,
	"DS":                CanUseDS,
	"HTTPS":             CanUseHTTPS,
	"LOC":               CanUseLOC,
	"NAPTR":             CanUseNAPTR,
	"PTR":               CanUsePTR,
	"R53_ALIAS":         CanUseRoute53Alias,
	"RP":                CanUseRP,
	"SMIMEA":            CanUseSMIMEA,
	"SOA":               CanUseSOA,
	"SRV":               CanUseSRV,
	"SSHFP":             CanUseSSHFP,
	"SVCB":              CanUseSVCB,
	"TLSA":              CanUseTLSA,
	"DNSKEY":            CanUseDNSKEY,
	"OPENPGPKEY":        CanUseOPENPGPKEY,
}

// Populate the legacy Features map based on supported record types.
func createFeaturesForRecordTypes(features DocumentationNotes, rtypeInfo map[string]DocumentationNote) DocumentationNotes {
	for rtype, doc := range rtypeInfo {
		if cap, ok := nameToCap[rtype]; ok {
			features[cap] = &doc
		}
	}
	return features
}

func createFeaturesForOther(features DocumentationNotes, hasFeature map[Capability]DocumentationNote) DocumentationNotes {
	for cap, doc := range hasFeature {
		//fmt.Printf("DEBUG: createFeaturesForOther: cap=%q doc=%+v\n", cap, doc)
		features[cap] = &doc
	}
	return features
}
