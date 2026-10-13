package diff2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	dnsv2 "codeberg.org/miekg/dns"
)

// fieldDiff describes how the rdata changed, field by field. Every field is
// listed in order, and changed fields are shown as name=(old->new):
//
//	preference=(10->20) mx=mx.example.com.
//
// ok is false when a field-level description would not help, and the caller
// should show both records in full instead. That happens when:
//   - the rdata types differ, or either rdata is not a struct;
//   - the struct has a field that is unexported or not a plain number, bool
//     or string (for example TXT strings, type bitmaps, SVCB parameters);
//   - no fields changed, or every field changed.
func fieldDiff(a, b dnsv2.RDATA) (string, bool) {
	va, vb := structValue(a), structValue(b)
	if !va.IsValid() || !vb.IsValid() || va.Type() != vb.Type() {
		return "", false
	}

	t := va.Type()
	parts := make([]string, 0, t.NumField())
	changed := 0
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() || !isScalarKind(f.Type.Kind()) {
			return "", false
		}
		name := strings.ToLower(f.Name)
		from, to := formatField(va.Field(i)), formatField(vb.Field(i))
		if from == to {
			parts = append(parts, name+"="+from)
			continue
		}
		changed++
		parts = append(parts, fmt.Sprintf("%s=(%s->%s)", name, from, to))
	}

	if changed == 0 || changed == t.NumField() {
		return "", false
	}
	return strings.Join(parts, " "), true
}

// structValue returns the struct held by rd, or an invalid Value if rd is not
// a struct (or a pointer to one).
func structValue(rd dnsv2.RDATA) reflect.Value {
	v := reflect.ValueOf(rd)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Value{}
	}
	return v
}

func isScalarKind(k reflect.Kind) bool {
	switch k {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	}
	return false
}

// formatField formats a scalar field. Strings that are empty or contain
// spaces or quotes are quoted so that each field stays one word.
func formatField(v reflect.Value) string {
	if v.Kind() == reflect.String {
		s := v.String()
		if s == "" || strings.ContainsAny(s, " \t\"") {
			return strconv.Quote(s)
		}
		return s
	}
	return fmt.Sprint(v.Interface())
}
