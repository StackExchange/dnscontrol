package exoscale

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	egoscale "github.com/exoscale/egoscale/v3"
	"github.com/exoscale/egoscale/v3/credentials"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// ErrDomainNotFound error indicates domain name is not managed by Exoscale.
var ErrDomainNotFound = errors.New("domain not found")

type exoscaleProvider struct {
	client *egoscale.Client
}

// NewExoscale creates a new Exoscale DNS provider.
func NewExoscale(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	apiKey, secretKey := m["apikey"], m["secretkey"]

	creds := credentials.NewStaticCredentials(apiKey, secretKey)
	client, err := egoscale.NewClient(creds)
	if err != nil {
		return nil, err
	}

	// Endpoint is only for internal use now, not for production.
	endpoint := os.Getenv("EXOSCALE_API_ENDPOINT")
	if endpoint != "" {
		client = client.WithEndpoint(egoscale.Endpoint(endpoint))
	}

	ctx := context.Background()
	if z, ok := m["apizone"]; ok {
		endpoint, err := client.GetZoneAPIEndpoint(ctx, egoscale.ZoneName(z))
		if err != nil {
			return nil, fmt.Errorf("switch client zone: %w", err)
		}
		client = client.WithEndpoint(endpoint)
	}

	return &exoscaleProvider{
		client: client,
	}, nil
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Unimplemented(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported"),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("Exoscale does not allow sufficient control over the apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "EXOSCALE"
	const providerMaintainer = "@pierre-emmanuelJ"
	fns := providers.DspFuncs{
		Initializer:   NewExoscale,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// EnsureZoneExists creates a zone if it does not exist
func (c *exoscaleProvider) EnsureZoneExists(domain string) error {
	_, err := c.findDomainByName(domain)
	if errors.Is(err, egoscale.ErrNotFound) {
		_, err = c.client.CreateDNSDomain(context.Background(), egoscale.CreateDNSDomainRequest{
			UnicodeName: domain,
		})
	}

	return err
}

// GetNameservers returns the nameservers for domain.
func (c *exoscaleProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *exoscaleProvider) GetZoneRecords(domainName string, meta map[string]string) (models.Records, error) {
	//dc.Punycode()

	domain, err := c.findDomainByName(domainName)
	if err != nil {
		return nil, err
	}
	domainID := domain.ID

	ctx := context.Background()
	records, err := c.client.ListDNSDomainRecords(ctx, domainID)
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0, len(records.DNSDomainRecords))
	for _, r := range records.DNSDomainRecords {
		record, err := c.client.GetDNSDomainRecord(ctx, domainID, r.ID)
		if err != nil {
			return nil, err
		}

		var rcontent string

		if record.Type == "SOA" || record.Type == "NS" {
			continue
		}
		if record.Name == "" {
			record.Name = "@"
		}
		if record.Type == "CNAME" || record.Type == "MX" || record.Type == "ALIAS" || record.Type == "SRV" {
			t := rcontent + "."
			// for SRV records we need to aditionally prefix target with priority, which API handles as separate field.
			if record.Type == "SRV" && record.Priority != 0 {
				t = fmt.Sprintf("%d %s", record.Priority, t)
			}
			rcontent = t
		}
		// exoscale adds these odd txt records that mirror the alias records.
		// they seem to manage them on deletes and things, so we'll just pretend they don't exist
		if record.Type == "TXT" && strings.HasPrefix(rcontent, "ALIAS for ") {
			continue
		}

		rc := &models.RecordConfig{
			Original: record,
		}
		if record.Ttl != 0 {
			rc.TTL = uint32(record.Ttl)
		}
		rc.SetLabel(record.Name, domainName)

		switch record.Type {
		case "ALIAS", "URL":
			rc.Type = string(record.Type)
			rc.SetTarget(rcontent)
		case "MX":
			var prio uint16
			if record.Priority != 0 {
				prio = uint16(record.Priority)
			}
			err = rc.SetTargetMX(prio, rcontent)
		default:
			err = rc.PopulateFromString(string(record.Type), rcontent, domainName)
		}
		if err != nil {
			return nil, fmt.Errorf("unparsable record received from exoscale: %w", err)
		}

		existingRecords = append(existingRecords, rc)
	}

	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *exoscaleProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	removeOtherNS(dc)
	domain, err := c.findDomainByName(dc.Name)
	if err != nil {
		return nil, 0, err
	}

	toReport, create, toDelete, modify, actualChangeCount, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	for _, del := range toDelete {
		record := del.Existing.Original.(*egoscale.DNSDomainRecord)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(record.ID, domain.ID),
		})
	}

	for _, cre := range create {
		rc := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   c.createRecordFunc(rc, domain.ID),
		})
	}

	for _, mod := range modify {
		old := mod.Existing.Original.(*egoscale.DNSDomainRecord)
		new := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(old, new, domain.ID),
		})
	}

	return corrections, actualChangeCount, nil
}

