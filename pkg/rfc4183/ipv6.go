package rfc4183

import (
	"fmt"
	"net"
)

// copied from go source.
// https://github.com/golang/go/blob/bfc164c64d33edfaf774b5c29b9bf5648a6447fb/src/net/dnsclient.go#L15

// reverseaddr returns the in-addr.arpa. or ip6.arpa. hostname of the IP
// address addr suitable for rDNS (PTR) record lookup or an error if it fails
// to parse the IP address.
func reverseIPv6(ip net.IP, maskbits int) (arpa string, err error) {
	// Must be IPv6
	if len(ip) != 16 {
		return "", fmt.Errorf("not IPv6 %s/%d", ip.String(), maskbits)
	}

	buf := make([]byte, 0, len(ip)*4+len("ip6.arpa"))
	// Add it, in reverse, to the buffer
	for i := len(ip) - 1; i >= 0; i-- {
		v := ip[i]
		buf = append(buf, hexDigit[v&0xF])
		buf = append(buf, '.')
		buf = append(buf, hexDigit[v>>4])
		buf = append(buf, '.')
	}
	// Append "ip6.arpa." and return (buf already has the final .)
	buf = append(buf, "ip6.arpa"...)
	return string(buf), nil
}

//// Convert unsigned integer to decimal string.
//func uitoa(val uint) string {
//	if val == 0 { // avoid string allocation
//		return "0"
//	}
//	var buf [20]byte // big enough for 64bit value base 10
//	i := len(buf) - 1
//	for val >= 10 {
//		q := val / 10
//		buf[i] = byte('0' + val - q*10)
//		i--
//		val = q
//	}
//	// val < 10
//	buf[i] = byte('0' + val)
//	return string(buf[i:])
//}

const hexDigit = "0123456789abcdef"
