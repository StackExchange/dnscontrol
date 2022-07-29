package exoscale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	egoscale "github.com/exoscale/egoscale/v2"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

const (
	defaultAPIZone = "ch-gva-2"
)

var ErrDomainNotFound = errors.New("domain not found")

type exoscaleProvider struct {
	client  *egoscale.Client
	apiZone string
}

// NewExoscale creates a new Exoscale DNS provider.
func NewExoscale(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	endpoint, apiKey, secretKey := m["dns-endpoint"], m["apikey"], m["secretkey"]

	client, err := egoscale.NewClient(
		apiKey,
		secretKey,
		egoscale.ClientOptWithAPIEndpoint(endpoint),
	)
	if err != nil {
		return nil, err
	}

	provider := exoscaleProvider{
		client:  client,
		apiZone: defaultAPIZone,
	}

	if z, ok := m["apizone"]; ok {
		provider.apiZone = z
	}

	return &provider, nil
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Unimplemented(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported"),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("Exoscale does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewExoscale,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("EXOSCALE", fns, features)
}

// EnsureDomainExists adds a domain if it is not managed by Exoscale.
func (c *exoscaleProvider) EnsureDomainExists(domainName string) error {
	ctx := context.Background()
	_, err := c.findDomainByName(domainName)
	if errors.Is(err, ErrDomainNotFound) {
		_, err = c.client.CreateDNSDomain(ctx, c.apiZone, &egoscale.DNSDomain{UnicodeName: &domainName})
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

	domain, err := c.findDomainByName(dc.Name)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	records, err := c.client.ListDNSDomainRecords(ctx, c.apiZone, *domain.ID)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0, len(records))
	for _, r := range records {
		record, err := c.client.GetDNSDomainRecord(ctx, c.apiZone, *domain.ID, *r.ID)
		if err != nil {
			return nil, err
		}

		if *record.Type == "SOA" || *record.Type == "NS" {
			continue
		}
		if *r.Name == "" {
			t := "@"
			record.Name = &t
		}
		if *record.Type == "CNAME" || *record.Type == "MX" || *record.Type == "ALIAS" || *record.Type == "SRV" {
			t := *record.Content + "."
			record.Content = &t
		}
		// exoscale adds these odd txt records that mirror the alias records.
		// they seem to manage them on deletes and things, so we'll just pretend they don't exist
		if *record.Type == "TXT" && strings.HasPrefix(*record.Content, "ALIAS for ") {
			continue
		}

		rc := &models.RecordConfig{
			TTL:      uint32(*record.TTL),
			Original: record,
		}
		rc.SetLabel(*record.Name, dc.Name)

		switch rtype := *record.Type; rtype {
		case "ALIAS", "URL":
			rc.Type = *record.Type
			rc.SetTarget(*record.Content)
		case "MX":
			var prio uint16
			if record.Priority != nil {
				prio = uint16(*record.Priority)
			}
			err = rc.SetTargetMX(prio, *record.Content)
			if err != nil {
				return nil, fmt.Errorf("unparsable record received from exoscale: %w", err)
			}
		default:
			err := rc.PopulateFromString(*record.Type, *record.Content, dc.Name)
			if err != nil {
				return nil, fmt.Errorf("unparsable record received from exoscale: %w", err)
			}
		}
		existingRecords = append(existingRecords, rc)
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
		record := del.Existing.Original.(*egoscale.DNSDomainRecord)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(*record.ID, *domain.ID),
		})
	}

	for _, cre := range create {
		rc := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   c.createRecordFunc(rc, *domain.ID),
		})
	}

	for _, mod := range modify {
		old := mod.Existing.Original.(*egoscale.DNSDomainRecord)
		new := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(old, new, *domain.ID),
		})
	}

	return corrections, nil
}

// Returns a function that can be invoked to create a record in a zone.
func (c *exoscaleProvider) createRecordFunc(rc *models.RecordConfig, domainID string) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.GetTargetField()
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record := egoscale.DNSDomainRecord{
			Name:    &name,
			Type:    &rc.Type,
			Content: &target,
		}
		if rc.TTL != 0 {
			ttl := int64(rc.TTL)
			record.TTL = &ttl
		}
		if rc.MxPreference != 0 {
			prio := int64(rc.MxPreference)
			record.Priority = &prio
		}

		_, err := c.client.CreateDNSDomainRecord(context.Background(), c.apiZone, domainID, &record)

		return err
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *exoscaleProvider) deleteRecordFunc(recordID, domainID string) func() error {
	return func() error {
		return c.client.DeleteDNSDomainRecord(
			context.Background(),
			c.apiZone,
			domainID,
			&egoscale.DNSDomainRecord{ID: &recordID},
		)
	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *exoscaleProvider) updateRecordFunc(record *egoscale.DNSDomainRecord, rc *models.RecordConfig, domainID string) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.GetTargetField()
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record.Name = &name
		record.Type = &rc.Type
		record.Content = &target
		if rc.TTL != 0 {
			ttl := int64(rc.TTL)
			record.TTL = &ttl
		}
		if rc.MxPreference != 0 {
			prio := int64(rc.MxPreference)
			record.Priority = &prio
		}

		return c.client.UpdateDNSDomainRecord(
			context.Background(),
			c.apiZone,
			domainID,
			record,
		)
	}
}

func (c *exoscaleProvider) findDomainByName(name string) (*egoscale.DNSDomain, error) {
	domains, err := c.client.ListDNSDomains(context.Background(), c.apiZone)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if *domain.UnicodeName == name {
			return &domain, nil
		}
	}

	return nil, ErrDomainNotFound
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
			printer.Printf("Warning: exoscale.com(.io, .ch, .net) does not allow NS records to be modified. %s will not be added.\n", rec.GetTargetField())
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
