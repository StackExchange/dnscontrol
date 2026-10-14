package providers

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	dnsv2 "codeberg.org/miekg/dns"
	"github.com/DNSControl/dnscontrol/v5/pkg/privatetypes"
)

func TestSupportedTypesDefaults(t *testing.T) {
	for _, tc := range []struct {
		name      string
		selectors []string
		features  DocumentationNotes
		allowed   []string
	}{
		{"nil", nil, nil, []string{"A", "AAAA", "CAA", "CNAME", "MX", "NS", "SRV", "TXT"}},
		{"Basic8", []string{"Basic8"}, nil, []string{"A", "AAAA", "CAA", "CNAME", "MX", "NS", "SRV", "TXT"}},
		{"empty", []string{}, nil, nil},
		{"nil with features", nil, DocumentationNotes{CanUseDS: Can()}, []string{"A", "AAAA", "CAA", "CNAME", "DS", "MX", "NS", "SRV", "TXT"}},
		{"empty with features", []string{}, DocumentationNotes{CanUseCAA: Can()}, []string{"CAA"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			isolateDefinitions(t)
			Register[*definitionDNS]("TEST", Definition{FriendlyName: "Test", SupportedTypes: tc.selectors, Features: tc.features})
			d, _ := GetDefinition("TEST")
			for _, typ := range privatetypes.RecordTypes() {
				if got := d.RecordTypeSupport(typ.Name).HasFeature; got != slices.Contains(tc.allowed, typ.Name) {
					t.Errorf("support for %s = %v", typ.Name, got)
				}
			}
		})
	}
}

