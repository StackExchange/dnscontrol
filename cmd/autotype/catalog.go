package main

import (
	"fmt"
)

type TypeCatalog map[string]RTypeConfig

type RTypeConfig struct {
	Token  string
	Fields []Field
	Tags   string
}

type Field struct {
	Name string
	Type string
	Tags string
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
	for k, v := range *cat {
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
