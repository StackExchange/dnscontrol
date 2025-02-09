package main

import (
	"fmt"
	"strings"
)

type TypeCatalog map[string]RTypeConfig

type RTypeConfig struct {
	Name   string  // Name of the type ("A, "MX", etc)
	Token  string  // String to use in type RecordConfig.Type
	Tags   string  // Tags for the struct.
	Fields []Field // A description of each field.

	// Generated fields:

	// Number of fields in the struct.
	NumFields int

	// A string of the form "field1: type1, field2: type2, etc"
	ConstructAll string

	// A comma-separated list of the field types.
	FieldTypesCommaSep string

	// A comma-separated list of the field names, accessed as n.Field.
	ReturnIndividualFieldsList string

	// used to return the fields, each converted to a string.
	ReturnAsStringsList string
}

type Field struct {
	Name string // Name of the field ("Port", "Target", etc)
	Type string // Go type of the field ("uint16", "string", etc)
	Tags string // Go "tags" for the field.

	// Generated fields:
	NameLower string // name of the field in lowercase
	Parser    string // Go code to parse the field.
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

func mkTagString(t string) string {
	return t
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
		ac = append(ac, fmt.Sprintf("%s: %s", field.Name, field.NameLower))
	}
	return strings.Join(ac, ", ")
}

func mkFieldTypesCommaSep(fields []Field) string {
	var ac []string
	for _, field := range fields {
		ac = append(ac, field.Type)
	}
	return strings.Join(ac, ", ")
}

func mkReturnIndividualFieldsList(fields []Field) string {
	var ac []string
	for _, field := range fields {
		if strings.Contains(field.Tags, `dns:"a"`) {
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
		if strings.Contains(field.Tags, `dns:"a"`) {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "fieldtypes.IPv4" {
			ac = append(ac, fmt.Sprintf("n.%s.String()", field.Name))
		} else if field.Type == "uint16" {
			ac = append(ac, fmt.Sprintf("strconv.Itoa(int(n.%s))", field.Name))
		} else {
			ac = append(ac, fmt.Sprintf("n.%s", field.Name))
		}
	}
	return fmt.Sprintf("[%d]string{", len(fields)) + strings.Join(ac, ", ") + "}"
}

// FixTypes generates the NumFields and FieldsAndTypes fields of each RTypeConfig.
func (cat *TypeCatalog) FixTypes() {
	for catName, rtype := range *cat {
		t := (*cat)[catName]
		t.NumFields = len(rtype.Fields)
		t.ConstructAll = mkConstructAll(rtype.Fields)
		t.FieldTypesCommaSep = mkFieldTypesCommaSep(rtype.Fields)
		t.ReturnIndividualFieldsList = mkReturnIndividualFieldsList(rtype.Fields)
		t.ReturnAsStringsList = mkReturnAsStringsList(rtype.Fields)
		(*cat)[catName] = t
	}
}

// Fix Fields:

func capFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

func parserFor(i int, f Field) string {
	switch ty := f.Type; ty {
	case "string":
		if strings.Contains(f.Tags, `dns:"cdomain-name"`) || strings.Contains(f.Tags, `dns:"domain-name"`) {
			return fmt.Sprintf(`fieldtypes.ParseHostnameDot(rawfields[%d], "", origin)`, i)
		}
		return fmt.Sprintf(`fieldtypes.ParseStringTrimmed(rawfields[%d])`, i)
	case "fieldtypes.IPv4":
		return fmt.Sprintf(`fieldtypes.ParseIPv4(rawfields[%d])`, i)
	}

	return fmt.Sprintf(`fieldtypes.Parse%s(rawfields[%d])`, capFirst(f.Type), i)
}

// FixFields generates the NameLower and Parser fields of each Field.
func (cat *TypeCatalog) FixFields() {
	// Generate per-field data
	for _, rtype := range *cat {
		for i, field := range rtype.Fields {
			(*cat)[rtype.Name].Fields[i].NameLower = strings.ToLower(field.Name)
			(*cat)[rtype.Name].Fields[i].Parser = parserFor(i, field)
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
				//  x := (*cat)[typeName]
				//  x.Token = conf.Token
				x := (*cat)[typeName]
				x.Token = conf.Token
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
				}
			}

		}
	}

	return nil
}
