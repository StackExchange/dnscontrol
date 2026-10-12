package models

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	dnsv2 "codeberg.org/miekg/dns"
	dnsrdatav2 "codeberg.org/miekg/dns/rdata"
)

// RDtoFields extracts the exported fields from an RD into []any.
// Fileds prefixed with "RT_" are not included, as they are "Runtime" values, not part of the RD itself.
func RDtoFields(s dnsv2.RDATA) ([]any, error) {
	v := reflect.ValueOf(s)

	// Dereference pointer if a pointer to a struct was passed
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
		// } else {
		// 	return nil, fmt.Errorf("expected pointer to struct, got %v", v.Kind())
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %v", v.Kind())
	}

	t := v.Type()

	var result []any
	for i := 0; i < v.NumField(); i++ {
		fieldType := t.Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" && !fieldType.Anonymous {
			continue // skip private fields
		}
		if strings.HasPrefix(fieldType.Name, "RT_") {
			continue
		}

		result = append(result, v.Field(i).Interface())
	}

	return result, nil
}

// RDtoFieldsStrings returns the fields of a RDATA struct as strings.
// Floats are restricted to 2 digits after the decimal.
func RDtoFieldsStrings(s dnsv2.RDATA) ([]string, error) {
	items, err := RDtoFields(s)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(items))
	for i, item := range items {
		if types, ok := item.([]uint16); ok && hasTypeBitMap(s) {
			result[i] = typeBitMapString(types)
			continue
		}
		switch v := item.(type) {
		case string:
			result[i] = v
		case float32, float64:
			result[i] = fmt.Sprintf("%.2f", v)
		default:
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// RDtoFieldsJS returns the fields of a RDATA struct as Javascript-constants.
// Examples: int: `2`; string:  `"mystring"`
// Floats are restricted to 2 digits after the decimal.
func RDtoFieldsJS(s dnsv2.RDATA) ([]string, error) {

	// Special cases:

	switch v := s.(type) {
	case dnsrdatav2.A, dnsrdatav2.AAAA:
		return []string{`"` + v.String() + `"`}, nil
	}

	// General case:

	items, err := RDtoFields(s)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(items))
	for i, item := range items {
		if types, ok := item.([]uint16); ok && hasTypeBitMap(s) {
			b, err := json.Marshal(typeBitMapString(types))
			if err != nil {
				return nil, err
			}
			result[i] = string(b)
			continue
		}
		switch v := item.(type) {
		case string:
			// https://stackoverflow.com/questions/51691901
			b, err := json.Marshal(item)
			if err != nil {
				return nil, err
			}
			result[i] = string(b)
		case float32, float64:
			result[i] = fmt.Sprintf("%.2f", v)
		default:
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// hasTypeBitMap reports whether rd is a type whose []uint16 field is a type
// bitmap (a list of record types). Other types may use []uint16 for other
// purposes, so their fields are left alone.
func hasTypeBitMap(rd dnsv2.RDATA) bool {
	switch rd.(type) {
	case dnsrdatav2.CSYNC, dnsrdatav2.NSEC, dnsrdatav2.NSEC3:
		return true
	}
	return false
}

// typeBitMapString formats a type bitmap as a space-separated list of type
// names, such as "A NS AAAA".
func typeBitMapString(types []uint16) string {
	names := make([]string, len(types))
	for i, typ := range types {
		names[i] = dnsv2.TypeToString[typ]
		if names[i] == "" {
			names[i] = fmt.Sprintf("TYPE%d", typ)
		}
	}
	return strings.Join(names, " ")
}
