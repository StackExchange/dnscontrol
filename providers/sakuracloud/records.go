package sakuracloud

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
)

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (s *sakuracloudProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	itemMap, err := s.api.GetCommonServiceItemMap()
	if err != nil {
		return nil, err
	}

	item, ok := itemMap[domain]
	if !ok {
		return nil, errNoExist{domain}
	}

	existingRecords := make([]*models.RecordConfig, 0, len(item.Status.NS)+len(item.Settings.DNS.ResourceRecordSets))

	for _, ns := range item.Status.NS {
		// CommonServiceItem.Status.NS fields do not end with a dot.
		// Therefore, a dot is added at the end to make it an absolute domain name.
		//
		//      "Status": {
		//        "Zone": "example.com",
		//        "NS": [
		//          "ns1.gslbN.sakura.ne.jp",
		//          "ns2.gslbN.sakura.ne.jp"
		//        ]
		//      },
		rc := &models.RecordConfig{
			Type:     "NS",
			TTL:      defaultTTL,
			Original: ns,
		}
		rc.SetLabel("@", domain)
		if err := rc.PopulateFromString("NS", ns+".", domain); err != nil {
			return nil, fmt.Errorf("unparsable record received: %w", err)
		}
		existingRecords = append(existingRecords, rc)
	}

	for _, dr := range item.Settings.DNS.ResourceRecordSets {
		rc := toRc(domain, dr)
		existingRecords = append(existingRecords, rc)
	}
	return existingRecords, nil
}

// GetZoneRecordsCorrections gets the records of a zone and returns them in RecordConfig format.
func (s *sakuracloudProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	// The name servers for the Sakura cloud provider cannot be changed.
	// These default TTL is 3600 and the default TTL of DNSControl is 300, so NS corrections can be found.
	// To prevent this, match TTL of DNSControl to one of Sakura Cloud provider.
	for _, rc := range dc.Records {
		if rc.Type == "NS" && rc.Name == "@" {
			rc.TTL = defaultTTL
		}
	}

	msgs, changes, actualChangeCount, err := diff2.ByZone(existing, dc, nil)
	if err != nil {
		return nil, 0, err
	}
	if !changes {
		return nil, actualChangeCount, nil
	}
	msg := strings.Join(msgs, "\n")

	corrections = append(corrections,
		&models.Correction{
			Msg: msg,
			F: func() error {
				drs := make([]domainRecord, 0, len(dc.Records))
				for _, rc := range dc.Records {
					drs = append(drs, toNative(rc))
				}
				return s.api.UpdateZone(dc.Name, drs)
			},
		},
	)

	return corrections, actualChangeCount, nil
}
