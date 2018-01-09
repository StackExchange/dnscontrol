package namedotcom

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/namedotcom/go/namecom"

	"github.com/miekg/dns/dnsutil"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

var defaultNameservers = []*models.Nameserver{
	{Name: "ns1.name.com"},
	{Name: "ns2.name.com"},
	{Name: "ns3.name.com"},
	{Name: "ns4.name.com"},
}

func (n *NameCom) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	records, err := n.getRecords(dc.Name)
	if err != nil {
		return nil, err
	}
	actual := make([]*models.RecordConfig, len(records))
	for i, r := range records {
		actual[i] = toRecord(r)
	}

	for _, rec := range dc.Records {
		if rec.Type == "ALIAS" {
			rec.Type = "ANAME"
		}
	}

	checkNSModifications(dc)

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(actual)
	corrections := []*models.Correction{}

	for _, d := range del {
		rec := d.Existing.Original.(*namecom.Record)
		c := &models.Correction{Msg: d.String(), F: func() error { return n.deleteRecord(rec.ID, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, cre := range create {
		rec := cre.Desired
		c := &models.Correction{Msg: cre.String(), F: func() error { return n.createRecord(rec, dc.Name) }}
		corrections = append(corrections, c)
	}
	for _, chng := range mod {
		old := chng.Existing.Original.(*namecom.Record)
		new := chng.Desired
		c := &models.Correction{Msg: chng.String(), F: func() error {
			err := n.deleteRecord(old.ID, dc.Name)
			if err != nil {
				return err
			}
			return n.createRecord(new, dc.Name)
		}}
		corrections = append(corrections, c)
	}
	return corrections, nil
}

func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.NameFQDN == dc.Name {
			continue // Apex NS records are automatically created for the domain's nameservers and cannot be managed otherwise via the name.com API.
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

func toRecord(r *namecom.Record) *models.RecordConfig {
	rc := &models.RecordConfig{
		NameFQDN: r.Fqdn,
		Type:     r.Type,
		Target:   r.Answer,
		TTL:      r.TTL,
		Original: r,
	}
	switch r.Type { // #rtype_variations
	case "A", "AAAA", "ANAME", "CNAME", "NS", "TXT":
		// nothing additional.
	case "MX":
		rc.MxPreference = uint16(r.Priority)
	case "SRV":
		parts := strings.Split(r.Answer, " ")
		weight, _ := strconv.ParseInt(parts[0], 10, 32)
		port, _ := strconv.ParseInt(parts[1], 10, 32)
		rc.SrvWeight = uint16(weight)
		rc.SrvPort = uint16(port)
		rc.SrvPriority = uint16(r.Priority)
		rc.MxPreference = 0
		rc.Target = parts[2] + "."
	default:
		panic(fmt.Sprintf("toRecord unimplemented rtype %v", r.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	return rc
}

func (n *NameCom) getRecords(domain string) ([]*namecom.Record, error) {
	var (
		err      error
		records  []*namecom.Record
		response *namecom.ListRecordsResponse
	)

	request := &namecom.ListRecordsRequest{
		DomainName: domain,
		Page:       1,
	}

	for request.Page > 0 {
		response, err = n.client.ListRecords(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Records...)
		request.Page = response.NextPage
	}

	for _, rc := range records {
		if rc.Type == "CNAME" || rc.Type == "ANAME" || rc.Type == "MX" || rc.Type == "NS" {
			rc.Answer = rc.Answer + "."
		}
	}
	return records, nil
}

func (n *NameCom) createRecord(rc *models.RecordConfig, domain string) error {
	record := &namecom.Record{
		DomainName: domain,
		Host:       dnsutil.TrimDomainName(rc.NameFQDN, domain),
		Type:       rc.Type,
		Answer:     rc.Target,
		TTL:        rc.TTL,
		Priority:   uint32(rc.MxPreference),
	}

	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "ANAME", "CNAME", "MX", "NS", "TXT":
		// nothing
	case "SRV":
		record.Answer = fmt.Sprintf("%d %d %v", rc.SrvWeight, rc.SrvPort, rc.Target)
		record.Priority = uint32(rc.SrvPriority)
	default:
		panic(fmt.Sprintf("createRecord rtype %v unimplemented", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	_, err := n.client.CreateRecord(record)
	return err
}

func (n *NameCom) deleteRecord(id int32, domain string) error {
	request := &namecom.DeleteRecordRequest{
		DomainName: domain,
		ID:         id,
	}

	_, err := n.client.DeleteRecord(request)
	return err
}
