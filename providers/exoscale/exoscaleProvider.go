package exoscale

import (
	"encoding/json"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/pkg/errors"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

type exoscaleProvider struct {
	client *egoscale.Client
}

// NewExoscale creates a new Exoscale DNS provider.
func NewExoscale(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	endpoint, apiKey, secretKey := m["dns-endpoint"], m["apikey"], m["secretkey"]

	return &exoscaleProvider{client: egoscale.NewClient(endpoint, apiKey, secretKey)}, nil
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	providers.RegisterDomainServiceProviderType("EXOSCALE", NewExoscale, features)
}

// EnsureDomainExists returns an error if domain doesn't exist.
func (c *exoscaleProvider) EnsureDomainExists(domain string) error {
	_, err := c.client.GetDomain(domain)
	if err != nil {
		_, err := c.client.CreateDomain(domain)
		if err != nil {
			return err
		}
	}
	return err
}

// GetNameservers returns the nameservers for domain.
func (c *exoscaleProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetDomainCorrections returns a list of corretions for the  domain.
func (c *exoscaleProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	records, err := c.client.GetRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0, len(records))
	for _, r := range records {
		if r.RecordType == "SOA" || r.RecordType == "NS" {
			continue
		}
		if r.Name == "" {
			r.Name = "@"
		}
		if r.RecordType == "CNAME" || r.RecordType == "MX" || r.RecordType == "ALIAS" || r.RecordType == "SRV" {
			r.Content += "."
		}
		// exoscale adds these odd txt records that mirror the alias records.
		// they seem to manage them on deletes and things, so we'll just pretend they don't exist
		if r.RecordType == "TXT" && strings.HasPrefix(r.Content, "ALIAS for ") {
			continue
		}
		rec := &models.RecordConfig{
			TTL:      uint32(r.TTL),
			Original: r,
		}
		rec.SetLabel(r.Name, dc.Name)
		switch rtype := r.RecordType; rtype {
		case "ALIAS", "URL":
			rec.Type = r.RecordType
			rec.SetTarget(r.Content)
		case "MX":
			if err := rec.SetTargetMX(uint16(r.Prio), r.Content); err != nil {
				panic(errors.Wrap(err, "unparsable record received from dnsimple"))
			}
		default:
			if err := rec.PopulateFromString(r.RecordType, r.Content, dc.Name); err != nil {
				panic(errors.Wrap(err, "unparsable record received from dnsimple"))
			}
		}
		existingRecords = append(existingRecords, rec)
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, delete, modify := differ.IncrementalDiff(existingRecords)

	var corrections = []*models.Correction{}

	for _, del := range delete {
		rec := del.Existing.Original.(egoscale.DNSRecord)
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
		old := mod.Existing.Original.(egoscale.DNSRecord)
		new := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(&old, new, dc.Name),
		})
	}

	return corrections, nil
}

// Returns a function that can be invoked to create a record in a zone.
func (c *exoscaleProvider) createRecordFunc(rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.client

		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.Target
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record := egoscale.DNSRecord{
			Name:       name,
			RecordType: rc.Type,
			Content:    target,
			TTL:        int(rc.TTL),
			Prio:       int(rc.MxPreference),
		}
		_, err := client.CreateRecord(domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *exoscaleProvider) deleteRecordFunc(recordID int64, domainName string) func() error {
	return func() error {
		client := c.client

		if err := client.DeleteRecord(domainName, recordID); err != nil {
			return err
		}

		return nil

	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *exoscaleProvider) updateRecordFunc(old *egoscale.DNSRecord, rc *models.RecordConfig, domainName string) func() error {
	return func() error {
		client := c.client

		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.Target
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record := egoscale.UpdateDNSRecord{
			Name:       name,
			RecordType: rc.Type,
			Content:    target,
			TTL:        int(rc.TTL),
			Prio:       int(rc.MxPreference),
			ID:         old.ID,
		}

		_, err := client.UpdateRecord(domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}