func TestSupportedTypesResolution(t *testing.T) {
	for _, tc := range []struct {
		name      string
		selectors []string
		features  DocumentationNotes
		rtype     string
		want      *DocumentationNote
	}{
		{"RFC ordinary", []string{"RFC"}, nil, "HINFO", Can()},
		{"lowercase Basic8", []string{"basic8"}, nil, "CAA", Can()},
		{"RFC pseudo", []string{"RFC"}, nil, "AKAMAITLC", Cannot()},
		{"star pseudo", []string{"*"}, nil, "AKAMAITLC", Can()},
		{"unknown", []string{"*"}, nil, "DOES_NOT_EXIST", Cannot()},
		{"protocol type", []string{"*"}, nil, "AXFR", Cannot()},
		{"internal command", []string{"*"}, nil, "IMPORT_TRANSFORM", Cannot()},
		{"prefix", []string{"BUNNY_*"}, nil, "BUNNY_DNS_PZ", Can()},
		{"whole name", []string{"BUNNY*:Can"}, nil, "A", Cannot()},
		{"no match", []string{"FUTURE_*"}, nil, "A", Cannot()},
		{"lowercase type", []string{"caa:Can"}, nil, "caa", Can()},
		{"exact reason", []string{"Basic8", "SOA:Cannot:Vendor promises this in 2027"}, nil, "SOA", Cannot("Vendor promises this in 2027")},
		{"supported reason", []string{"CAA:Can:API ready"}, nil, "CAA", Can("API ready")},
		{"unimplemented reason", []string{"SOA:Unimplemented:Awaiting API support"}, nil, "SOA", Unimplemented("Awaiting API support")},
		{"pattern reason", []string{"BUNNY_*:Cannot:Needs vendor support"}, nil, "BUNNY_DNS_PZ", Cannot("Needs vendor support")},
		{"reason containing colon", []string{"SOA:Cannot:ETA: 2027"}, nil, "SOA", Cannot("ETA: 2027")},
		{"reason with URL", []string{"CAA:Can:Read the API guide: https://example.com/docs"}, nil, "CAA", Can("Read the API guide", "https://example.com/docs")},
		{"HTTP URL", []string{"SOA:Cannot:API limitation: http://example.com/docs"}, nil, "SOA", Cannot("API limitation", "http://example.com/docs")},
		{"URL must be last word", []string{"CAA:Can:Read: https://example.com/docs before using"}, nil, "CAA", Can("Read: https://example.com/docs before using")},
		{"exact exclusion", []string{"RFC", "CAA:Cannot"}, nil, "CAA", Cannot()},
		{"Basic8 exclusion", []string{"Basic8", "CAA:Cannot"}, nil, "CAA", Cannot()},
		{"Basic8 legacy exception", []string{"Basic8"}, DocumentationNotes{CanUseCAA: Cannot("legacy limit")}, "CAA", Cannot("legacy limit")},
		{"Basic8 legacy unimplemented", []string{"Basic8"}, DocumentationNotes{CanUseSRV: Unimplemented("pending")}, "SRV", Unimplemented("pending")},
		{"pattern exclusion", []string{"*", "BUNNY_*:Cannot"}, nil, "BUNNY_DNS_PZ", Cannot()},
		{"exact inclusion", []string{"*:Cannot", "CAA"}, nil, "CAA", Can()},
		{"unimplemented pattern", []string{"*", "AKAMAI*:Unimplemented"}, nil, "AKAMAITLC", Unimplemented()},
		{"addresses do not imply TLC", []string{"A", "AAAA"}, nil, "AKAMAITLC", Cannot()},
		{"legacy exception", []string{"*"}, DocumentationNotes{CanUseCAA: Cannot("legacy limit")}, "CAA", Cannot("legacy limit")},
		{"legacy unimplemented", []string{"RFC"}, DocumentationNotes{CanUseCAA: Unimplemented("pending")}, "CAA", Unimplemented("pending")},
		{"annotated pattern beats legacy", []string{"C*:Can"}, DocumentationNotes{CanUseCAA: Cannot("obsolete")}, "CAA", Can()},
		{"exact beats legacy", []string{"CAA:Cannot"}, DocumentationNotes{CanUseCAA: Can("obsolete")}, "CAA", Cannot()},
		{"keep compatible note", []string{"CAA"}, DocumentationNotes{CanUseCAA: Can("API note", "https://example.com/api")}, "CAA", Can("API note", "https://example.com/api")},
		{"inline reason and legacy link", []string{"CAA:Cannot:new reason"}, DocumentationNotes{CanUseCAA: Cannot("old reason", "https://example.com/api")}, "CAA", Cannot("new reason", "https://example.com/api")},
		{"inline URL beats legacy URL", []string{"CAA:Can:new reason: https://example.com/new"}, DocumentationNotes{CanUseCAA: Can("old reason", "https://example.com/old")}, "CAA", Can("new reason", "https://example.com/new")},
		{"replace distinct status", []string{"CAA:Cannot"}, DocumentationNotes{CanUseCAA: Unimplemented("obsolete")}, "CAA", Cannot()},
		{"duplicates", []string{"CAA", "caa:Can"}, nil, "CAA", Can()},
		{"resolve patterns", []string{"CA*:Cannot", "*AA:Unimplemented", "CAA"}, nil, "CAA", Can()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, reverse := range []bool{false, true} {
				t.Run(map[bool]string{false: "forward", true: "reverse"}[reverse], func(t *testing.T) {
					isolateDefinitions(t)
					selectors := slices.Clone(tc.selectors)
					if reverse {
						slices.Reverse(selectors)
					}
					Register[*definitionDNS]("TEST", Definition{FriendlyName: "Test", Aliases: []string{"TEST_ALIAS"}, SupportedTypes: selectors, Features: tc.features})
					d, _ := GetDefinition("TEST_ALIAS")
					got := d.RecordTypeSupport(tc.rtype)
					if got != *tc.want {
						t.Fatalf("got %+v, want %+v", got, *tc.want)
					}
					if capability, ok := RecordTypeCapabilities[strings.ToUpper(tc.rtype)]; ok && *d.DerivedFeatures[capability] != got {
						t.Fatal("derived capability disagrees with type support")
					}
				})
			}
		})
	}
}

func TestSupportedTypesErrors(t *testing.T) {
	for _, selector := range []string{"", " CAA", "CAA ", "CA?", "[A]", "A,B", "CAA:", "CAA:can", "CAA:cAn", "CAA::reason", "CAA:Cannot:", "CAA:Cannot:   ", "CAA:Can:: https://example.com", "CAA:Can:See docs: httpbad", "RFC:Cannot", "Basic8:Can"} {
		t.Run(selector, func(t *testing.T) {
			isolateDefinitions(t)
			assertRegistrationPanics(t, "provider \"INVALID\"", func() {
				Register[*definitionDNS]("INVALID", Definition{FriendlyName: "Invalid", SupportedTypes: []string{
					selector,
				}})
			})
		})
	}
	for _, selectors := range [][]string{
		{"Standard"}, {"Default"}, {"UNREGISTERED"}, {"IMPORT_TRANSFORM"}, {"CAA", "CAA:Cannot"},
		{"CAA:Cannot", "CAA:Unimplemented"}, {"CA*:Cannot", "*AA:Can"},
	} {
		t.Run(strings.Join(selectors, ","), func(t *testing.T) {
			isolateDefinitions(t)
			Register[*definitionDNS]("INVALID", Definition{FriendlyName: "Invalid", SupportedTypes: selectors})
			err := Finalize()
			if err == nil || !strings.Contains(err.Error(), "INVALID") {
				t.Fatalf("error = %v", err)
			}
			for _, selector := range selectors {
				if !strings.Contains(err.Error(), selector) {
					t.Errorf("error %q omits %q", err, selector)
				}
			}
		})
	}
}

