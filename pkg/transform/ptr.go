package transform

import (
	"net"
	"strings"

	"github.com/pkg/errors"
)

func PtrNameMagic(name, domain string) (string, error) {
	// Implement the PTR name magic.  If the name is a properly formed
	// IPv4 or IPv6 address, we replace it with the right string (i.e
	// reverse it and truncate it).

	// If the name is already in-addr.arpa or ipv6.arpa,
	// make sure the domain matches.
	if strings.HasSuffix(name, ".in-addr.arpa.") || strings.HasSuffix(name, ".ip6.arpa.") {
		if strings.HasSuffix(name, "."+domain+".") {
			return strings.TrimSuffix(name, "."+domain+"."), nil
		} else {
			return name, errors.Errorf("PTR record %v in wrong domain (%v)", name, domain)
		}
	}

	// If the domain is .arpa, we do magic.
	if strings.HasSuffix(domain, ".in-addr.arpa") {
		return ipv4magic(name, domain)
	} else if strings.HasSuffix(domain, ".ip6.arpa") {
		return ipv6magic(name, domain)
	} else {
		return name, nil
	}
}

func ipv4magic(name, domain string) (string, error) {
	// Not a valid IPv4 address. Leave it alone.
	ip := net.ParseIP(name)
	if ip == nil || ip.To4() == nil || !strings.Contains(name, ".") {
		return name, nil
	}

	// Reverse it.
	rev, err := ReverseDomainName(ip.String() + "/32")
	if err != nil {
		return name, err
	}
	if !strings.HasSuffix(rev, "."+domain) {
		err = errors.Errorf("ERROR: PTR record %v in wrong IPv4 domain (%v)", name, domain)
	}
	return strings.TrimSuffix(rev, "."+domain), err
}

func ipv6magic(name, domain string) (string, error) {
	// Not a valid IPv6 address. Leave it alone.
	ip := net.ParseIP(name)
	if ip == nil || len(ip) != 16 || !strings.Contains(name, ":") {
		return name, nil
	}

	// Reverse it.
	rev, err := ReverseDomainName(ip.String() + "/128")
	if err != nil {
		return name, err
	}
	if !strings.HasSuffix(rev, "."+domain) {
		err = errors.Errorf("ERROR: PTR record %v in wrong IPv6 domain (%v)", name, domain)
	}
	return strings.TrimSuffix(rev, "."+domain), err
}
