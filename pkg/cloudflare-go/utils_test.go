package cloudflare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testExample struct {
	A string `url:"a,omitempty"`
	C string `url:"c,omitempty"`

	PaginationOptions
}

func Test_buildURI(t *testing.T) {
	tests := map[string]struct {
		path   string
		params interface{}
		want   string
	}{
		"multi level path without params":        {path: "/accounts/foo", params: testExample{}, want: "/accounts/foo"},
		"multi level path with params":           {path: "/zones/foo", params: testExample{A: "b"}, want: "/zones/foo?a=b"},
		"multi level path with multiple params":  {path: "/zones/foo", params: testExample{A: "b", C: "d"}, want: "/zones/foo?a=b&c=d"},
		"multi level path with nested fields":    {path: "/zones/foo", params: testExample{A: "b", C: "d", PaginationOptions: PaginationOptions{PerPage: 10}}, want: "/zones/foo?a=b&c=d&per_page=10"},
		"single level path without params":       {path: "/foo", params: testExample{}, want: "/foo"},
		"single level path with params":          {path: "/bar", params: testExample{C: "d"}, want: "/bar?c=d"},
		"single level path with multiple params": {path: "/foo", params: testExample{A: "b", C: "d"}, want: "/foo?a=b&c=d"},
		"single level path with nested fields":   {path: "/foo", params: testExample{A: "b", C: "d", PaginationOptions: PaginationOptions{PerPage: 10}}, want: "/foo?a=b&c=d&per_page=10"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := buildURI(tc.path, tc.params)
			assert.Equal(t, tc.want, got)
		})
	}
}
