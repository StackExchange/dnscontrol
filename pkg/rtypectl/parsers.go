package rtypectl

import (
	"fmt"
	"net/netip"
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
