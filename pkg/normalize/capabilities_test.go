package normalize

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

const providersImportDir = "../../providers"
const providersPackageName = "providers"

func TestCapabilitiesAreFiltered(t *testing.T) {
	// Any capabilities which we wish to whitelist because it's not directly
	// something we can test against.
	skipCheckCapabilities := make(map[string]struct{})
	skipCheckCapabilities["CanUseTXTMulti"] = struct{}{}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, providersImportDir, nil, 0)
	if err != nil {
		t.Fatalf("unable to load Go code from providers: %s", err)
	}
	providers, ok := pkgs[providersPackageName]
	if !ok {
		t.Fatalf("did not find package %q in %q", providersPackageName, providersImportDir)
	}

	constantNames := make([]string, 0, 50)
	capabilityInts := make(map[string]int, 50)

	// providers.Scope was nil in my testing
	for fileName := range providers.Files {
		scope := providers.Files[fileName].Scope
		for itemName, obj := range scope.Objects {
			if obj.Kind != ast.Con {
				continue
			}
			// In practice, the object.Type is nil here so we can't filter for
			// capabilities so easily.
			if !strings.HasPrefix(itemName, "CanUse") {
				continue
			}
			constantNames = append(constantNames, itemName)
			capabilityInts[itemName] = obj.Data.(int)
		}
	}
	sort.Strings(constantNames)

	if len(providerCapabilityChecks) == 0 {
		t.Fatal("missing entries in providerCapabilityChecks")
	}

	capIntsToNames := make(map[int]string, len(providerCapabilityChecks))
	for _, pair := range providerCapabilityChecks {
		for _, cap := range pair.caps {
			capIntsToNames[int(cap)] = pair.rType
		}
	}

	for _, capName := range constantNames {
		capInt := capabilityInts[capName]
		if _, ok := skipCheckCapabilities[capName]; ok {
			t.Logf("ok: providers.%s (%d) is exempt from checkProviderCapabilities", capName, capInt)
		} else if rType, ok := capIntsToNames[capInt]; ok {
			t.Logf("ok: providers.%s (%d) is checked for with %q", capName, capInt, rType)
		} else {
			t.Errorf("MISSING: providers.%s (%d) is not checked by checkProviderCapabilities", capName, capInt)
		}
	}

}
