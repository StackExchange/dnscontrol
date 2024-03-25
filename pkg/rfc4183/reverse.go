package rfc4183

import (
	"fmt"
	"net/netip"
	"strings"
)

// ReverseDomainName implements RFC4183 for turning a CIDR block into
// a in-addr name.  IP addresses are assumed to be /32 or /128 CIDR blocks.
// CIDR host bits are changed to 0s.
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

	// RFC4183 4.1 step 4: The notion of fewer than 8 mask bits is not reasonable.
	if p.Bits() < 8 {
		return "", fmt.Errorf("mask fewer than 8 bits is unreasonable: %s", cidr)
	}

	// Handle IPv6 separately:
	if p.Addr().Is6() {
		return reverseIPv6(p.Addr().AsSlice(), p.Bits())
	}

	// Zero out any host bits.
	p = p.Masked()

	// IPv4: Implement the RFC4183 process:

	// 4.p Step 1
	b := p.Addr().AsSlice()
	x, y, z, w := b[0], b[1], b[2], b[3]
	m := p.Bits()

	if m == 8 {
		return fmt.Sprintf("%d.in-addr.arpa", x), nil
	}
	if m == 16 {
		return fmt.Sprintf("%d.%d.in-addr.arpa", y, x), nil
	}
	if m == 24 {
		return fmt.Sprintf("%d.%d.%d.in-addr.arpa", z, y, x), nil
	}
	if m == 32 {
		return fmt.Sprintf("%d.%d.%d.%d.in-addr.arpa", w, z, y, x), nil
	}

	// 4.1 Step 2
	n := w // I don't understand why the RFC changes variable names at this point, but it does.
	if m >= 24 && m <= 32 {
		return fmt.Sprintf("%d-%d.%d.%d.%d.in-addr.arpa", n, m, z, y, x), nil
	}
	if m >= 16 && m < 24 {
		return fmt.Sprintf("%d-%d.%d.%d.in-addr.arpa", z, m, y, x), nil
	}
	if m >= 8 && m < 16 {
		return fmt.Sprintf("%d-%d.%d.in-addr.arpa", y, m, x), nil
	}
	return "", fmt.Errorf("fewer than 8 mask bits is not reasonable: %v", cidr)

}
