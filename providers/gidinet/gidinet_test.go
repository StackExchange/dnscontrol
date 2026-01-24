package gidinet

import "testing"

func TestFixTTL(t *testing.T) {
	tests := []struct {
		name     string
		given    uint32
		expected uint32
	}{
		{"zero becomes 60", 0, 60},
		{"1 becomes 60", 1, 60},
		{"59 becomes 60", 59, 60},
		{"60 stays 60", 60, 60},
		{"61 becomes 300", 61, 300},
		{"299 becomes 300", 299, 300},
		{"300 stays 300", 300, 300},
		{"301 becomes 600", 301, 600},
		{"3600 stays 3600", 3600, 3600},
		{"3601 becomes 7200", 3601, 7200},
		{"86400 stays 86400", 86400, 86400},
		{"172800 stays 172800", 172800, 172800},
		{"200000 becomes 172800", 200000, 172800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fixTTL(tt.given)
			if result != tt.expected {
				t.Errorf("fixTTL(%d) = %d, want %d", tt.given, result, tt.expected)
			}
		})
	}
}

func TestToFQDN(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		domain   string
		expected string
	}{
		{"@ becomes domain", "@", "example.com", "example.com"},
		{"empty becomes domain", "", "example.com", "example.com"},
		{"www becomes www.domain", "www", "example.com", "www.example.com"},
		{"subdomain becomes subdomain.domain", "sub.www", "example.com", "sub.www.example.com"},
		{"already fqdn stays same", "www.example.com", "example.com", "www.example.com"},
		{"trailing dot removed", "www.", "example.com", "www"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toFQDN(tt.hostname, tt.domain)
			if result != tt.expected {
				t.Errorf("toFQDN(%q, %q) = %q, want %q", tt.hostname, tt.domain, result, tt.expected)
			}
		})
	}
}

func TestFromFQDN(t *testing.T) {
	tests := []struct {
		name     string
		fqdn     string
		domain   string
		expected string
	}{
		{"domain becomes @", "example.com", "example.com", "@"},
		{"domain with dot becomes @", "example.com.", "example.com", "@"},
		{"www.domain becomes www", "www.example.com", "example.com", "www"},
		{"sub.www.domain becomes sub.www", "sub.www.example.com", "example.com", "sub.www"},
		{"unrelated stays same", "other.net", "example.com", "other.net"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fromFQDN(tt.fqdn, tt.domain)
			if result != tt.expected {
				t.Errorf("fromFQDN(%q, %q) = %q, want %q", tt.fqdn, tt.domain, result, tt.expected)
			}
		})
	}
}
