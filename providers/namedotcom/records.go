package namedotcom

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/namedotcom/go/namecom"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
)

var defaultNameservers = []*models.Nameserver{
	{Name: "ns1.name.com"},
	{Name: "ns2.name.com"},
	{Name: "ns3.name.com"},
	{Name: "ns4.name.com"},
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (n *namedotcomProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := n.getRecords(domain)
	if err != nil {
		return nil, err
	}

	actual := make([]*models.RecordConfig, len(records))
	for i, r := range records {
		actual[i] = toRecord(r, domain)
	}

	return actual, nil
}

// GetDomainCorrections gathers correctios that would bring n to match dc.
func (n *namedotcomProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	actual, err := n.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
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
	_, create, del, mod, err := differ.IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

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
		if rec.Type == "NS" && rec.GetLabel() == "@" {
			continue // Apex NS records are automatically created for the domain's nameservers and cannot be managed otherwise via the name.com API.
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}

func toRecord(r *namecom.Record, origin string) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      r.TTL,
		Original: r,
	}
	if !strings.HasSuffix(r.Fqdn, ".") {
		panic(fmt.Errorf("namedotcom suddenly changed protocol. Bailing. (%v)", r.Fqdn))
	}
	fqdn := r.Fqdn[:len(r.Fqdn)-1]
	rc.SetLabelFromFQDN(fqdn, origin)
	switch rtype := r.Type; rtype { // #rtype_variations
	case "TXT":
		rc.SetTargetTXTs(decodeTxt(r.Answer))
	case "MX":
		if err := rc.SetTargetMX(uint16(r.Priority), r.Answer); err != nil {
			panic(fmt.Errorf("unparsable MX record received from ndc: %w", err))
		}
	case "SRV":
		if err := rc.SetTargetSRVPriorityString(uint16(r.Priority), r.Answer+"."); err != nil {
			panic(fmt.Errorf("unparsable SRV record received from ndc: %w", err))
		}
	default: // "A", "AAAA", "ANAME", "CNAME", "NS"
		if err := rc.PopulateFromString(rtype, r.Answer, r.Fqdn); err != nil {
			panic(fmt.Errorf("unparsable record received from ndc: %w", err))
		}
	}
	return rc
}

func (n *namedotcomProvider) getRecords(domain string) ([]*namecom.Record, error) {
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

func (n *namedotcomProvider) createRecord(rc *models.RecordConfig, domain string) error {
	record := &namecom.Record{
		DomainName: domain,
		Host:       rc.GetLabel(),
		Type:       rc.Type,
		Answer:     rc.GetTargetField(),
		TTL:        rc.TTL,
		Priority:   uint32(rc.MxPreference),
	}
	switch rc.Type { // #rtype_variations
	case "A", "AAAA", "ANAME", "CNAME", "MX", "NS":
		// nothing
	case "TXT":
		record.Answer = encodeTxt(rc.TxtStrings)
	case "SRV":
		if rc.GetTargetField() == "." {
			return errors.New("SRV records with empty targets are not supported (as of 2019-11-05, the API returns 'Parameter Value Error - Invalid Srv Format')")
		}
		record.Answer = fmt.Sprintf("%d %d %v", rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
		record.Priority = uint32(rc.SrvPriority)
	default:
		panic(fmt.Sprintf("createRecord rtype %v unimplemented", rc.Type))
		// We panic so that we quickly find any switch statements
		// that have not been updated for a new RR type.
	}
	_, err := n.client.CreateRecord(record)
	return err
}

// makeTxt encodes TxtStrings for sending in the CREATE/MODIFY API:
func encodeTxt(txts []string) string {
	ans := txts[0]

	if len(txts) > 1 {
		ans = ""
		for _, t := range txts {
			ans += `"` + strings.Replace(t, `"`, `\"`, -1) + `"`
		}
	}
	return ans
}

// finds a string surrounded by quotes that might contain an escaped quote character.
var quotedStringRegexp = regexp.MustCompile(`"((?:[^"\\]|\\.)*)"`)

// decodeTxt decodes the TXT record as received from name.com and
// returns the list of strings.
func decodeTxt(s string) []string {

	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		txtStrings := []string{}
		for _, t := range quotedStringRegexp.FindAllStringSubmatch(s, -1) {
			txtString := strings.Replace(t[1], `\"`, `"`, -1)
			txtStrings = append(txtStrings, txtString)
		}
		return txtStrings
	}
	return []string{s}
}

func (n *namedotcomProvider) deleteRecord(id int32, domain string) error {
	request := &namecom.DeleteRecordRequest{
		DomainName: domain,
		ID:         id,
	}

	_, err := n.client.DeleteRecord(request)
	return err
}
