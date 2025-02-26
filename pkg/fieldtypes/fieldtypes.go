package fieldtypes

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

// ParseLabel3 returns a short name and FQDN given 3 components: short name, subdomain, and origin.
func ParseLabel3(short, subdomain, origin string) (string, string, error) {

	// Make sure the function is being used correctly:
	if strings.HasSuffix(origin, ".") {
		return "FAIL1", "", fmt.Errorf("originPL3 (%s) is not supposed to end with a dot", origin)
	}
	if strings.ToLower(origin) != origin {
		return "FAIL2", "", fmt.Errorf("origin (%s) must be lowercase", origin)
	}
	if strings.ToLower(subdomain) != subdomain {
		return "FAIL3", "", fmt.Errorf("subdomain (%s) must be lowercase", subdomain)
	}
	if short == "." {
		return "FAIL4", "", fmt.Errorf("label (%s) must not be just a dot", short)

	}

	short = strings.ToLower(short)

	if origin == "" {
		// Legacy mode (no origin specified because parameters doesn't know it yet)

		if lastCharIs(short, '.') {
			return "FAIL5", "", fmt.Errorf("label (%s) can not end in dot in legacy mode", short)
		}

		if subdomain != "" {
			// D_EXTEND() mode (subdomain)
			if short == "" || short == "@" {
				return subdomain, "", nil
			}
			return short + "." + subdomain, "", nil
		} else {
			// D() mode (no subdomain)
			if short == "" || short == "@" {
				return "@", "", nil
			}
			return short, "", nil
		}

	}

	if lastCharIs(short, '.') {
		if short == (origin + ".") {
			return "@", origin, nil
		}
		if strings.HasSuffix(short, "."+origin+".") {
			return short[0 : len(short)-len(origin)-2], short[:len(short)-1], nil
		}
		return "FAIL6", "", fmt.Errorf("short2 (%s) must end with (%s.)", short, origin)
	}

	// Treat *.in-addr.arpa as a FQDN even though it lacks a trailing dot.
	// This is required because REV("1.2.3.4") returns "3.2.1.in-addr.arpa" (no
	// trailing dot), but we want to be able to use it as a label.
	// For example, NS(REV("1.2.3.4), "ns.example.com.").
	if strings.HasSuffix(short, ".in-addr.arpa") || strings.HasSuffix(short, ".ip6.arpa") {
		if strings.HasSuffix(short, "."+origin) {
			r1, r2 := short[0:len(short)-len(origin)-1], short
			return r1, r2, nil
		}
		return "FAIL7", "", fmt.Errorf("shortrev (%s) must end with (%s)", short, origin)
	}

	if subdomain != "" {
		// If D_EXTEND() is in use...
		if short == "" || short == "@" {
			return subdomain, subdomain + "." + origin, nil
		}
		return short + "." + subdomain, short + "." + subdomain + "." + origin, nil
	}

	if short == "" || short == "@" {
		return "@", origin, nil
	}

	return short, short + "." + origin, nil
}

func lastCharIs(s string, c rune) bool {
	return strings.HasSuffix(s, string(c))
}

// ParseHostnameDot is a hostname with a trailing dot.
// FYI: "." is a valid hostname for MX and SRV records. Therefore they are permitted.
// FYI: This calls ToLower on short. After this, we can always assume .target (or whatever) is lowercase.
func ParseHostnameDot(short, subdomain, origin string) (string, error) {

	// Make sure the function is being used correctly:
	if strings.HasSuffix(origin, ".") {
		return "FAIL", fmt.Errorf("originPHD (%s) is not supposed to end with a dot", origin)
	}
	if strings.ToLower(origin) != origin {
		return "FAIL", fmt.Errorf("origin (%s) must be lowercase", origin)
	}
	if strings.ToLower(subdomain) != subdomain {
		return "FAIL", fmt.Errorf("subdomain (%s) must be lowercase", subdomain)
	}
	if short == "" {
		return "FAIL", fmt.Errorf("short must not be empty")
	}
	if strings.HasSuffix(short, "..") {
		return "FAIL", fmt.Errorf("short must not end with '..'")
	}

	short = strings.ToLower(short)

	if lastCharIs(short, '.') {
		return short, nil
	}

	if subdomain != "" {
		// If D_EXTEND() is in use...
		if short == "" || short == "@" {
			return (subdomain + "." + origin + "."), nil
		}
		result := short + "." + subdomain + "." + origin + "."
		return result, nil
	}

	if short == "@" {
		return (origin + "."), nil
	}

	result := short + "." + origin + "."
	return result, nil
}

