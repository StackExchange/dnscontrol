package transform

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/pkg/rfc4183"
)

// ReverseDomainName turns a CIDR block into a reversed (in-addr) name.
func ReverseDomainName(cidr string) (string, error) {

	// Mask missing? Add it.
	if !strings.Contains(cidr, "/") {
		a, err := netip.ParseAddr(cidr)
		if err != nil {
			return "", fmt.Errorf("not an IP address: %w", err)
		}
		if a.Is4() {
			cidr = cidr + "/32"
		} else {
			cidr = cidr + "/128"
		}
	}

	// Parse the CIDR.
	p, err := netip.ParsePrefix(cidr)
	if err != nil {
		return "", fmt.Errorf("not a CIDR block: %w", err)
	}
	bits := p.Bits()

	if p.Masked() != p {
		return "", fmt.Errorf("CIDR %v has 1 bits beyond the mask", cidr)
	}

	// Cases where RFC4183 is the same as RFC2317:
	// IPV6, /0 - /24, /32
	if strings.Contains(cidr, ":") || bits <= 24 || bits == 32 {
		// There is no p.Is6() so we test for ":" as a workaround.
		return rfc4183.ReverseDomainName(cidr)
	}

	// LEGACY CODE

	// Handle IPv4 "Classless in-addr.arpa delegation" RFC2317:
	// if bits >= 25 && bits < 32 {
	// first address / netmask . Class-b-arpa.

	a, c, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}

	base, err := reverseaddr(a)
	if err != nil {
		return "", err
	}

	fparts := strings.Split(c.IP.String(), ".")
	first := fparts[3]
	bits, _ = c.Mask.Size()
	bparts := strings.SplitN(base, ".", 2)
	return fmt.Sprintf("%s/%d.%s", first, bits, bparts[1]), nil
}

// copied from go source.
// https://github.com/golang/go/blob/bfc164c64d33edfaf774b5c29b9bf5648a6447fb/src/net/dnsclient.go#L15

func reverseaddr(ip net.IP) (arpa string, err error) {
	return uitoa(uint(ip[15])) + "." + uitoa(uint(ip[14])) + "." + uitoa(uint(ip[13])) + "." + uitoa(uint(ip[12])) + ".in-addr.arpa", nil
}

// Convert unsigned integer to decimal string.
func uitoa(val uint) string {
	if val == 0 { // avoid string allocation
		return "0"
	}
	var buf [20]byte // big enough for 64bit value base 10
	i := len(buf) - 1
	for val >= 10 {
		q := val / 10
		buf[i] = byte('0' + val - q*10)
		i--
		val = q
	}
	// val < 10
	buf[i] = byte('0' + val)
	return string(buf[i:])
}
