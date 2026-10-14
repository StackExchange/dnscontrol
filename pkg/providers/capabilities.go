//go:generate go tool stringer -type=Capability

package providers

// Capability identifies an operational or documentation feature of a provider.
type Capability int

const (
	// CanAutoDNSSEC indicates that the provider can automatically handle DNSSEC,
	// so folks can ask for that.
	CanAutoDNSSEC Capability = iota

	// CanConcur indicates the provider can be used concurrently to gather zone data.
	// The default is Can(); providers that cannot be used concurrently must
	// declare Cannot(). Unimplemented() is invalid.
	CanConcur

	// CanGetZones indicates the provider can enumerate zones (get-zones ... all).
	// Reading explicitly named zones does not require this capability.
	CanGetZones

	// CanUseDSForChildren indicates the provider can handle DS record types, but
	// only for children records, not at the root of the zone.
	CanUseDSForChildren

	// DocCreateDomains means provider can add domains with the `dnscontrol create-domains` command.
	DocCreateDomains

	// CanDualHost means provider allows full management of apex NS records, so we can safely dual-host with another provider.
	CanDualHost

	// OfficiallySupported means it is actively used and maintained by Stack Exchange.
	OfficiallySupported
)

// ProviderHasCapability returns true if provider has capability.
func ProviderHasCapability(pType string, capa Capability) bool {
	if def, ok := GetDefinition(pType); ok {
		note := def.DerivedFeatures[capa]
		return note != nil && note.HasFeature
	}
	return false
}

// DocumentationNote is a way for providers to give more detail about what features they support.
type DocumentationNote struct {
	HasFeature    bool
	Unimplemented bool
	Comment       string
	Link          string
}

// DocumentationNotes is a full list of notes for a single provider.
type DocumentationNotes map[Capability]*DocumentationNote

// Can is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
func Can(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: true,
	}
	n.addStrings(comments)
	return n
}

// Cannot is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
func Cannot(comments ...string) *DocumentationNote {
	n := &DocumentationNote{
		HasFeature: false,
	}
	n.addStrings(comments)
	return n
}

// Unimplemented is a small helper for concisely creating Documentation Notes
// comments are variadic for easy omission. First is comment, second is link, the rest are ignored.
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
