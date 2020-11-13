package ovh

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/ovh/go-ovh/ovh"
)

type ovhProvider struct {
	client *ovh.Client
	zones  map[string]bool
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.DocCreateDomains:       providers.Cannot("New domains require registration"),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
}

func newOVH(m map[string]string, metadata json.RawMessage) (*ovhProvider, error) {
	appKey, appSecretKey, consumerKey := m["app-key"], m["app-secret-key"], m["consumer-key"]

	c, err := ovh.NewClient(ovh.OvhEU, appKey, appSecretKey, consumerKey)
	if c == nil {
		return nil, err
	}

	ovh := &ovhProvider{client: c}
	if err := ovh.fetchZones(); err != nil {
		return nil, err
	}
	return ovh, nil
}

func newDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newOVH(conf, metadata)
}

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newOVH(conf, nil)
}

func init() {
	providers.RegisterRegistrarType("OVH", newReg)
	providers.RegisterDomainServiceProviderType("OVH", newDsp, features)
}

func (c *ovhProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	_, ok := c.zones[domain]
	if !ok {
		return nil, fmt.Errorf("'%s' not a zone in ovh account", domain)
	}

	ns, err := c.fetchRegistrarNS(domain)
	if err != nil {
		return nil, err
	}

	return models.ToNameservers(ns)
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in your ovh account", e.domain)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *ovhProvider) GetZoneRecords(domain string) (models.Records, error) {
	if !c.zones[domain] {
		return nil, errNoExist{domain}
	}

	records, err := c.fetchRecords(domain)
	if err != nil {
		return nil, err
	}

	var actual models.Records
	for _, r := range records {
		rec, err := nativeToRecord(r, domain)
		if err != nil {
			return nil, err
		}
		if rec != nil {
			actual = append(actual, rec)
		}
	}
	return actual, nil
}

func (c *ovhProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()
	//dc.CombineMXs()

	actual, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(actual)
	if err != nil {
		return nil, err
	}

	corrections := []*models.Correction{}

	for _, del := range delete {
		rec := del.Existing.Original.(*Record)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(rec.ID, dc.Name),
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

func nativeToRecord(r *Record, origin string) (*models.RecordConfig, error) {
	if r.FieldType == "SOA" {
		return nil, nil
	}
	rec := &models.RecordConfig{
		TTL:      uint32(r.TTL),
		Original: r,
	}

	rtype := r.FieldType

	// ovh uses a custom type for SPF and DKIM
	if rtype == "SPF" || rtype == "DKIM" {
		rtype = "TXT"
	}

	rec.SetLabel(r.SubDomain, origin)
	if err := rec.PopulateFromString(rtype, r.Target, origin); err != nil {
		return nil, fmt.Errorf("unparsable record received from ovh: %w", err)
	}

	// ovh default is 3600
	if rec.TTL == 0 {
		rec.TTL = 3600
	}

	return rec, nil
}

func (c *ovhProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	// get the actual in-use nameservers
	actualNs, err := c.fetchRegistrarNS(dc.Name)
	if err != nil {
		return nil, err
	}

	// get the actual used ones + the configured one through dnscontrol
	expectedNs := []string{}
	for _, d := range dc.Nameservers {
		expectedNs = append(expectedNs, d.Name)
	}

	sort.Strings(actualNs)
	actual := strings.Join(actualNs, ",")

	sort.Strings(expectedNs)
	expected := strings.Join(expectedNs, ",")

	// check if we need to change something
	if actual != expected {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Change Nameservers from '%s' to '%s'", actual, expected),
				F: func() error {
					err := c.updateNS(dc.Name, expectedNs)
					if err != nil {
						return err
					}
					return nil
				}},
		}, nil
	}

	return nil, nil
}
