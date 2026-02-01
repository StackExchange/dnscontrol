package providers

import (
	"testing"
)

func TestParseFieldTypeSpec(t *testing.T) {
	tests := []struct {
		name   string
		fields []string
		want   map[string]FieldType
	}{
		{
			name:   "empty slice",
			fields: []string{},
			want:   map[string]FieldType{},
		},
		{
			name:   "single field without type defaults to string",
			fields: []string{"api_key"},
			want: map[string]FieldType{
				"api_key": FieldTypeString,
			},
		},
		{
			name:   "single field with explicit string type",
			fields: []string{"api_key:string"},
			want: map[string]FieldType{
				"api_key": FieldTypeString,
			},
		},
		{
			name:   "single field with bool type",
			fields: []string{"enabled:bool"},
			want: map[string]FieldType{
				"enabled": FieldTypeBool,
			},
		},
		{
			name:   "multiple fields with mixed types",
			fields: []string{"api_key", "secret:string", "debug:bool"},
			want: map[string]FieldType{
				"api_key": FieldTypeString,
				"secret":  FieldTypeString,
				"debug":   FieldTypeBool,
			},
		},
		{
			name:   "field name with underscores",
			fields: []string{"api_secret_key:string"},
			want: map[string]FieldType{
				"api_secret_key": FieldTypeString,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseFieldTypeSpec(tt.fields)
			if len(got) != len(tt.want) {
				t.Errorf("parseFieldTypeSpec() returned map with %d entries, want %d", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if gotV, ok := got[k]; !ok {
					t.Errorf("parseFieldTypeSpec() missing key %q", k)
				} else if gotV != v {
					t.Errorf("parseFieldTypeSpec()[%q] = %v, want %v", k, gotV, v)
				}
			}
		})
	}
}
