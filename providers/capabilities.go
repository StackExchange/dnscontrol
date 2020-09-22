//go:generate stringer -type=Capability

package providers

import (
	"log"
)

// Capability is a bitmasked set of "features" that a provider supports. Only use constants from this package.
type Capability uint32

const (
	// If you add something to this list, you probably want to add it to pkg/normalize/validate.go checkProviderCapabilities() or somewhere near there.

	// CanUseAlias indicates the provider support ALIAS records (or flattened CNAMES). Up to the provider to translate them to the appropriate record type.
	CanUseAlias Capability = iota

	// CanUseCAA indicates the provider can handle CAA records
	CanUseCAA

	// CanUseDS indicates that the provider can handle DS record types. This
	// implies CanUseDSForChildren without specifying the latter explicitly.
	CanUseDS

	// CanUseDSForChildren indicates the provider can handle DS record types, but
	// only for children records, not at the root of the zone.
	CanUseDSForChildren

	// CanUsePTR indicates the provider can handle PTR records
	CanUsePTR

	// CanUseNAPTR indicates the provider can handle NAPTR records
	CanUseNAPTR

	// CanUseSRV indicates the provider can handle SRV records
	CanUseSRV

	// CanUseSSHFP indicates the provider can handle SSHFP records
	CanUseSSHFP

	// CanUseTLSA indicates the provider can handle TLSA records
	CanUseTLSA

	// CanUseTXTMulti indicates the provider can handle TXT records with multiple strings
	CanUseTXTMulti

	// CanAutoDNSSEC indicates that the provider can automatically handle DNSSEC,
	// so folks can ask for that.
	CanAutoDNSSEC

	// CantUseNOPURGE indicates NO_PURGE is broken for this provider. To make it
	// work would require complex emulation of an incremental update mechanism,
	// so it is easier to simply mark this feature as not working for this
	// provider.
	CantUseNOPURGE

	// DocOfficiallySupported means it is actively used and maintained by stack exchange
	DocOfficiallySupported
	// DocDualHost means provider allows full management of apex NS records, so we can safely dual-host with anothe provider
	DocDualHost
	// DocCreateDomains means provider can add domains with the `dnscontrol create-domains` command
	DocCreateDomains

	// CanUseRoute53Alias indicates the provider support the specific R53_ALIAS records that only the Route53 provider supports
	CanUseRoute53Alias

	// CanGetZones indicates the provider supports the get-zones subcommand.
	CanGetZones

	// CanUseAzureAlias indicates the provider support the specific Azure_ALIAS records that only the Azure provider supports
	CanUseAzureAlias
)

var providerCapabilities = map[string]map[Capability]bool{}

// ProviderHasCapability returns true if provider has capability.
func ProviderHasCapability(pType string, cap Capability) bool {
	if providerCapabilities[pType] == nil {
		return false
	}
	return providerCapabilities[pType][cap]
}

// DocumentationNote is a way for providers to give more detail about what features they support.
type DocumentationNote struct {
	HasFeature    bool
	Unimplemented bool
	Comment       string
	Link          string
}

// DocumentationNotes is a full list of notes for a single provider
type DocumentationNotes map[Capability]*DocumentationNote

// ProviderMetadata is a common interface for DocumentationNotes and Capability to be used interchangeably
type ProviderMetadata interface{}

// Notes is a collection of all documentation notes, keyed by provider type
var Notes = map[string]DocumentationNotes{}

func unwrapProviderCapabilities(pName string, meta []ProviderMetadata) {
	if providerCapabilities[pName] == nil {
		providerCapabilities[pName] = map[Capability]bool{}
	}
	for _, pm := range meta {
		switch x := pm.(type) {
		case Capability:
			providerCapabilities[pName][x] = true
		case DocumentationNotes:
			if Notes[pName] == nil {
				Notes[pName] = DocumentationNotes{}
			}
			for k, v := range x {
				Notes[pName][k] = v
				providerCapabilities[pName][k] = v.HasFeature
			}
		default:
			log.Fatalf("Unrecognized ProviderMetadata type: %T", pm)
		}

	}
}

// Can is a small helper for concisely creating Documentation Notes
// comments are variadic for easy ommission. First is comment, second is link, the rest are ignored.
func Can(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: true,
	}
	n.addStrings(comments)
	return n
}

// Cannot is a small helper for concisely creating Documentation Notes
// comments are variadic for easy ommission. First is comment, second is link, the rest are ignored.
func Cannot(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: false,
	}
	n.addStrings(comments)
	return n
}

// Unimplemented is a small helper for concisely creating Documentation Notes
// comments are variadic for easy ommission. First is comment, second is link, the rest are ignored.
func Unimplemented(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature:    false,
		Unimplemented: true,
	}
	n.addStrings(comments)
	return n
}

func (n *DocumentationNote) addStrings(comments []string) {
	if len(comments) > 0 {
		n.Comment = comments[0]
	}
	if len(comments) > 1 {
		n.Link = comments[1]
	}
}
