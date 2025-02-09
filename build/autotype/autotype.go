package main

import (
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

type Values struct {
	Types              TypeCatalog
	TypeNames          []string
	TypeNamesAndFields []struct {
		Name   string
		Fields []Field
		Tags   string
	}
}

// getTypeStruct will take a type and the package scope, and return the
// (innermost) struct if the type is considered a RR type (currently defined as
// those structs beginning with a RR_Header, could be redefined as implementing
// the RR interface). The bool return value indicates if embedded structs were
// resolved.
func getTypeStruct(t types.Type, scope *types.Scope) (*types.Struct, bool) {
	st, ok := t.Underlying().(*types.Struct)
	if !ok {
		return nil, false
	}
	if st.NumFields() == 0 {
		return nil, false
	}
	if st.Field(0).Type().String() == "github.com/miekg/dns.RR_Header" {
		return st, false
	}
	if st.Field(0).Type() == scope.Lookup("RR_Header").Type() {
		return st, false
	}
	if st.Field(0).Anonymous() {
		st, _ := getTypeStruct(st.Field(0).Type(), scope)
		return st, true
	}
	return nil, false
}

// loadModule retrieves package description for a given module.
func loadModule(name string) (*types.Package, error) {
	conf := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(&conf, name)
	if err != nil {
		return nil, err
	}
	return pkgs[0].Types, nil
}

// ExtractTypeDataFromModule reads the Go source code from modName and extracts
// TypeCatalog data from it. The filter is a set of type names to extract. If
// filter is empty, all types are extracted.
func ExtractTypeDataFromModule(modName string, filter map[string]struct{}) (TypeCatalog, error) {

	//fmt.Printf("DEBUG: Reading module %s; filter=%v\n", modName, filter)

	// Import and type-check the package
	pkg, err := loadModule(modName)
	fatalIfErr(err)
	scope := pkg.Scope()

	// Collect actual types (*X)
	var namedTypes []string
	for _, name := range scope.Names() {
		o := scope.Lookup(name)
		if o == nil || !o.Exported() {
			continue
		}
		if st, _ := getTypeStruct(o.Type(), scope); st == nil {
			continue
		}
		if name == "PrivateRR" {
			continue
		}

		namedTypes = append(namedTypes, o.Name())
	}

	cat := TypeCatalog{}

	for _, typeName := range namedTypes {

		if len(filter) != 0 {
			if _, ok := filter[typeName]; !ok {
				continue
			}
		}
		fmt.Printf("DEBUG: DOING %s\n", typeName)

		o := scope.Lookup(typeName)
		st, isEmbedded := getTypeStruct(o.Type(), scope)
		if isEmbedded {
			continue
		}

		var fields []Field
		for i := 1; i < st.NumFields(); i++ {
			if _, ok := st.Field(i).Type().(*types.Slice); ok {
				fieldname := st.Field(i).Name()
				slicetype := st.Field(i).Type().String()
				fieldtags := st.Tag(i)

				fields = append(fields, Field{
					Name: fieldname,
					Type: slicetype,
					Tags: mkTagString(fieldtags),
				})

			} else {
				fieldname := st.Field(i).Name()
				fieldtype := st.Field(i).Type().String()
				if fieldtype == "net.IP" {
					fieldtype = "fieldtypes.IPv4"
				}
				fieldtags := st.Tag(i)

				fields = append(fields, Field{
					Name: fieldname,
					Type: fieldtype,
					Tags: fieldtags,
				})

			}
		}
		if len(fields) != st.NumFields()-1 {
			fmt.Printf("WARNING: field count mismatch len(fields)=%d st.NumFields()=%d\n", len(fields), st.NumFields())
		}

		cat[typeName] = RTypeConfig{
			Name:   typeName,
			Fields: fields,
		}
	}

	return cat, nil
}

func fatalIfErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func fatalIfErr2(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func writeTo(filename string, contents []byte) {
	formatted, err := format.Source(contents)
	if err != nil {
		log.Printf("failed to format: %s", err)
		formatted = contents
	}
	f, err := os.Create(filename)
	fatalIfErr(err)
	defer f.Close()
	_, err = f.Write(formatted)
	fatalIfErr(err)
}

func main() {
	var err error

	catalog := TypeCatalog{}

	typeNames, hints := GetHints()
	fatalIfErr2(err, "failed to get hints")
	filter := hints.TypeNamesAsSet()

	for _, moduleName := range []string{
		"github.com/miekg/dns",
		"github.com/StackExchange/dnscontrol/v4/providers/cloudflare/customtypes",
	} {

		td, err := ExtractTypeDataFromModule(moduleName, filter)
		if err != nil {
			log.Fatalf("failed to extract from %s: %s", moduleName, err)
		}
		err = catalog.MergeCat(td)
		if err != nil {
			log.Fatalf("failed to merge from %s: %s", moduleName, err)
		}
	}

	catalog.MergeHints(hints) // Overwrite catalog items with data from hints.
	//fmt.Printf("DEBUG: cat+hints = %+v\n", catalog)

	catalog.FixFields() // must be called before FixTypes
	catalog.FixTypes()
	values := Values{
		Types:              catalog,
		TypeNames:          typeNames,
		TypeNamesAndFields: catalog.TypeNamesAndFields(),
	}

	var m []byte

	// Generate init() with the MustRegisterTypes() statements.
	m = append(m, makeInit(values)...)

	// Generate the RecordType interface constraint.
	m = append(m, makeInterfaceConstraint(values)...)

	//fmt.Printf("DEBUG: Types: %s\n", values.TypeNames)
	for _, typeName := range values.TypeNames {
		fmt.Printf("DEBUG: Generating for %s\n", typeName)
		// Generate Type$TYPE type.
		m = append(m, makeTypeTYPE(values.Types[typeName])...)
		// Generate Parse$TYPE
		m = append(m, makeParseTYPE(values.Types[typeName])...)
		// Generate PopulateFromRawA
		m = append(m, makePopulateFromRawTYPE(values.Types[typeName])...)
		// Generate AsA
		m = append(m, makeAsTYPE(values.Types[typeName])...)
		// Generate GetFields()
		m = append(m, makeGetFieldsTYPE(values.Types[typeName])...)
		// Generate GetFieldsAsStringsA()
		m = append(m, makeGetFieldsAsStringsTYPE(values.Types[typeName])...)
	}

	writeTo("generated_types.go", m)
}
