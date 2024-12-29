package rtypectl

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

func ParseIPv4(raw string) ([4]byte, error) {
	var ip [4]byte

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

func ParseRedirectCode(raw string) (uint16, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil || (nt != 301 && nt != 302) {
		return 0, fmt.Errorf("redirect code is %q, must be 301 or 302", raw)
	}
	return uint16(nt), nil
}

func ParseStringTrimmed(raw string) (string, error) {
	return strings.TrimSpace(raw), nil
}

func ParseUint16(raw string) (uint16, error) {
	nt, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid uint16: %q", raw)
	}
	return uint16(nt), nil
}

func ParseDottedHost(raw, subdomain, origin string) (string, error) {

	return raw + "." + subdomain + "." + origin, nil
}
