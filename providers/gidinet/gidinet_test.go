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

func TestChunkTXT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"short string unchanged", "hello world", "hello world"},
		{"exactly 250 chars unchanged", string(make([]byte, 250)), string(make([]byte, 250))},
		{"251 chars gets chunked", "A" + string(make([]byte, 250)), `"A` + string(make([]byte, 249)) + `" "` + string(make([]byte, 1)) + `"`},
		{"500 chars splits into two", func() string {
			// Create 500 char string that splits into two 250-char chunks
			s := ""
			for i := 0; i < 500; i++ {
				s += "A"
			}
			return s
		}(), `"` + func() string {
			s := ""
			for i := 0; i < 250; i++ {
				s += "A"
			}
			return s
		}() + `" "` + func() string {
			s := ""
			for i := 0; i < 250; i++ {
				s += "A"
			}
			return s
		}() + `"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := chunkTXT(tt.input)
			if result != tt.expected {
				t.Errorf("chunkTXT() result length = %d, want %d", len(result), len(tt.expected))
			}
		})
	}
}

func TestChunkTXT_Lengths(t *testing.T) {
	// Test that chunking produces correct lengths
	tests := []struct {
		name          string
		inputLen      int
		expectChunked bool
	}{
		{"100 chars not chunked", 100, false},
		{"250 chars not chunked", 250, false},
		{"251 chars chunked", 251, true},
		{"500 chars chunked", 500, true},
		{"1000 chars chunked", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]byte, tt.inputLen)
			for i := range input {
				input[i] = 'A'
			}
			result := chunkTXT(string(input))

			hasQuotes := len(result) > 0 && result[0] == '"'
			if hasQuotes != tt.expectChunked {
				t.Errorf("chunkTXT(%d chars) chunked = %v, want %v", tt.inputLen, hasQuotes, tt.expectChunked)
			}
		})
	}
}

func TestUnchunkTXT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"plain string unchanged", "hello world", "hello world"},
		{"single quoted string", `"hello world"`, "hello world"},
		{"two quoted chunks", `"hello" "world"`, "helloworld"},
		{"three quoted chunks", `"one" "two" "three"`, "onetwothree"},
		{"with extra whitespace", `"hello"   "world"`, "helloworld"},
		{"with tabs", "\"hello\"\t\"world\"", "helloworld"},
		{"DKIM-like value", `"v=DKIM1; k=rsa; p=AAAA" "BBBBCCCC"`, "v=DKIM1; k=rsa; p=AAAABBBBCCCC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := unchunkTXT(tt.input)
			if result != tt.expected {
				t.Errorf("unchunkTXT(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestChunkUnchunkRoundTrip(t *testing.T) {
	// Test that chunk -> unchunk returns the original value
	tests := []struct {
		name  string
		input string
	}{
		{"short string", "v=spf1 include:example.com ~all"},
		{"250 chars", func() string {
			s := ""
			for i := 0; i < 250; i++ {
				s += "X"
			}
			return s
		}()},
		{"500 chars", func() string {
			s := ""
			for i := 0; i < 500; i++ {
				s += "Y"
			}
			return s
		}()},
		{"DKIM key", "v=DKIM1; k=rsa; p=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmiIdsXY9mqIoAj52xijzQXnKU/qoQZUL5T8bitQrCDpPWQSxBlABwoXs33i+VIMVyK4cLSDiIVG5GWZD2JZHzhW65ALcZg+jvLI7Qloa02VkpJPXePjMasnWHXQfSiImVITh7vLrENRDqKZ29H628kkek7hpvRDj4thBAdlKgkBLiUd6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunked := chunkTXT(tt.input)
			result := unchunkTXT(chunked)
			if result != tt.input {
				t.Errorf("Round trip failed: input len=%d, result len=%d", len(tt.input), len(result))
				t.Errorf("Input:  %q", tt.input)
				t.Errorf("Result: %q", result)
			}
		})
	}
}