// ParseHostnameDotNullIsDot is like ParseHostnameDot but returns "." if short is empty.
func ParseHostnameDotNullIsDot(short, subdomain, origin string) (string, error) {
	if short == "" {
		return ".", nil
	}
	return ParseHostnameDot(short, subdomain, origin)
}

// IPv4 is an IPv4 address.
type IPv4 [4]byte

func ParseIPv4(raw string) (IPv4, error) {
	var ip IPv4

	// Is this formatted as a.b.c.d?
	addr, err := netip.ParseAddr(raw)
	if err == nil && addr.Is4() {
		a4 := addr.As4()
		ip[0] = a4[0]
		ip[1] = a4[1]
		ip[2] = a4[2]
		ip[3] = a4[3]

	} else if n, err := strconv.ParseUint(raw, 10, 32); err == nil {
		// Integer-encoded IP address?
		binary.BigEndian.PutUint32(ip[:], uint32(n))
	} else { // No, its an error.
		return ip, fmt.Errorf("not an IPv4 address: %q", raw)
	}
	return ip, nil
}

// MustParseIPv4 is like ParseIPv4 but panics on error. For use in tests and init() functions only.
func MustParseIPv4(raw string) IPv4 {
	ip, err := ParseIPv4(raw)
	if err != nil {
		panic(err)
	}
	return ip
}

func (a *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3])
}
func (a IPv4) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%d.%d.%d.%d"`, a[0], a[1], a[2], a[3])), nil
}
func (a *IPv4) UnmarshalJSON(data []byte) error {
	// Remove the quotes from the JSON string
	str := strings.Trim(string(data), `"`)
	parsedIP, err := ParseIPv4(str)
	if err != nil {
		return err
	}
	*a = parsedIP
	return nil
}

// IPv6 is an IPv6 address.
type IPv6 [16]byte

func ParseIPv6(raw string) (IPv6, error) {
	var ip IPv6

	// Is this formatted as an IPv6 address?
	addr, err := netip.ParseAddr(raw)
	if err == nil && addr.Is6() {
		a6 := addr.As16()
		ip = a6

	} else { // No, its an error.
		return ip, fmt.Errorf("not an IPv6 address: %q", raw)
	}
	return ip, nil
}

func (aaaa *IPv6) String() string {
	addr, _ := netip.AddrFromSlice(aaaa[:])
	return addr.String()
}
func (aaaa IPv6) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf(`"%s"`, aaaa)
	return []byte(s), nil
}
func (aaaa *IPv6) UnmarshalJSON(data []byte) error {
	// Remove the quotes from the JSON string
	str := strings.Trim(string(data), `"`)
	parsedIP, err := ParseIPv6(str)
	if err != nil {
		return err
	}
	*aaaa = parsedIP
	return nil
}

type RedirectCode uint16

func ParseRedirectCode(raw string) (RedirectCode, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil || (nt != 301 && nt != 302) {
		return 0, fmt.Errorf("redirect code is %q, must be 301 or 302", raw)
	}
	return RedirectCode(nt), nil
}

//type StringTrimmed string

func ParseStringTrimmed(raw string) (string, error) {
	return strings.TrimSpace(raw), nil
}

//type StringICTrimmed string (IC = ignore case)

func ParseStringTrimmedAllLower(raw string) (string, error) {
	return strings.TrimSpace(strings.ToLower(raw)), nil
}

//type Uint16 uint16

func ParseUint16(raw string) (uint16, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid uint16: %q", raw)
	}
	return uint16(nt), nil
}

//type Uint8 uint8

func ParseUint8(raw string) (uint8, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid uint8: %q", raw)
	}
	return uint8(nt), nil
}
