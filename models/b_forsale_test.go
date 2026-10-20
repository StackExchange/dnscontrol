package models

import (
	"strings"
	"testing"
)

// The options object arrives from dnsconfig.js as JSON (see b_m365_test.go), so
// lists are []any.

func TestBuilderFORSALE(t *testing.T) {
	tests := []struct {
		name          string
		zone          string
		subdomain     string
		wantSubDomain string
		args          []any
		want          []string
	}{
		{
			name: "no options",
			zone: "example.com",
			args: []any{"@"},
			want: []string{`_for-sale 300 IN TXT "v=FORSALE1;"`},
		},
		{
			name: "empty options",
			zone: "example.com",
			args: []any{"@", map[string]any{}},
			want: []string{`_for-sale 300 IN TXT "v=FORSALE1;"`},
		},
		{
			// RFC 10023, Appendix A.5, in the builder's order.
			name: "every content tag",
			zone: "example.com",
			args: []any{"@", map[string]any{
				"furi": "https://fs.example.com/",
				"ftxt": "This domain name is for sale",
				"fval": "EUR500",
				"fcod": []any{"EXCO-ZGVhZGJlZWYx", "XYZ1-MTExLTIyMi0zMzMtNDQ0"},
			}},
			want: []string{
				`_for-sale 300 IN TXT "v=FORSALE1;fcod=EXCO-ZGVhZGJlZWYx"`,
				`_for-sale 300 IN TXT "v=FORSALE1;fcod=XYZ1-MTExLTIyMi0zMzMtNDQ0"`,
				`_for-sale 300 IN TXT "v=FORSALE1;ftxt=This domain name is for sale"`,
				`_for-sale 300 IN TXT "v=FORSALE1;furi=https://fs.example.com/"`,
				`_for-sale 300 IN TXT "v=FORSALE1;fval=EUR500"`,
			},
		},
		{
			// RFC 10023, Appendix A.3 and A.4.
			name: "URI schemes and prices",
			zone: "example.com",
			args: []any{"@", map[string]any{
				"furi": []any{"mailto:hq@example.com?subject=foo", "tel:+1-201-555-0123", "https://example.com/foo%20bar"},
				"fval": []any{"BTC0.000010", "USD750"},
			}},
			want: []string{
				`_for-sale 300 IN TXT "v=FORSALE1;furi=mailto:hq@example.com?subject=foo"`,
				`_for-sale 300 IN TXT "v=FORSALE1;furi=tel:+1-201-555-0123"`,
				`_for-sale 300 IN TXT "v=FORSALE1;furi=https://example.com/foo%20bar"`,
				`_for-sale 300 IN TXT "v=FORSALE1;fval=BTC0.000010"`,
				`_for-sale 300 IN TXT "v=FORSALE1;fval=USD750"`,
			},
		},
		{
			// The value is published as written, including characters that
			// need quoting in a zone file.
			name: "free text is not altered",
			zone: "example.com",
			args: []any{"@", map[string]any{
				"ftxt": ` Only "$99" at ExCo; call `,
			}},
			want: []string{`_for-sale 300 IN TXT "v=FORSALE1;ftxt= Only \"$99\" at ExCo; call "`},
		},
		{
			name: "below a label",
			zone: "example.com",
			args: []any{"exco", map[string]any{"furi": "https://example.com/buy"}},
			want: []string{`_for-sale.exco 300 IN TXT "v=FORSALE1;furi=https://example.com/buy"`},
		},
		{
			name:          "D_EXTEND",
			zone:          "example.com",
			subdomain:     "SUB",
			wantSubDomain: "sub",
			args:          []any{"@", map[string]any{"fval": "EUR999"}},
			want:          []string{`_for-sale.sub 300 IN TXT "v=FORSALE1;fval=EUR999"`},
		},
		{
			name:          "D_EXTEND, below a label",
			zone:          "example.com",
			subdomain:     "sub",
			wantSubDomain: "sub",
			args:          []any{"www", map[string]any{"fval": "EUR999"}},
			want:          []string{`_for-sale.www.sub 300 IN TXT "v=FORSALE1;fval=EUR999"`},
		},
		{
			// 255 octets in total: the longest value that fits in one
			// character-string.
			name: "longest value",
			zone: "example.com",
			args: []any{"@", map[string]any{"fcod": strings.Repeat("a", 239)}},
			want: []string{`_for-sale 300 IN TXT "v=FORSALE1;fcod=` + strings.Repeat("a", 239) + `"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := BuilderFORSALE(MustNewDomainConfig(tt.zone), 300, tt.args, tt.subdomain)
			if err != nil {
				t.Fatalf("BuilderFORSALE() error = %v, want none", err)
			}

			var got []string
			for _, rec := range records {
				got = append(got, rec.LineString())
			}
			if len(got) != len(tt.want) {
				t.Fatalf("BuilderFORSALE() created %d records, want %d:\ngot:  %q\nwant: %q", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("record %d = %q, want %q", i, got[i], tt.want[i])
				}
				if records[i].SubDomain != tt.wantSubDomain {
					t.Errorf("record %d SubDomain = %q, want %q", i, records[i].SubDomain, tt.wantSubDomain)
				}
				// Section 2.4: one character-string per record.
				if n := len(records[i].GetTargetTXTSegmented()); n != 1 {
					t.Errorf("record %d has %d TXT strings, want 1", i, n)
				}
			}
		})
	}
}

func TestBuilderFORSALEErrors(t *testing.T) {
	tests := []struct {
		name    string
		zone    string
		args    []any
		wantErr string
	}{
		{
			name:    "no label",
			zone:    "example.com",
			args:    []any{},
			wantErr: "the first argument must be the label",
		},
		{
			name:    "options object instead of the label",
			zone:    "example.com",
			args:    []any{map[string]any{"furi": "https://example.com/"}},
			wantErr: "the first argument must be the label",
		},
		{
			name:    "options are not an object",
			zone:    "example.com",
			args:    []any{"@", "https://example.com/"},
			wantErr: "the second argument must be an options object, got string",
		},
		{
			name:    "too many arguments",
			zone:    "example.com",
			args:    []any{"@", map[string]any{}, map[string]any{}},
			wantErr: "got 3 arguments",
		},
		{
			name:    "unknown option",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"uri": "https://example.com/"}},
			wantErr: `unknown option "uri"`,
		},
		{
			name:    "option is a number",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fval": float64(999)}},
			wantErr: `option "fval" must be a string or a list of strings, got number`,
		},
		{
			name:    "list contains a number",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fcod": []any{"A", float64(1)}}},
			wantErr: "found number in the list",
		},
		{
			name:    "empty list",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"ftxt": []any{}}},
			wantErr: `option "ftxt" must not be an empty list`,
		},
		{
			name:    "empty value",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fcod": ""}},
			wantErr: `option "fcod" must not be empty`,
		},
		{
			name:    "value too long",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"ftxt": strings.Repeat("a", 240)}},
			wantErr: "is 240 octets long; RFC 10023 allows at most 239",
		},
		{
			// Octets, not characters: 120 two-octet characters.
			name:    "value too long in octets",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"ftxt": strings.Repeat("é", 120)}},
			wantErr: "is 240 octets long",
		},
		{
			name:    "duplicate value",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fcod": []any{"EXCO-1", "EXCO-1"}}},
			wantErr: `lists "EXCO-1" twice`,
		},
		{
			name:    "relative URI",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"furi": "example.com/buy"}},
			wantErr: `option "furi" must be an absolute URI`,
		},
		{
			name:    "URI with a space",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"furi": "https://example.com/foo bar"}},
			wantErr: "percent-encode spaces",
		},
		{
			name:    "two URIs",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"furi": "https://a.example/ https://b.example/"}},
			wantErr: "must be a single URI",
		},
		{
			name:    "price in lower case",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fval": "eur999"}},
			wantErr: `option "fval" must be an upper case currency code`,
		},
		{
			name:    "price with a separator",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fval": "USD1,000"}},
			wantErr: `option "fval" must be an upper case currency code`,
		},
		{
			name:    "price without an amount",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fval": "EUR"}},
			wantErr: `option "fval" must be an upper case currency code`,
		},
		{
			name:    "price with a trailing decimal point",
			zone:    "example.com",
			args:    []any{"@", map[string]any{"fval": "EUR5."}},
			wantErr: `option "fval" must be an upper case currency code`,
		},
		{
			name:    "below .arpa",
			zone:    "2.0.192.in-addr.arpa",
			args:    []any{"@"},
			wantErr: "does not apply to names below .arpa",
		},
		{
			name:    "wildcard",
			zone:    "example.com",
			args:    []any{"*"},
			wantErr: "a wildcard cannot be put up for sale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			records, err := BuilderFORSALE(MustNewDomainConfig(tt.zone), 300, tt.args, "")
			if err == nil {
				t.Fatalf("BuilderFORSALE() = %d records, want an error containing %q", len(records), tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("BuilderFORSALE() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
