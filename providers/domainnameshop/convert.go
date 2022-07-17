package domainnameshop

import (
	"strconv"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/miekg/dns/dnsutil"
)

func toRecordConfig(domain string, currentRecords *domainNameShopRecord) *models.RecordConfig {
	name := dnsutil.AddOrigin(currentRecords.Host, domain)

	target := currentRecords.Data

	t := &models.RecordConfig{
		Type:         currentRecords.Type,
		TTL:          fixTTL(uint32(currentRecords.TTL)),
		MxPreference: uint16(currentRecords.ActualPriority),
		SrvPriority:  uint16(currentRecords.ActualPriority),
		SrvWeight:    uint16(currentRecords.ActualWeight),
		SrvPort:      uint16(currentRecords.ActualPort),
		Original:     currentRecords,
		CaaTag:       currentRecords.CAATag,
		CaaFlag:      uint8(currentRecords.CAAFlag),
	}

	t.SetTarget(target)
	t.SetLabelFromFQDN(name, domain)

	switch rtype := currentRecords.Type; rtype {
	case "TXT":
		t.SetTargetTXT(target)
	case "CAA":
		if currentRecords.CAATag == "0" {
			t.CaaTag = "issue"
		} else if currentRecords.CAATag == "1" {
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
		Priority:      strconv.Itoa(int(rc.MxPreference) + int(rc.SrvPriority)),
		Weight:        strconv.Itoa(int(rc.SrvWeight)),
		Port:          strconv.Itoa(int(rc.SrvPort)),
		ActualWeight:  rc.SrvWeight,
		ActualPort:    rc.SrvPort,
		CAAFlag:       uint64(int(rc.CaaFlag)),
		ActualCAAFlag: strconv.Itoa(int(rc.CaaFlag)),
		DomainID:      domainID,
	}

	if rc.Type == "CAA" {
		// Actual CAA FLAG
		switch rc.CaaTag {
		case "issue":
			dnsR.CAATag = "0"
		case "issuewild":
			dnsR.CAATag = "1"
		case "iodef":
			dnsR.CAATag = "2"
		}
	}

	return dnsR, nil
}
