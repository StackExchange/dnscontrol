package main

import (
	"sort"

	"github.com/StackExchange/dnscontrol/v3/providers"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
)

func generateFeatureMatrix() error {
	matrix := matrixData()
	markdownTable, err := markdownTable(matrix)

	if err != nil {
		return err
	}

	_ = markdownTable

	return nil
}

func markdownTable(matrix *FeatureMatrix) (string, error) {
	var tableHeaders []string
	tableHeaders = append(tableHeaders, "Provider name")
	tableHeaders = append(tableHeaders, matrix.Features...)

	var tableData [][]string
	for _, providerName := range allProviderNames() {
		featureMap := matrix.Providers[providerName]

		var tableDataRow []string
		tableDataRow = append(tableDataRow, "`"+providerName+"`")
		for _, featureName := range matrix.Features {
			tableDataRow = append(tableDataRow, featureEmoji(featureMap, featureName))
		}
		tableData = append(tableData, tableDataRow)
	}

	var markdownTable, err = markdown.NewTableFormatterBuilder().
		Build(tableHeaders...).
		Format(tableData)
	if err != nil {
		return "", err
	}

	return markdownTable, nil
}

func featureEmoji(
	featureMap FeatureMap,
	featureName string,
) string {
	if featureMap[featureName] == nil {
		return "❔"
	}

	if featureMap[featureName].HasFeature {
		return "✅"
	} else if featureMap[featureName].Unimplemented {
		return "❔"
	}
	return "❌"
}

func matrixData() *FeatureMatrix {
	const (
		OfficialSupport      = "Official Support"
		ProviderDNSProvider  = "DNS Provider"
		ProviderRegistrar    = "Registrar"
		DomainModifierAlias  = "ALIAS"
		DomainModifierDnssec = "AUTODNSSEC"
		DomainModifierCaa    = "CAA"
		DomainModifierPtr    = "PTR"
		DomainModifierNaptr  = "NAPTR"
		DomainModifierSoa    = "SOA"
		DomainModifierSrv    = "SRV"
		DomainModifierSshfp  = "SSHFP"
		DomainModifierTlsa   = "TLSA"
		DomainModifierDs     = "DS"
		DualHost             = "dual host"
		CreateDomains        = "create-domains"
		NoPurge              = "NO_PURGE"
		GetZones             = "get-zones"
	)

	matrix := &FeatureMatrix{
		Providers: map[string]FeatureMap{},
		Features: []string{
			OfficialSupport,
			ProviderDNSProvider,
			ProviderRegistrar,
			DomainModifierAlias,
			DomainModifierDnssec,
			DomainModifierCaa,
			DomainModifierPtr,
			DomainModifierNaptr,
			DomainModifierSoa,
			DomainModifierSrv,
			DomainModifierSshfp,
			DomainModifierTlsa,
			DomainModifierDs,
			DualHost,
			CreateDomains,
			NoPurge,
			GetZones,
		},
	}

	for _, providerName := range allProviderNames() {
		featureMap := FeatureMap{}
		providerNotes := providers.Notes[providerName]
		if providerNotes == nil {
			providerNotes = providers.DocumentationNotes{}
		}

		setCapability := func(
			featureName string,
			capability providers.Capability,
		) {
			if providerNotes[capability] != nil {
				featureMap[featureName] = providerNotes[capability]
				return
			}
			featureMap.SetSimple(
				featureName,
				true,
				func() bool { return providers.ProviderHasCapability(providerName, capability) },
			)
		}

		setDocumentation := func(
			featureName string,
			capability providers.Capability,
			defaultNo bool,
		) {
			if providerNotes[capability] != nil {
				featureMap[featureName] = providerNotes[capability]
			} else if defaultNo {
				featureMap[featureName] = &providers.DocumentationNote{
					HasFeature: false,
				}
			}
		}

		setDocumentation(
			OfficialSupport,
			providers.DocOfficiallySupported,
			true,
		)
		featureMap.SetSimple(
			ProviderDNSProvider,
			false,
			func() bool { return providers.DNSProviderTypes[providerName].Initializer != nil },
		)
		featureMap.SetSimple(
			ProviderRegistrar,
			false,
			func() bool { return providers.RegistrarTypes[providerName] != nil },
		)
		setCapability(
			DomainModifierAlias,
			providers.CanUseAlias,
		)
		setCapability(
			DomainModifierDnssec,
			providers.CanAutoDNSSEC,
		)
		setCapability(
			DomainModifierCaa,
			providers.CanUseCAA,
		)
		setCapability(
			DomainModifierDs,
			providers.CanUseDS,
		)
		setCapability(
			DomainModifierNaptr,
			providers.CanUseNAPTR,
		)
		setCapability(
			DomainModifierPtr,
			providers.CanUsePTR,
		)
		setCapability(
			DomainModifierSoa,
			providers.CanUseSOA,
		)
		setCapability(
			DomainModifierSrv,
			providers.CanUseSRV,
		)
		setCapability(
			DomainModifierSshfp,
			providers.CanUseSSHFP,
		)
		setCapability(
			DomainModifierTlsa,
			providers.CanUseTLSA,
		)
		setCapability(
			GetZones,
			providers.CanGetZones,
		)
		setDocumentation(
			CreateDomains,
			providers.DocCreateDomains,
			true,
		)
		setDocumentation(
			DualHost,
			providers.DocDualHost,
			false,
		)

		// no purge is a freaky double negative
		cantUseNOPURGE := providers.CantUseNOPURGE
		if providerNotes[cantUseNOPURGE] != nil {
			featureMap[NoPurge] = providerNotes[cantUseNOPURGE]
		} else {
			featureMap.SetSimple(
				NoPurge,
				false,
				func() bool { return !providers.ProviderHasCapability(providerName, cantUseNOPURGE) },
			)
		}
		matrix.Providers[providerName] = featureMap
	}
	return matrix
}

func allProviderNames() []string {
	const ProviderNameNone = "NONE"

	allProviderNames := map[string]bool{}
	for providerName := range providers.RegistrarTypes {
		if providerName == ProviderNameNone {
			continue
		}
		allProviderNames[providerName] = true
	}
	for providerName := range providers.DNSProviderTypes {
		if providerName == ProviderNameNone {
			continue
		}
		allProviderNames[providerName] = true
	}

	var allProviderNamesAsString []string
	for providerName := range allProviderNames {
		allProviderNamesAsString = append(allProviderNamesAsString, providerName)
	}
	sort.Strings(allProviderNamesAsString)

	return allProviderNamesAsString
}

// FeatureMap maps provider names to compliance documentation.
type FeatureMap map[string]*providers.DocumentationNote

// SetSimple configures a provider's setting in featureMap.
func (featureMap FeatureMap) SetSimple(
	name string,
	unknownsAllowed bool,
	f func() bool,
) {
	if f() {
		featureMap[name] = &providers.DocumentationNote{HasFeature: true}
	} else if !unknownsAllowed {
		featureMap[name] = &providers.DocumentationNote{HasFeature: false}
	}
}

// FeatureMatrix describes features and which providers support it.
type FeatureMatrix struct {
	Features  []string
	Providers map[string]FeatureMap
}
