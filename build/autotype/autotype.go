package main

import (
	"encoding/json"
	"fmt"
	"go/format"
	"go/types"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

type TypeInfo struct {
	Name   string
	Config *RTypeConfig
	Fields []Field
	Tags   string
}

type Values struct {
	Types              TypeCatalog
	TypeNames          []string
	TypeNamesAndFields []TypeInfo
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
					Tags: MustParseTags(fieldtags),
				})

			} else {
				fieldname := st.Field(i).Name()
				fieldtags := MustParseTags(st.Tag(i))
				fieldtype := st.Field(i).Type().String()
				if HasTagOption(fieldtags, "dns", "a") {
					fieldtype = "fieldtypes.IPv4"
				}
				if HasTagOption(fieldtags, "dns", "aaaa") {
					fieldtype = "fieldtypes.IPv6"
				}

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

func writeTo(contents []byte, filename string) {
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

func writeToJS(contents []byte, filename string) {
	formatted := contents
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
		TypeNamesAndFields: catalog.TypeNamesAndFields(typeNames),
	}

	// models/generated_types.go
	var mgt []byte
	// Generate init() with the MustRegisterTypes() statements.
	mgt = append(mgt, makeInit(values)...)
	// Generate the RecordType interface constraint.
	mgt = append(mgt, makeInterfaceConstraint(values)...)
	// Generate the makeImportFromLegacy() function.
	//	fmt.Printf("DEBUG: Values: %+v\n", values)
	x, _ := json.MarshalIndent(values, "", "    ")
	fmt.Printf("DEBUG: Values: %s\n", x)
	mgt = append(mgt, makeImportFromLegacy(values)...)
	mgt = append(mgt, makeSeal(values)...)
	mgt = append(mgt, makeCopy(values)...)
	mgt = append(mgt, makePopulateFromFields(values)...)
	mgt = append(mgt, makeGetTargetField(values)...)
	//mgt = append(mgt, makeSetTarget(values)...)

	// integrationTest/generated_helpers.go
	var ith = makeIntTestHeader()

	// pkg/js/helpers-types.js
	var hrt = makehelpersRawRecordBuilder(values)

	//fmt.Printf("DEBUG: Types: %s\n", values.TypeNames)
	for _, typeName := range values.TypeNames {
		fmt.Printf("DEBUG: Generating for %s\n", typeName)

		// models/generated_types.go

		// Generate Type$TYPE type.
		mgt = append(mgt, makeTypeTYPE(values.Types[typeName])...)
		// Generate Parse$TYPE
		mgt = append(mgt, makeParseTYPE(values.Types[typeName])...)
		// Generate PopulateFromRawA
		mgt = append(mgt, makePopulateFromRawTYPE(values.Types[typeName])...)
		// Generate AsA
		mgt = append(mgt, makeAsTYPE(values.Types[typeName])...)
		// Generate GetFields()
		mgt = append(mgt, makeGetFieldsTYPE(values.Types[typeName])...)
		// Generate GetFieldsAsStringsA()
		mgt = append(mgt, makeGetFieldsAsStringsTYPE(values.Types[typeName])...)
		// Generate SetTargetTYPE()
		mgt = append(mgt, makeSetTargetTYPE(values.Types[typeName])...)

		// integrationTest/generated_helpers.go

		// Generate type() constructor
		ith = append(ith, makeIntTestConstructor(values.Types[typeName])...)

	}

	writeTo(mgt, "generated_types.go")
	writeTo(ith, "../integrationTest/generated_helpers.go")
	writeToJS(hrt, "../pkg/js/helpers-types.js")
	writeTo(ith, "../integrationTest/generated_helpers.go")
}
