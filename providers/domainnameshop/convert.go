package domainnameshop

import (
	"strconv"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns/dnsutil"
)

func toRecordConfig(domain string, currentRecord *domainNameShopRecord) *models.RecordConfig {
	name := dnsutil.AddOrigin(currentRecord.Host, domain)

	target := currentRecord.Data

	t := &models.RecordConfig{
		Type:         currentRecord.Type,
		TTL:          fixTTL(uint32(currentRecord.TTL)),
		MxPreference: uint16(currentRecord.ActualPriority),
		SrvPriority:  uint16(currentRecord.ActualPriority),
		SrvWeight:    uint16(currentRecord.ActualWeight),
		SrvPort:      uint16(currentRecord.ActualPort),
		Original:     currentRecord,
		CaaTag:       currentRecord.CAATag,
		CaaFlag:      uint8(currentRecord.CAAFlag),
	}

	t.SetTarget(target)
	t.SetLabelFromFQDN(name, domain)

	switch rtype := currentRecord.Type; rtype {
	case "TXT":
		t.SetTargetTXT(target)
	case "CAA":
		if currentRecord.CAATag == "0" {
			t.CaaTag = "issue"
		} else if currentRecord.CAATag == "1" {
			t.CaaTag = "issuewild"
		} else {
			t.CaaTag = "iodef"
		}
	default:
		// nothing additional required
	}
	return t
}

func (api *domainNameShopProvider) fromRecordConfig(domainName string, rc *models.RecordConfig) (*domainNameShopRecord, error) {
	domainID, err := api.getDomainID(domainName)
	if err != nil {
		return nil, err
	}

	data := ""
	if rc.Type == "TXT" {
		data = rc.GetTargetTXTJoined()
	} else {
		data = rc.GetTargetField()
	}

	dnsR := &domainNameShopRecord{
		ID:            0,
		Host:          rc.GetLabel(),
		TTL:           uint16(fixTTL(rc.TTL)),
		Type:          rc.Type,
		Data:          data,
		Weight:        strconv.Itoa(int(rc.SrvWeight)),
		Port:          strconv.Itoa(int(rc.SrvPort)),
		ActualWeight:  rc.SrvWeight,
		ActualPort:    rc.SrvPort,
		CAAFlag:       uint64(int(rc.CaaFlag)),
		ActualCAAFlag: strconv.Itoa(int(rc.CaaFlag)),
		DomainID:      domainID,
	}

	switch rc.Type {
	case "CAA":
		// Actual CAA FLAG
		switch rc.CaaTag {
		case "issue":
			dnsR.CAATag = "0"
		case "issuewild":
			dnsR.CAATag = "1"
		case "iodef":
			dnsR.CAATag = "2"
		}
	case "MX":
		dnsR.Priority = strconv.Itoa(int(rc.MxPreference))
	case "SRV":
		dnsR.Priority = strconv.Itoa(int(rc.SrvPriority))
	default:
		// pass through
	}

	return dnsR, nil
}
