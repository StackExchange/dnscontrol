package powerdns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns/dnsutil"
	"github.com/mittwald/go-powerdns/apis/zones"
	"strings"
)

// toRecordConfig converts a PowerDNS DNSRecord to a RecordConfig. #rtype_variations
func toRecordConfig(domain string, r zones.Record, ttl int, name string, rtype string) (*models.RecordConfig, error) {
	// trimming trailing dot and domain from name
	name = strings.TrimSuffix(name, domain+".")
	name = strings.TrimSuffix(name, ".")

	rc := &models.RecordConfig{
		TTL:      uint32(ttl),
		Original: r,
		Type:     rtype,
	}
	rc.SetLabel(name, domain)

	content := r.Content
	switch rtype {
	case "ALIAS":
		return rc, rc.SetTarget(r.Content)
	case "CNAME", "NS":
		return rc, rc.SetTarget(dnsutil.AddOrigin(content, domain))
	case "CAA":
		return rc, rc.SetTargetCAAString(content)
	case "DS":
		return rc, rc.SetTargetDSString(content)
	case "MX":
		return rc, rc.SetTargetMXString(content)
	case "SRV":
		return rc, rc.SetTargetSRVString(content)
	case "NAPTR":
		return rc, rc.SetTargetNAPTRString(content)
	default:
		return rc, rc.PopulateFromString(rtype, content, domain)
	}
}
