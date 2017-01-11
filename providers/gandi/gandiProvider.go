package gandi

import (
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"

	gandidomain "github.com/prasmussen/gandi-api/domain"
	gandirecord "github.com/prasmussen/gandi-api/domain/zone/record"
)

/*

Gandi API DNS provider:

Info required in `creds.json`:
   - apikey

*/

type GandiApi struct {
	ApiKey      string
	domainIndex map[string]int64 // Map of domainname to index
	nameservers map[string][]*models.Nameserver
	ZoneId      int64
}

type gandiRecord struct {
	gandirecord.RecordInfo
}

func (c *GandiApi) getDomainInfo(domain string) (*gandidomain.DomainInfo, error) {
	if err := c.fetchDomainList(); err != nil {
		return nil, err
	}
	_, ok := c.domainIndex[domain]
	if !ok {
		return nil, fmt.Errorf("%s not listed in zones for gandi account", domain)
	}
	return c.fetchDomainInfo(domain)
}

func (c *GandiApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	domaininfo, err := c.getDomainInfo(domain)
	if err != nil {
		return nil, err
	}
	ns := []*models.Nameserver{}
	for _, nsname := range domaininfo.Nameservers {
		ns = append(ns, &models.Nameserver{Name: nsname})
	}
	return ns, nil
}

func (c *GandiApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	domaininfo, err := c.getDomainInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	foundRecords, err := c.getZoneRecords(domaininfo.ZoneId)
	if err != nil {
		return nil, err
	}

	expectedRecordSets := make([]gandirecord.RecordSet, len(dc.Records))
	for i, rec := range dc.Records {
		if rec.Type == "MX" {
			rec.Target = fmt.Sprintf("%d %s", rec.Priority, rec.Target)
		}
		if rec.Type == "TXT" {
			rec.Target = "\"" + rec.Target + "\"" // FIXME(tlim): Should do proper quoting.
		}
		expectedRecordSets[i] = gandirecord.RecordSet{}
		expectedRecordSets[i]["type"] = rec.Type
		expectedRecordSets[i]["name"] = rec.Name
		expectedRecordSets[i]["value"] = rec.Target
		expectedRecordSets[i]["ttl"] = rec.TTL
	}
	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)

	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	for _, i := range create {
		changes = true
		fmt.Println(i)
	}
	for _, i := range del {
		changes = true
		fmt.Println(i)
	}
	for _, i := range mod {
		changes = true
		fmt.Println(i)
	}

	msg := fmt.Sprintf("GENERATE_ZONE: %s (%d records)", dc.Name, len(dc.Records))
	corrections := []*models.Correction{}
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: msg,
				F: func() error {
					fmt.Printf("CREATING ZONE: %v\n", dc.Name)
					return c.createGandiZone(dc.Name, domaininfo.ZoneId, expectedRecordSets)
				},
			})
	}

	return corrections, nil
}

func newGandi(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &GandiApi{}
	api.ApiKey = m["apikey"]
	if api.ApiKey == "" {
		return nil, fmt.Errorf("Gandi apikey must be provided.")
	}

	return api, nil
}

func init() {
	providers.RegisterDomainServiceProviderType("GANDI", newGandi)
}
