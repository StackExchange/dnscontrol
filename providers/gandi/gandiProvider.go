package gandi

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"

	"strings"

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
	dc.Punycode()
	dc.CombineMXs()
	domaininfo, err := c.getDomainInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	foundRecords, err := c.getZoneRecords(domaininfo.ZoneId, dc.Name)
	if err != nil {
		return nil, err
	}

	expectedRecordSets := make([]gandirecord.RecordSet, 0, len(dc.Records))
	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.TTL < 300 {
			log.Printf("WARNING: Gandi does not support ttls < 300. %s will not be set to %d.", rec.NameFQDN, rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			return nil, errors.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
		}
		if rec.Type == "TXT" {
			rec.Target = "\"" + rec.Target + "\"" // FIXME(tlim): Should do proper quoting.
		}
		if rec.Type == "NS" && rec.Name == "@" {
			if !strings.HasSuffix(rec.Target, ".gandi.net.") {
				log.Printf("WARNING: Gandi does not support changing apex NS records. %s will not be added.", rec.Target)
			}
			continue
		}
		rs := gandirecord.RecordSet{
			"type":  rec.Type,
			"name":  rec.Name,
			"value": rec.Target,
			"ttl":   rec.TTL,
		}
		expectedRecordSets = append(expectedRecordSets, rs)
		recordsToKeep = append(recordsToKeep, rec)
	}
	dc.Records = recordsToKeep
	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)

	// Print a list of changes. Generate an actual change that is the zone
	changes := false
	desc := ""
	for _, i := range create {
		changes = true
		desc += "\n" + i.String()
	}
	for _, i := range del {
		changes = true
		desc += "\n" + i.String()
	}
	for _, i := range mod {
		changes = true
		desc += "\n" + i.String()
	}

	msg := fmt.Sprintf("GENERATE_ZONE: %s (%d records)%s", dc.Name, len(dc.Records), desc)
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
	providers.RegisterDomainServiceProviderType("GANDI", newGandi, providers.CanUsePTR)
}
