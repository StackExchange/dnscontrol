package normalize

import (
	"go/ast"
	"go/token"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const (
	providersImportDir   = "../../pkg/providers"
	providersPackageName = "providers"
)

func TestCapabilitiesAreFiltered(t *testing.T) {
	// Any capabilities which we wish to whitelist because it's not directly
	// something we can test against.
	skipCheckCapabilities := make(map[string]struct{})
	// skipCheckCapabilities["CanUseBlahBlahBlah"] = struct{}{}

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps,
		Dir:  providersImportDir,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil || len(pkgs) == 0 {
		t.Fatalf("unable to load Go code from providers: %v", err)
	}
	var providers *packages.Package
	for _, pkg := range pkgs {
		if pkg.Name == providersPackageName {
			providers = pkg
			break
		}
	}
	if providers == nil {
		t.Fatalf("did not find package %q in %q", providersPackageName, providersImportDir)
	}

	constantNames := make([]string, 0, 50)
	capabilityInts := make(map[string]int, 50)

	for _, file := range providers.Syntax {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.CONST {
				continue
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range valueSpec.Names {
					if !strings.HasPrefix(name.Name, "CanUse") {
						continue
					}
					constantNames = append(constantNames, name.Name)
					// We can't get the int value easily without type info, so just use a dummy value
					capabilityInts[name.Name] = 0
				}
			}
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
