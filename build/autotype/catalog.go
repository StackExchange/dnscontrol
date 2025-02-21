package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/structtag"
)

type TypeCatalog map[string]RTypeConfig

type RTypeConfig struct {
	Name    string  // Name of the type ("A, "MX", etc)
	Token   string  // String to use in type RecordConfig.Type
	Tags    string  // Tags for the struct.
	NoLabel bool    // First element of RawFields is data, not a label.
	Fields  []Field // A description of each field.

	// Generated fields:

	// .Name all lowercase
	NameLower string

	// Number of fields in the struct.
	NumRawFields int

	// A string of the form "field1: type1, field2: type2, etc"
	ConstructAll string

	// A comma-separated list of the field types.
	FieldTypesCommaSep string

	// A comma-separated list of the field names, accessed as n.Field.
	ReturnIndividualFieldsList string

	// used to return the fields, each converted to a string.
	ReturnAsStringsList string

	// Fields as used in a function signature.
	InputFieldsAsSignature string

	// the field names, prefixed by "s" if they are not a string.
	FieldsAsSVars string

	// If tag dnscontrol "ttl1" is present, this field is set to true
	TTL1 bool

	// Assign fields to their legacy counterparts.
	ConstructFromLegacyFields string

	// Go expression that generates the .Comparable field.
	ComparableExpr string

	// NB(tlim): Fields in this struct are populated by FixTypes().
}

type Field struct {
	Name string          // Name of the field ("Port", "Target", etc)
	Type string          // Go type of the field ("uint16", "string", etc)
	Tags *structtag.Tags // Go "tags" for the field.

	// legacy field name from RecordConfig.
	LegacyName string

	// Generated fields:

	// name of the field in lowercase
	NameLower string

	// Go code to parse the field.
	Parser string

	// Tags as a string
	TagsString string

	// Go code to convert this field to string (or "" if it is a string)
	ConvertToString string

	// This field does not come from user-input (i.e. not part of RawFields)
	NoRaw bool

	// NB(tlim): If you add a new field here:
	// * FixFields() should populate it.
	// * Update the Merge() method.
}

func (cat *TypeCatalog) TypeNamesAsSet() map[string]struct{} {
	keys := map[string]struct{}{}
	for k := range *cat {
		keys[k] = struct{}{}
	}
	return keys
}

func (cat *TypeCatalog) TypeNamesAsSlice() []string {
	var keys []string
	for k := range *cat {
		keys = append(keys, k)
	}
	return keys
}

func (cat *TypeCatalog) TypeNamesAndFields(order []string) []TypeInfo {

	x, _ := json.MarshalIndent(order, "", "    ")
	fmt.Printf("DEBUG: TypeNamesAndFields: %v\n", x)

	var keys []TypeInfo
	for _, k := range order {
		v := (*cat)[k]
		keys = append(keys, TypeInfo{
			Name:   k,
			Config: &v,
			Fields: v.Fields,
			Tags:   v.Tags,
		})
	}
	return keys
}

// Fix Types:

// FixTypes generates the generated fields of each RTypeConfig.
func (cat *TypeCatalog) FixTypes() {
	for catName := range *cat {
		t := (*cat)[catName]
		{
			// Default token to Name.
			token := t.Token
			if token == "" {
				token = t.Name
			}

			t.NameLower = strings.ToLower(t.Name)
			t.Token = token
			t.NumRawFields = countFields(t.Fields)
			t.ConstructAll = mkConstructAll(t.Fields)
			t.FieldTypesCommaSep = mkFieldTypesCommaSep(t.Fields)
			t.ReturnIndividualFieldsList = mkReturnIndividualFieldsList(t.Fields)
			t.ReturnAsStringsList = mkReturnAsStringsList(t.Fields)

			t.InputFieldsAsSignature = mkInputFieldsAsSignature(t.Fields)
			t.FieldsAsSVars = mkFieldsAsSVars(t.Fields)
			t.ConstructFromLegacyFields = mkConstructFromLegacyFields(t.Fields)
			t.ComparableExpr = mkComparableExpr(t.Fields)
		}
		(*cat)[catName] = t
	}
}

func countFields(fields []Field) int {
	c := 0
	for _, field := range fields {
		if !HasTagOption(field.Tags, "dnscontrol", "noraw") {
			c++
		}
	}
	return c
}

// Fix Fields:

// FixFields generates the NameLower and Parser fields of each Field.
func (cat *TypeCatalog) FixFields() {
	// Generate per-field data
	for _, rtype := range *cat {
		for i := range rtype.Fields {
			f := (*cat)[rtype.Name].Fields[i]
			{
				f.NameLower = strings.ToLower(f.Name)
				f.ConvertToString = mkConvertToString(f)
				f.TagsString = f.Tags.String()
				f.NoRaw = HasTagOption(f.Tags, "dnscontrol", "noraw")
				f.Parser = mkParser(i, f)
			}
			(*cat)[rtype.Name].Fields[i] = f
		}
	}
}

func capFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

// MergeHints applies hints to the catalog.
func (cat *TypeCatalog) MergeHints(overlay TypeCatalog) { _ = cat.Merge(overlay, true) }

// MergeCat merges a catalog into the catalog.
func (cat *TypeCatalog) MergeCat(overlay TypeCatalog) error { return cat.Merge(overlay, false) }

// Merge merges a catalog into the catalog. If a duplicate RType is found, it
// is only an error if dupesOk == false.
func (cat *TypeCatalog) Merge(overlay TypeCatalog, dupesOk bool) error {

	for typeName, conf := range overlay {
		//fmt.Printf("KEY=%v VALUE=%v\n", typeName, conf)

		if !dupesOk {
			if _, ok := (*cat)[typeName]; ok {
				return fmt.Errorf("duplicate Rtype Name: %v", typeName)
			}
		}

		if _, ok := (*cat)[typeName]; !ok {
			(*cat)[typeName] = conf
		} else {
			// Merge Token.
			if conf.Token != "" {
				x := (*cat)[typeName]
				x.Token = conf.Token
				(*cat)[typeName] = x
			}
			if conf.NoLabel {
				x := (*cat)[typeName]
				x.NoLabel = true
				(*cat)[typeName] = x
			}
			if conf.TTL1 {
				x := (*cat)[typeName]
				x.TTL1 = true
				(*cat)[typeName] = x
			}

			// Merge Fields.

			// Gather hint info as a map:
			fieldHints := map[string]Field{}
			for _, field := range conf.Fields {
				fieldHints[field.Name] = field
			}

			for i, field := range (*cat)[typeName].Fields {
				// Do we have a hint?
				if hint, ok := fieldHints[field.Name]; ok {
					// Overwrite the type
					if hint.Type != "" {
						(*cat)[typeName].Fields[i].Type = hint.Type
					}
					if hint.Tags != nil {
						(*cat)[typeName].Fields[i].Tags = hint.Tags
					}
					if hint.LegacyName != "" {
						(*cat)[typeName].Fields[i].LegacyName = hint.LegacyName
					}
				}
			}

		}
	}

	return nil
}
