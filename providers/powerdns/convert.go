package powerdns

import (
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/decode"
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

	switch rtype {
	case "TXT":
		// PowerDNS API accepts long TXTs without requiring to split them
		// The API then returns them as they initially came in, e.g. "averylooooooo[...]oooooongstring" or "string" "string"
		if result, err := decode.QuotedFields(r.Content); err != nil {
			return nil, err
		} else {
			return rc, rc.SetTargetTXTs(result)
		}
	default:
		return rc, rc.PopulateFromString(rtype, r.Content, domain)
	}
}
