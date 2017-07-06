package normalize

import (
	"net"
	"strings"

	"github.com/StackExchange/dnscontrol/pkg/transform"
	"github.com/pkg/errors"
)

func PtrNameMagic(name, domain string) (string, error) {
	// Implement the PTR name magic.  If the name is a properly formed
	// IPv4 or IPv6 address, we replace it with the right string (i.e
	// reverse it and truncate it).

	if strings.HasSuffix(domain, ".in-addr.arpa.") {
		return ptrmagic(name, domain, 4)
	} else if strings.HasSuffix(domain, ".ip6.arpa.") {
		return ptrmagic(name, domain, 16)
	} else {
		return name, nil
	}
}

func ptrmagic(name, domain string, al int) (string, error) {
	ip := net.ParseIP(name)
	if ip == nil || (al == 4 && ip.To4() == nil) || (al == 16 && ip.To16() == nil) {
		// Not a valid IP address, or correct IP version. Leave it alone.
		return name, nil
	}
	rev, err := transform.ReverseDomainName(ip.String() + "/32")
	if err != nil {
		return name, err
	}
	if !strings.HasSuffix(rev, "."+domain) {
		err = errors.Errorf("ERROR: PTR record %v in wrong domain (%v)", name, domain)
	}
	return strings.TrimSuffix(rev, "."+domain), err
}
