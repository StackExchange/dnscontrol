package gandi

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	gandidomain "github.com/prasmussen/gandi-api/domain"
	gandirecord "github.com/prasmussen/gandi-api/domain/zone/record"

	"github.com/StackExchange/dnscontrol/v2/models"
	"github.com/StackExchange/dnscontrol/v2/pkg/printer"
	"github.com/StackExchange/dnscontrol/v2/providers"
	"github.com/StackExchange/dnscontrol/v2/providers/diff"
)

/*

Gandi API DNS provider:

Info required in `creds.json`:
   - apikey

*/

var deprecationWarned bool

var features = providers.DocumentationNotes{
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("Can only manage domains registered through their service"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	providers.RegisterDomainServiceProviderType("GANDI", newDsp, features)
	providers.RegisterRegistrarType("GANDI", newReg)
}

// GandiApi is the API handle for this module.
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
		return nil, fmt.Errorf("'%s' not a zone in gandi account", domain)
	}
	return c.fetchDomainInfo(domain)
}

// GetNameservers returns the nameservers for domain.
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

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (client *GandiApi) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns a list of corrections recommended for this domain.
func (c *GandiApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
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
			printer.Warnf("Gandi does not support ttls < 300. Setting %s from %d to 300\n", rec.GetLabelFQDN(), rec.TTL)
			rec.TTL = 300
		}
		if rec.TTL > 2592000 {
			return nil, fmt.Errorf("ERROR: Gandi does not support TTLs > 30 days (TTL=%d)", rec.TTL)
		}
		if rec.Type == "TXT" {
			rec.SetTarget("\"" + rec.GetTargetField() + "\"") // FIXME(tlim): Should do proper quoting.
		}
		if rec.Type == "NS" && rec.GetLabel() == "@" {
			if !strings.HasSuffix(rec.GetTargetField(), ".gandi.net.") {
				printer.Warnf("Gandi does not support changing apex NS records. %s will not be added.\n", rec.GetTargetField())
			}
			continue
		}
		rs := gandirecord.RecordSet{
			"type":  rec.Type,
			"name":  rec.GetLabel(),
			"value": rec.GetTargetCombined(),
			"ttl":   rec.TTL,
		}
		expectedRecordSets = append(expectedRecordSets, rs)
		recordsToKeep = append(recordsToKeep, rec)
	}
	dc.Records = recordsToKeep

	// Normalize
	models.PostProcessRecords(foundRecords)

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
					printer.Printf("CREATING ZONE: %v\n", dc.Name)
					return c.createGandiZone(dc.Name, domaininfo.ZoneId, expectedRecordSets)
				},
			})
	}

	return corrections, nil
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newGandi(conf, metadata)
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newGandi(conf, nil)
}

func newGandi(m map[string]string, metadata json.RawMessage) (*GandiApi, error) {
	if !deprecationWarned {
		deprecationWarned = true
		fmt.Printf("WARNING: GANDI is deprecated and will disappear in 3.0. Please migrate to GANDI_V5.\n")
	}
	api := &GandiApi{}
	api.ApiKey = m["apikey"]
	if api.ApiKey == "" {
		return nil, fmt.Errorf("missing Gandi apikey")
	}

	return api, nil
}

// GetRegistrarCorrections returns a list of corrections for this registrar.
func (c *GandiApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	domaininfo, err := c.getDomainInfo(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(domaininfo.Nameservers)
	found := strings.Join(domaininfo.Nameservers, ",")
	desiredNs := []string{}
	for _, d := range dc.Nameservers {
		desiredNs = append(desiredNs, d.Name)
	}
	sort.Strings(desiredNs)
	desired := strings.Join(desiredNs, ",")
	if found != desired {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", found, desired),
				F: func() (err error) {
					_, err = c.setDomainNameservers(dc.Name, desiredNs)
					return
				}},
		}, nil
	}
	return nil, nil
}
