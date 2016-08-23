package gandi

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"

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

type cfRecord struct {
	gandirecord.RecordInfo
}

func (c *cfRecord) GetName() string {
	return c.Name
}

func (c *cfRecord) GetType() string {
	return c.Type
}

func (c *cfRecord) GetTtl() int64 {
	return c.Ttl
}

func (c *cfRecord) GetValue() string {
	return c.Value
}

func (c *cfRecord) GetContent() string {
	switch c.Type {
	case "MX":
		parts := strings.SplitN(c.Value, " ", 2)
		// TODO(tlim): This should check for more errors.
		return strings.Join(parts[1:], " ")
	default:
	}
	return c.Value
}

func (c *cfRecord) GetComparisionData() string {
	if c.Type == "MX" {
		parts := strings.SplitN(c.Value, " ", 2)
		priority, err := strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Sprintf("%s %#v", c.Ttl, parts[0])
		}
		return fmt.Sprintf("%d %d", c.Ttl, priority)
	}
	return fmt.Sprintf("%d", c.Ttl)
}

func (c *GandiApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	if c.domainIndex == nil {
		if err := c.fetchDomainList(); err != nil {
			return nil, err
		}
	}

	_, ok := c.domainIndex[dc.Name]
	if !ok {
		return nil, fmt.Errorf("%s not listed in zones for gandi account", dc.Name)
	}

	domaininfo, err := c.fetchDomainInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	for _, nsname := range domaininfo.Nameservers {
		dc.Nameservers = append(dc.Nameservers, &models.Nameserver{Name: nsname})
	}
	foundRecords, err := c.getZoneRecords(domaininfo.ZoneId)
	if err != nil {
		return nil, err
	}

	// Convert to []diff.Records and compare:
	foundDiffRecords := make([]diff.Record, len(foundRecords))
	for i, rec := range foundRecords {
		n := &cfRecord{}
		n.Id = 0
		n.Name = rec.Name
		n.Ttl = int64(rec.Ttl)
		n.Type = rec.Type
		n.Value = rec.Value
		foundDiffRecords[i] = n
	}
	expectedDiffRecords := make([]diff.Record, len(dc.Records))
	expectedRecordSets := make([]gandirecord.RecordSet, len(dc.Records))
	for i, rec := range dc.Records {
		n := &cfRecord{}
		n.Id = 0
		n.Name = rec.Name
		n.Ttl = int64(rec.TTL)
		if n.Ttl == 0 {
			n.Ttl = 3600
		}
		n.Type = rec.Type
		switch n.Type {
		case "MX":
			n.Value = fmt.Sprintf("%d %s", rec.Priority, rec.Target)
		case "TXT":
			n.Value = "\"" + rec.Target + "\"" // FIXME(tlim): Should do proper quoting.
		default:
			n.Value = rec.Target
		}
		expectedDiffRecords[i] = n
		expectedRecordSets[i] = gandirecord.RecordSet{}
		expectedRecordSets[i]["type"] = n.Type
		expectedRecordSets[i]["name"] = n.Name
		expectedRecordSets[i]["value"] = n.Value
		if n.Ttl != 0 {
			expectedRecordSets[i]["ttl"] = n.Ttl
		}
	}
	_, create, del, mod := diff.IncrementalDiff(foundDiffRecords, expectedDiffRecords)

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

	msg := fmt.Sprintf("GENERATE_ZONE: %s (%d records)", dc.Name, len(expectedDiffRecords))
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