func TestSupportedTypesFinalization(t *testing.T) {
	isolateDefinitions(t)
	codepoint := uint16(64001)
	for dnsv2.TypeToString[codepoint] != "" {
		codepoint++
	}
	lateName := fmt.Sprintf("STAGEFIVELATE%d", codepoint)
	Register[*definitionDNS]("WILDCARD", Definition{FriendlyName: "Wildcard", SupportedTypes: []string{
		"*",
	}})
	Register[*definitionDNS]("RFC", Definition{FriendlyName: "RFC", SupportedTypes: []string{
		"RFC",
	}})
	wildcard, _ := GetDefinition("WILDCARD")
	Register[*definitionDNS]("EXACT", Definition{FriendlyName: "Exact", SupportedTypes: []string{
		lateName,
	}})
	if err := Finalize(); err == nil {
		t.Fatal("unknown exact type accepted")
	}
	// Deliberately register after providers (and after an earlier finalization).
	// No underscore, and outside the private-use numeric range: classification
	// must come from the catalog, not the name or number.
	privatetypes.Register(codepoint, lateName, func() dnsv2.RR { return &dnsv2.A{} }, nil)
	if err := Finalize(); err != nil {
		t.Fatal(err)
	}
	exact, _ := GetDefinition("EXACT")
	rfc, _ := GetDefinition("RFC")
	if !wildcard.RecordTypeSupport(lateName).HasFeature || !exact.RecordTypeSupport(lateName).HasFeature || rfc.RecordTypeSupport(lateName).HasFeature {
		t.Fatal("late catalog registration was not resolved using explicit classification")
	}
	before := wildcard.DerivedFeatures
	if err := Finalize(); err != nil {
		t.Fatal(err)
	}
	if reflect.ValueOf(before).Pointer() != reflect.ValueOf(wildcard.DerivedFeatures).Pointer() {
		t.Fatal("unchanged finalization replaced derived metadata")
	}
	// Accessors and type matching must be read-only once finalized.
	var wg sync.WaitGroup
	for range 10 {
		wg.Go(func() {
			for range 10 {
				def, _ := GetDefinition("WILDCARD")
				_ = def.RecordTypeSupport("AKAMAITLC")
				_ = ProviderHasCapability("WILDCARD", CanUseCAA)
				_ = AllDefinitions()
			}
		})
	}
	wg.Wait()

	// A later catalog addition can also reveal an overlap between patterns
	// that previously matched nothing. Finalization must check it again.
	conflictName := lateName + "CONFLICT"
	Register[*definitionDNS]("CONFLICT", Definition{FriendlyName: "Conflict", SupportedTypes: []string{
		conflictName + "*:Can",
		"*" + conflictName + ":Cannot",
	}})
	if err := Finalize(); err != nil {
		t.Fatal(err)
	}
	privatetypes.Register(codepoint+1, conflictName, func() dnsv2.RR { return &dnsv2.A{} }, nil)
	if err := Finalize(); err == nil || !strings.Contains(err.Error(), conflictName) {
		t.Fatalf("late overlap was not rejected: %v", err)
	}
}

func TestSupportedTypesOperationalCapabilities(t *testing.T) {
	isolateDefinitions(t)
	Register[*definitionDNS]("ALL", Definition{FriendlyName: "All", SupportedTypes: []string{
		"*",
	}, CanUseDSForChildren: Cannot("obsolete")})
	d, _ := GetDefinition("ALL")
	if !ProviderHasCapability("ALL", CanUseDSForChildren) || d.DerivedFeatures[CanUseDSForChildren].Comment != "" {
		t.Fatal("general DS support must imply child support without contradictory notes")
	}
	for _, capability := range []Capability{CanAutoDNSSEC, CanConcur, CanDualHost, OfficiallySupported} {
		if ProviderHasCapability("ALL", capability) {
			t.Fatalf("wildcard enabled operational capability %s", capability)
		}
	}
}
