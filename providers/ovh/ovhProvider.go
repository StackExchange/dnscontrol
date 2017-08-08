package ovh

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/miekg/dns/dnsutil"
	"github.com/xlucas/go-ovh/ovh"
)

type ovhProvider struct {
	client *ovh.Client
	zones  map[string]bool
}

var docNotes = providers.DocumentationNotes{
	providers.DocDualHost:            providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot("New domains require registration"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
}

func newOVH(m map[string]string, metadata json.RawMessage) (*ovhProvider, error) {
	appKey, appSecretKey, consumerKey := m["app-key"], m["app-secret-key"], m["consumer-key"]

	c := ovh.NewClient(ovh.ENDPOINT_EU_OVHCOM, appKey, appSecretKey, consumerKey, false)

	// Check for time lag
	if err := c.PollTimeshift(); err != nil {
		return nil, err
	}

	return &ovhProvider{client: c}, nil
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newOVH(conf, metadata)
}

func init() {
	providers.RegisterDomainServiceProviderType("OVH", newDsp, providers.CanUseSRV, providers.CanUseTLSA, docNotes)
}

func (c *ovhProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if err := c.fetchZones(); err != nil {
		return nil, err
	}
	_, ok := c.zones[domain]
	if !ok {
		return nil, fmt.Errorf("%s not listed in zones for ovh account", domain)
	}

	ns, err := c.fetchNS(domain)
	if err != nil {
		return nil, err
	}

	return models.StringsToNameservers(ns), nil
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in your ovh account", e.domain)
}

func (c *ovhProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	dc.CombineMXs()

	if !c.zones[dc.Name] {
		return nil, errNoExist{dc.Name}
	}

	records, err := c.fetchRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig
	for _, r := range records {
		if r.FieldType == "SOA" {
			continue
		}

		// skip ovh managed apex NS
		if r.FieldType == "NS" && r.SubDomain == "" {
			continue
		}

		if r.SubDomain == "" {
			r.SubDomain = "@"
		}

		// ovh uses a custom type for SPF and DKIM
		if r.FieldType == "SPF" || r.FieldType == "DKIM" {
			r.FieldType = "TXT"
		}

		// ovh default is 3600
		if r.TTL == 0 {
			r.TTL = 3600
		}

		rec := &models.RecordConfig{
			NameFQDN:       dnsutil.AddOrigin(r.SubDomain, dc.Name),
			Name:           r.SubDomain,
			Type:           r.FieldType,
			Target:         r.Target,
			TTL:            uint32(r.TTL),
			CombinedTarget: true,
			Original:       r,
		}
		actual = append(actual, rec)
	}

	// Normalize
	models.Downcase(actual)

	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		// apex NS can't be managed directly
		if rec.Type == "NS" && rec.NameFQDN == dc.Name {
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(actual)

	corrections := []*models.Correction{}

	for _, del := range delete {
		rec := del.Existing.Original.(*Record)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(rec.Id, dc.Name),
		})
	}

	for _, cre := range create {
		rec := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   c.createRecordFunc(rec, dc.Name),
		})
	}

	for _, mod := range modify {
		oldR := mod.Existing.Original.(*Record)
		newR := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(oldR, newR, dc.Name),
		})
	}

	if len(corrections) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: "REFRESH zone " + dc.Name,
			F: func() error {
				return c.refreshZone(dc.Name)
			},
		})
	}

	return corrections, nil
}
