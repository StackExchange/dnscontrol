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

			cat[typeName] = RTypeConfig{
				Name:   typeName,
				Fields: fields,
			}
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
	fatalIfErr(err)
	f, err := os.Create(filename)
	fatalIfErr(err)
	defer f.Close()
	_, err = f.Write(formatted)
	fatalIfErr(err)
}

func main() {
	var err error

	catalog := TypeCatalog{}

	hints := GetHints()
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
	fmt.Printf("DEBUG: cat+hints = %+v\n", catalog)

	values := Values{
		Types:              catalog,
		TypeNames:          catalog.TypeNamesAsSlice(),
		TypeNamesAndFields: catalog.TypeNamesAndFields(),
	}

	var modelsContents []byte

	/*

		Register
		RegType
		for x in type:
		   TypeTYPE
		   ParseTYPE
	*/

	// Generate init() and MustRegisterTypes() statements.
	modelsContents = append(modelsContents, makeInit(values)...)
	// Generate RecordType interface constraint.
	modelsContents = append(modelsContents, makeInterfaceConstraint(values)...)

	fmt.Printf("DEBUG: TypeNames = %v\n", values.TypeNames)
	for _, typeName := range values.TypeNames {
		// Generate TypeTYPE type.
		modelsContents = append(modelsContents, makeTypeTYPE(values.Types[typeName])...)
		// Generate ParseA
		//modelsContents = append(modelsContents, makeParseTYPE(values.Types[typeName])...)
		// Generate PopulateFromRawA
		// Generate AsA
		// Generate GetFields()
		// Generate GetFieldsAsStringsA()
	}

	writeTo("generated_types.go", modelsContents)
}
