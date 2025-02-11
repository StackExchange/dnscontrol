package main

import (
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
	FieldsAsSignature string

	// the field names, prefixed by "s" if they are not a string.
	FieldsAsSVars string

	// If tag dnscontrol "ttl1" is present, this field is set to true
	TTL1 bool
}

type Field struct {
	Name string          // Name of the field ("Port", "Target", etc)
	Type string          // Go type of the field ("uint16", "string", etc)
	Tags *structtag.Tags // Go "tags" for the field.

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

func (cat *TypeCatalog) TypeNamesAndFields() []struct {
	Name   string
	Fields []Field
	Tags   string
} {
	var keys []struct {
		Name   string
		Fields []Field
		Tags   string
	}
	for _, k := range (*cat).TypeNamesAsSlice() {
		v := (*cat)[k]
		keys = append(keys, struct {
			Name   string
			Fields []Field
			Tags   string
		}{
			Name:   k,
			Fields: v.Fields,
			Tags:   v.Tags,
		})
	}
	return keys
}

// Fix Types:

func mkConstructAll(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "srdisplay") {
			ac = append(ac, fmt.Sprintf(`%s: cfSingleRedirecttargetFromRaw(srname, code, srwhen, srthen)`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "parsereturnunknowable") {
			ac = append(ac, fmt.Sprintf(`%s: "UNKNOWABLE"`, field.Name))
		} else if HasTagOption(field.Tags, "dnscontrol", "noparsereturn") {
			// Skip this field.
		} else {
			ac = append(ac, fmt.Sprintf("%s: %s", field.Name, field.NameLower))
		}
	}
	return strings.Join(ac, ", ")
}

func mkFieldsAsSignature(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if field.Type == "fieldtypes.IPv4" {
			// accept input as string.
			ac = append(ac, fmt.Sprintf("%s string", field.NameLower))
		} else {
			ac = append(ac, fmt.Sprintf("%s %s", field.NameLower, field.Type))
		}
	}
	return strings.Join(ac, ", ")
}

func mkFieldTypesCommaSep(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if !HasTagOption(field.Tags, "dnscontrol", "noraw") {
			ac = append(ac, field.Type)
		}
	}
	return strings.Join(ac, ", ")
}

func mkReturnIndividualFieldsList(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noraw") {
			continue
		}
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		} else if field.Type == "fieldtypes.IPv4" {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return strings.Join(ac, ", ")
}

func mkReturnAsStringsList(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dnscontrol", "noraw") {
			continue
		}
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "fieldtypes.IPv4" {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "uint16" {
			ac = append(ac, fmt.Sprintf("strconv.Itoa(int(n.%s))", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return fmt.Sprintf("[%d]string{", len(ac)) + strings.Join(ac, ", ") + "}"
}

func mkFieldsAsSVars(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if HasTagOption(field.Tags, "dns", "a") {
			ac = append(ac, field.NameLower)
		} else if field.Type == "string" {
			ac = append(ac, field.NameLower)
		} else {
			ac = append(ac, "s"+field.NameLower)
		}
	}
	return strings.Join(ac, ", ")
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

			t.FieldsAsSignature = mkFieldsAsSignature(t.Fields)
			t.FieldsAsSVars = mkFieldsAsSVars(t.Fields)
		}
		(*cat)[catName] = t
	}
}

// Fix Fields:

func capFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func parserFor(i int, f Field) string {
	switch ty := f.Type; ty {
	case "int":
		if HasTagOption(f.Tags, "dnscontrol", "redirectcode") {
			return fmt.Sprintf(`fieldtypes.ParseRedirectCode(rawfields[%d], "", origin)`, i)
		}
		return fmt.Sprintf(`fieldtypes.ParseStringTrimmed(rawfields[%d])`, i)
	case "string":
		//fmt.Printf("DEBUG: parserFor(%d, %+v) ... %v\n", i, f, HasTagOption(f.Tags, "dns", "cdomain-name"))
		if HasTagOption(f.Tags, "dns", "cdomain-name") || HasTagOption(f.Tags, "dns", "domain-name") {
			return fmt.Sprintf(`fieldtypes.ParseHostnameDot(rawfields[%d], "", origin)`, i)
		}
		return fmt.Sprintf(`fieldtypes.ParseStringTrimmed(rawfields[%d])`, i)
	case "fieldtypes.IPv4":
		return fmt.Sprintf(`fieldtypes.ParseIPv4(rawfields[%d])`, i)
	}

	return fmt.Sprintf(`fieldtypes.Parse%s(rawfields[%d])`, capFirst(f.Type), i)
}
func mkConvertToString(f Field) string {
	if HasTagOption(f.Tags, "dns", "a") {
		//return fmt.Sprintf("s%s, _ := fieldtypes.ParseIPv4(%s)", f.NameLower, f.NameLower)
		return ""
	}

	switch f.Type {

	case "string":
		return ""

	case "uint16":
		return fmt.Sprintf("s%s := strconv.Itoa(int(%s))", f.NameLower, f.NameLower)

	case "int":
		return fmt.Sprintf("s%s := strconv.Itoa(%s)", f.NameLower, f.NameLower)
	}

	return fmt.Sprintf("s%s := UNKNOWN(int(%s))", f.NameLower, f.NameLower)

}

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
				f.Parser = parserFor(i, f)
			}
			(*cat)[rtype.Name].Fields[i] = f
		}
	}
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
				return fmt.Errorf("Duplicate Rtype Name: %v\n", typeName)
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
				}
			}

		}
	}

	return nil
}
