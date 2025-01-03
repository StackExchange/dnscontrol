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
		return "", "", fmt.Errorf("origin (%s) is not supposed to end with a dot", origin)
	}
	if strings.ToLower(origin) != origin {
		return "", "", fmt.Errorf("origin (%s) must be lowercase", origin)
	}
	if strings.ToLower(subdomain) != subdomain {
		return "", "", fmt.Errorf("subdomain (%s) must be lowercase", subdomain)
	}
	if short == "." {
		return "", "", fmt.Errorf("label (%s) must not be just a dot", short)

	}

	short = strings.ToLower(short)

	if lastCharIs(short, '.') {
		if short == (origin + ".") {
			return "@", origin, nil
		}
		if strings.HasSuffix(short, "."+origin+".") {
			return short[0 : len(short)-len(origin)-2], short[:len(short)-1], nil
		}
		return "", "", fmt.Errorf("short2 (%s) must end with (%s.)", short, origin)
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

// HostnameDot is a hostname with a trailing dot.

func ParseHostnameDot(short, subdomain, origin string) (string, error) {

	// Make sure the function is being used correctly:
	if strings.HasSuffix(origin, ".") {
		return "", fmt.Errorf("origin (%s) is not supposed to end with a dot", origin)
	}
	if strings.ToLower(origin) != origin {
		return "", fmt.Errorf("origin (%s) must be lowercase", origin)
	}
	if strings.ToLower(subdomain) != subdomain {
		return "", fmt.Errorf("subdomain (%s) must be lowercase", subdomain)
	}
	if short == "" {
		return "", fmt.Errorf("short must not be empty")
	}
	if strings.ToLower(short) != short {
		return "", fmt.Errorf("short (%s) must be lowercase", short)
	}
	if short == "." {
		return "", fmt.Errorf("label (%s) must not be just a dot", short)

	}

	if lastCharIs(short, '.') {
		return short, nil
	}

	if subdomain != "" {
		// If D_EXTEND() is in use...
		if short == "" || short == "@" {
			return (subdomain + "." + origin + "."), nil
		}
		return (short + "." + subdomain + "." + origin + "."), nil
	}

	if short == "@" {
		return (origin + "."), nil
	}

	return (short + "." + origin + "."), nil
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

//type Uint16 uint16

func ParseUint16(raw string) (uint16, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid uint16: %q", raw)
	}
	return uint16(nt), nil
}
