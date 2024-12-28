package rtypectl

import (
	"fmt"
	"net/netip"
	"strconv"
)

func ParseIPv4(raw any) ([4]byte, error) {
	var ip [4]byte
	switch v := raw.(type) {
	case string:
		addr, err := netip.ParseAddr(v)
		if err != nil {
			return ip, err
		}
		if addr.Is4() {
			a4 := addr.As4()
			ip[0] = a4[0]
			ip[1] = a4[1]
			ip[2] = a4[2]
			ip[3] = a4[3]
		} else {
			return ip, fmt.Errorf("not an IPv4 address")
		}

	case float64:
		n := int(v)
		ip[3] = byte(n & 0xff)
		ip[2] = byte(n & 0xff00 >> 8)
		ip[1] = byte(n & 0xff0000 >> 16)
		ip[0] = byte(n & 0xff000000 >> 24)

	default:
		return ip, fmt.Errorf("unsupported type for ipv4 (%T)", raw)
	}
	return ip, nil
}

func ParseLabel(raw any) (string, error) { return ParseString(raw) }

func ParseRedirectCode(raw any) (uint16, error) {
	var n uint16

	switch v := raw.(type) {
	case float64:
		n = uint16(v)
	case string:
		nt, err := strconv.Atoi(v)
		if err != nil {
			return 0, err
		}
		n = uint16(nt)
	default:
		return 0, fmt.Errorf("unsupported type for redirect code (%T)", raw)
	}

	if n == 301 || n == 302 {
		return n, nil
	}
	return 0, fmt.Errorf("invalid redirect code: %q", raw)
}

func ParseString(raw any) (string, error) {
	switch v := raw.(type) {
	case string:
		return v, nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}
