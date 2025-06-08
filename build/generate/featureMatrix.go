package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/providers"
	_ "github.com/StackExchange/dnscontrol/v4/providers/_all"
	"github.com/fbiville/markdown-table-formatter/pkg/markdown"
)

func generateFeatureMatrix() error {
	var replacementContent string = ""
	matrix := matrixData()

	for i := 0; i < len(matrix.FeatureTables); i++ {
		var tableTitle = matrix.FeatureTablesTitles[i]
		replacementContent += fmt.Sprintf("\n### %s <!--(table %d/%d)-->\n\n",
			tableTitle, i+1, len(matrix.FeatureTables))
		markdownTable, err := markdownTable(matrix, int32(i))
		if err != nil {
			return err
		}
		replacementContent += markdownTable
		replacementContent += "\n"
	}

	replaceInlineContent(
		"documentation/provider/index.md",
		"<!-- provider-matrix-start -->",
		"<!-- provider-matrix-end -->",
		replacementContent,
	)

	return nil
}

func markdownTable(matrix *FeatureMatrix, tableNumber int32) (string, error) {
	var tableHeaders []string
	tableHeaders = append(tableHeaders, "Provider name")
	tableHeaders = append(tableHeaders, matrix.FeatureTables[tableNumber]...)

	var tableData [][]string
	for _, providerName := range allProviderNames() {
		featureMap := matrix.Providers[providerName]

		var tableDataRow []string
		tableDataRow = append(tableDataRow, "[`"+providerName+"`]("+strings.ToLower(providerName)+".md)")
		for _, featureName := range matrix.FeatureTables[tableNumber] {
			tableDataRow = append(tableDataRow, featureEmoji(featureMap, featureName))
		}
		skipThisRow := true
		for status := range tableDataRow[1:] {
			if tableDataRow[status+1] != "❔" {
				skipThisRow = false
			}
		}
		if !skipThisRow {
			tableData = append(tableData, tableDataRow)
		}
	}

	markdownTable, err := markdown.NewTableFormatterBuilder().
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
		OfficialSupport      = "Official Support" // vs. community supported
		ProviderDNSProvider  = "DNS Provider"
		ProviderRegistrar    = "Registrar"
		ProviderThreadSafe   = "[Concurrency Verified](../advanced-feaures/concurrency-verified.md)"
		DomainModifierAlias  = "[`ALIAS`](../language-reference/domain-modifiers/ALIAS.md)"
		DomainModifierCaa    = "[`CAA`](../language-reference/domain-modifiers/CAA.md)"
		DomainModifierDnssec = "[`AUTODNSSEC`](../language-reference/domain-modifiers/AUTODNSSEC_ON.md)"
		DomainModifierHTTPS  = "[`HTTPS`](../language-reference/domain-modifiers/HTTPS.md)"
		DomainModifierLoc    = "[`LOC`](../language-reference/domain-modifiers/LOC.md)"
		DomainModifierNaptr  = "[`NAPTR`](../language-reference/domain-modifiers/NAPTR.md)"
		DomainModifierPtr    = "[`PTR`](../language-reference/domain-modifiers/PTR.md)"
		DomainModifierSoa    = "[`SOA`](../language-reference/domain-modifiers/SOA.md)"
		DomainModifierSrv    = "[`SRV`](../language-reference/domain-modifiers/SRV.md)"
		DomainModifierSshfp  = "[`SSHFP`](../language-reference/domain-modifiers/SSHFP.md)"
		DomainModifierSvcb   = "[`SVCB`](../language-reference/domain-modifiers/SVCB.md)"
		DomainModifierTlsa   = "[`TLSA`](../language-reference/domain-modifiers/TLSA.md)"
		DomainModifierDs     = "[`DS`](../language-reference/domain-modifiers/DS.md)"
		DomainModifierDhcid  = "[`DHCID`](../language-reference/domain-modifiers/DHCID.md)"
		DomainModifierDname  = "[`DNAME`](../language-reference/domain-modifiers/DNAME.md)"
		DomainModifierDnskey = "[`DNSKEY`](../language-reference/domain-modifiers/DNSKEY.md)"
		DualHost             = "[dual host](../advanced-features/dual-host.md)"
		CreateDomains        = "create-domains"
		GetZones             = "get-zones"
	)

	matrix := &FeatureMatrix{
		Providers: map[string]FeatureMap{},
		FeatureTablesTitles: []string{
			"Provider Type",
			"Provider API",
			"DNS extensions",
			"Service discovery",
			"Security",
			"DNSSEC",
		},
		FeatureTables: [][]string{
			[]string{ // provider type
				OfficialSupport,
				ProviderDNSProvider,
				ProviderRegistrar,
			},
			[]string{ // provider API
				ProviderThreadSafe,
				DualHost,
				CreateDomains,
				// NoPurge,
				GetZones,
			},
			[]string{ // DNS extensions
				DomainModifierAlias,
				DomainModifierDname,
				DomainModifierLoc,
				DomainModifierPtr,
				DomainModifierSoa,
			},
			[]string{ // service discovery
				DomainModifierDhcid,
				DomainModifierNaptr,
				DomainModifierSrv,
				DomainModifierSvcb,
			},
			[]string{ // security
				DomainModifierCaa,
				DomainModifierHTTPS,
				DomainModifierSshfp,
				DomainModifierTlsa,
			},
			[]string{ // dnssec
				DomainModifierDnssec,
				DomainModifierDnskey,
				DomainModifierDs,
			},
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
			ProviderThreadSafe,
			providers.CanConcur,
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
			DomainModifierDhcid,
			providers.CanUseDHCID,
		)
		setCapability(
			DomainModifierDname,
			providers.CanUseDNAME,
		)
		setCapability(
			DomainModifierDs,
			providers.CanUseDS,
		)
		setCapability(
			DomainModifierDnskey,
			providers.CanUseDNSKEY,
		)
		setCapability(
			DomainModifierHTTPS,
			providers.CanUseHTTPS,
		)
		setCapability(
			DomainModifierLoc,
			providers.CanUseLOC,
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
			DomainModifierSvcb,
			providers.CanUseSVCB,
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

		//		// no purge is a freaky double negative
		//		cantUseNOPURGE := providers.CantUseNOPURGE
		//		if providerNotes[cantUseNOPURGE] != nil {
		//			featureMap[NoPurge] = providerNotes[cantUseNOPURGE]
		//		} else {
		//			featureMap.SetSimple(
		//				NoPurge,
		//				false,
		//				func() bool { return !providers.ProviderHasCapability(providerName, cantUseNOPURGE) },
		//			)
		//		}
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
	Providers           map[string]FeatureMap
	FeatureTables       [][]string
	FeatureTablesTitles []string
}

func replaceInlineContent(
	file string,
	startMarker string,
	endMarker string,
	newContent string,
) {
	contentBytes, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	content := string(contentBytes)

	start := strings.Index(content, startMarker)
	end := strings.Index(content, endMarker)

	newContentString := startMarker + "\n" + newContent + endMarker
	newContentBytes := []byte(newContentString)
	contentBytes = []byte(content)
	contentBytes = append(contentBytes[:start], append(newContentBytes, contentBytes[end+len(endMarker):]...)...)

	err = os.WriteFile(file, contentBytes, 0o644)
	if err != nil {
		panic(err)
	}
}
