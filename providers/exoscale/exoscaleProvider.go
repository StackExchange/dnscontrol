package exoscale

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/exoscale/egoscale"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
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
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported"),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("Exoscale does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Unimplemented(),
}

func init() {
	providers.RegisterDomainServiceProviderType("EXOSCALE", NewExoscale, features)
}

// EnsureDomainExists returns an error if domain doesn't exist.
func (c *exoscaleProvider) EnsureDomainExists(domain string) error {
	ctx := context.Background()
	_, err := c.client.GetDomain(ctx, domain)
	if err != nil {
		_, err := c.client.CreateDomain(ctx, domain)
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

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *exoscaleProvider) GetZoneRecords(domain string) (models.Records, error) {
	return nil, fmt.Errorf("not implemented")
	// This enables the get-zones subcommand.
	// Implement this by extracting the code from GetDomainCorrections into
	// a single function.  For most providers this should be relatively easy.
}

// GetDomainCorrections returns a list of corretions for the  domain.
func (c *exoscaleProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	ctx := context.Background()
	records, err := c.client.GetRecords(ctx, dc.Name)
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
				return nil, fmt.Errorf("unparsable record received from exoscale: %w", err)
			}
		default:
			if err := rec.PopulateFromString(r.RecordType, r.Content, dc.Name); err != nil {
				return nil, fmt.Errorf("unparsable record received from exoscale: %w", err)
			}
		}
		existingRecords = append(existingRecords, rec)
	}
	removeOtherNS(dc)

	// Normalize
	models.PostProcessRecords(existingRecords)

	differ := diff.New(dc)
	_, create, delete, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

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
		ctx := context.Background()
		_, err := client.CreateRecord(ctx, domainName, record)
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

		ctx := context.Background()
		if err := client.DeleteRecord(ctx, domainName, recordID); err != nil {
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

		ctx := context.Background()
		_, err := client.UpdateRecord(ctx, domainName, record)
		if err != nil {
			return err
		}

		return nil
	}
}

func defaultNSSUffix(defNS string) bool {
	return (strings.HasSuffix(defNS, ".exoscale.io.") ||
		strings.HasSuffix(defNS, ".exoscale.com.") ||
		strings.HasSuffix(defNS, ".exoscale.ch.") ||
		strings.HasSuffix(defNS, ".exoscale.net."))
}

// remove all non-exoscale NS records from our desired state.
// if any are found, print a warning
func removeOtherNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" {
			// apex NS inside exoscale are expected.
			if rec.GetLabelFQDN() == dc.Name && defaultNSSUffix(rec.GetTargetField()) {
				continue
			}
			fmt.Printf("Warning: exoscale.com(.io, .ch, .net) does not allow NS records to be modified. %s will not be added.\n", rec.GetTargetField())
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