// Returns a function that can be invoked to create a record in a zone.
func (c *exoscaleProvider) createRecordFunc(rc *models.RecordConfig, domainID egoscale.UUID) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()
		var prio int64

		if rc.Type == "MX" {
			target = rc.GetTargetField()

			if rc.MxPreference != 0 {
				prio = int64(rc.MxPreference)
			}
		}

		if rc.Type == "SRV" {
			// API wants priority as a separate argument, here we will strip it from combined target.
			sp := strings.Split(target, " ")
			target = strings.Join(sp[1:], " ")
			p, err := strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				return err
			}
			prio = p
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record := egoscale.CreateDNSDomainRecordRequest{
			Name:     name,
			Type:     egoscale.CreateDNSDomainRecordRequestType(rc.Type),
			Content:  target,
			Priority: prio,
		}

		if rc.TTL != 0 {
			record.Ttl = int64(rc.TTL)
		}

		ctx := context.Background()
		op, err := c.client.CreateDNSDomainRecord(ctx, domainID, record)
		if err != nil {
			return err

		}
		_, err = c.client.Wait(ctx, op, egoscale.OperationStateSuccess)

		return err
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *exoscaleProvider) deleteRecordFunc(recordID, domainID egoscale.UUID) func() error {
	return func() error {
		ctx := context.Background()
		op, err := c.client.DeleteDNSDomainRecord(ctx, domainID, recordID)
		if err != nil {
			return err
		}

		_, err = c.client.Wait(ctx, op, egoscale.OperationStateSuccess)
		return err
	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *exoscaleProvider) updateRecordFunc(record *egoscale.DNSDomainRecord, rc *models.RecordConfig, domainID egoscale.UUID) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.GetTargetField()

			if rc.MxPreference != 0 {
				record.Priority = int64(rc.MxPreference)
			}
		}

		if rc.Type == "SRV" {
			// API wants priority as separate argument, here we will strip it from combined target.
			sp := strings.Split(target, " ")
			target = strings.Join(sp[1:], " ")
			p, err := strconv.ParseInt(sp[0], 10, 64)
			if err != nil {
				return err
			}
			record.Priority = p
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record.Name = name
		record.Type = egoscale.DNSDomainRecordType(rc.Type)
		record.Content = target
		if rc.TTL != 0 {
			record.Ttl = int64(rc.TTL)
		}

		ctx := context.Background()
		op, err := c.client.UpdateDNSDomainRecord(ctx, domainID, record.ID, egoscale.UpdateDNSDomainRecordRequest{
			Name:     record.Name,
			Content:  record.Content,
			Priority: record.Priority,
			Ttl:      record.Ttl,
		})
		if err != nil {
			return err
		}
		_, err = c.client.Wait(ctx, op, egoscale.OperationStateSuccess)
		return err
	}
}

func (c *exoscaleProvider) findDomainByName(name string) (egoscale.DNSDomain, error) {
	domains, err := c.client.ListDNSDomains(context.Background())
	if err != nil {
		return egoscale.DNSDomain{}, err
	}

	return domains.FindDNSDomain(name)
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
