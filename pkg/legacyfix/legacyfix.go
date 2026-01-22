package legacyfix

import (
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/rtype"
	"github.com/miekg/dns"
)

func CopyFromLegacyFields(rec *models.RecordConfig) {
	switch rec.Type {
	case "DS":
		rec.F = &rtype.DS{
			dns.DS{
				KeyTag:     rec.DsKeyTag,
				Algorithm:  rec.DsAlgorithm,
				DigestType: rec.DsDigestType,
				Digest:     rec.DsDigest,
			},
		}
	}
}
