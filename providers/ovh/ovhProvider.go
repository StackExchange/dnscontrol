package ovh

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
	"github.com/pkg/errors"
	"github.com/xlucas/go-ovh/ovh"
)

type ovhProvider struct {
	client *ovh.Client
	zones  map[string]bool
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot("New domains require registration"),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
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

func newReg(conf map[string]string) (providers.Registrar, error) {
	return newOVH(conf, nil)
}

func init() {
	providers.RegisterRegistrarType("OVH", newReg)
	providers.RegisterDomainServiceProviderType("OVH", newDsp, features)
}

func (c *ovhProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if err := c.fetchZones(); err != nil {
		return nil, err
	}
	_, ok := c.zones[domain]
	if !ok {
		return nil, errors.Errorf("%s not listed in zones for ovh account", domain)
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
	//dc.CombineMXs()

	if !c.zones[dc.Name] {
		return nil, errNoExist{dc.Name}
	}

	records, err := c.fetchRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	var actual []*models.RecordConfig
	for _, r := range records {
		rec := nativeToRecord(r, dc.Name)
		if rec != nil {
			actual = append(actual, rec)
		}
	}

	// Normalize
	models.PostProcessRecords(actual)

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(actual)

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

func nativeToRecord(r *Record, origin string) *models.RecordConfig {
	if r.FieldType == "SOA" {
		return nil
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
		panic(errors.Wrap(err, "unparsable record received from ovh"))
	}

	// ovh default is 3600
	if rec.TTL == 0 {
		rec.TTL = 3600
	}

	return rec
}

func (c *ovhProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {

	ns, err := c.fetchRegistrarNS(dc.Name)
	if err != nil {
		return nil, err
	}

	sort.Strings(ns)
	found := strings.Join(ns, ",")

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
				F: func() error {
					err := c.updateNS(dc.Name, desiredNs)
					if err != nil {
						return err
					}
					return nil
				}},
		}, nil
	}
	return nil, nil
}
