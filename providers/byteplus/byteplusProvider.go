package byteplus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	byteplus "github.com/byteplus-sdk/byteplus-sdk-golang/service/dns"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

// ErrDomainNotFound error indicates domain name is not managed by Byteplus.
var ErrDomainNotFound = errors.New("domain not found")

type byteplusProvider struct {
	client *byteplus.Client
}

type MyStruct struct {
	DomainID *string
}

// NewByteplus creates a new Byteplus DNS provider.
func NewByteplus(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	ak, sk := m["ak"], m["sk"]

	client := byteplus.InitDNSBytePlusClient()
	client.SetAccessKey(ak)
	client.SetSecretKey(sk)

	provider := byteplusProvider{client: client}

	return &provider, nil
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewByteplus,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("BYTEPLUS", fns, features)
}

// EnsureZoneExists creates a zone if it does not exist
func (c *byteplusProvider) EnsureZoneExists(domain string) error {
	_, err := c.findDomainByName(domain)

	return err
}

// GetNameservers returns the nameservers for domain.
func (c *byteplusProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return nil, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *byteplusProvider) GetZoneRecords(domainName string, meta map[string]string) (models.Records, error) {
	//dc.Punycode()

	domain, err := c.findDomainByName(domainName)
	if err != nil {
		return nil, err
	}
	domainID := *domain.ZID

	//byteplus package issue
	domainIDStr := strconv.Itoa(int(domainID))

	// Create the struct, setting the pointer to the string
	myStruct := MyStruct{
		DomainID: &domainIDStr, // Take the address of domainIDStr
	}

	ctx := context.Background()
	records, err := c.client.ListRecords(ctx, &byteplus.ListRecordsRequest{ZID: myStruct.DomainID})
	if err != nil {
		return nil, err
	}

	existingRecords := make([]*models.RecordConfig, 0, len(records.Records))
	for _, r := range records.Records {
		if r.RecordID == nil {
			continue
		}

		recordID := *r.RecordID

		record, err := c.client.QueryRecord(ctx, &byteplus.QueryRecordRequest{RecordID: &recordID})
		if err != nil {
			return nil, err
		}

		// nil pointers are not expected, but just to be on the safe side...
		var rtype, rcontent, rname string
		if record.Type == nil {
			continue
		}
		rtype = *record.Type
		if record.Value != nil {
			rcontent = *record.Value
		}
		if record.Host != nil {
			rname = *record.Host
		}

		if rtype == "SOA" || rtype == "NS" {
			continue
		}
		if rname == "" {
			t := "@"
			record.Host = &t
		}
		if rtype == "CNAME" || rtype == "MX" || rtype == "ALIAS" || rtype == "SRV" {
			t := rcontent + "."
			// for SRV records we need to aditionally prefix target with priority, which API handles as separate field.
			if rtype == "SRV" && record.Weight != nil {
				t = fmt.Sprintf("%d %s", *record.Weight, t)
			}
			rcontent = t
		}
		// byteplus adds these odd txt records that mirror the alias records.
		// they seem to manage them on deletes and things, so we'll just pretend they don't exist
		if rtype == "TXT" && strings.HasPrefix(rcontent, "ALIAS for ") {
			continue
		}

		rc := &models.RecordConfig{
			Original: record,
		}
		if record.TTL != nil {
			rc.TTL = uint32(*record.TTL)
		}
		rc.SetLabel(rname, domainName)

		switch rtype {
		case "ALIAS", "URL":
			rc.Type = rtype
			rc.SetTarget(rcontent)
		case "MX":
			var prio uint16
			if record.Weight != nil {
				prio = uint16(*record.Weight)
			}
			err = rc.SetTargetMX(prio, rcontent)
		default:
			err = rc.PopulateFromString(rtype, rcontent, domainName)
		}
		if err != nil {
			return nil, fmt.Errorf("unparsable record received from byteplus: %w", err)
		}

		existingRecords = append(existingRecords, rc)
	}

	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (c *byteplusProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {

	removeOtherNS(dc)
	domain, err := c.findDomainByName(dc.Name)
	if err != nil {
		return nil, err
	}
	domainID := domain.ZID

	toReport, create, toDelete, modify, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	for _, del := range toDelete {
		record := del.Existing.Original.(*byteplus.QueryRecordResponse)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   c.deleteRecordFunc(*record.RecordID),
		})
	}

	for _, cre := range create {
		rc := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   c.createRecordFunc(rc, domainID),
		})
	}

	for _, mod := range modify {
		old := mod.Existing.Original.(*byteplus.QueryRecordResponse)
		new := mod.Desired
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   c.updateRecordFunc(old, new, domainID),
		})
	}

	return corrections, nil
}

// Returns a function that can be invoked to create a record in a zone.
func (c *byteplusProvider) createRecordFunc(rc *models.RecordConfig, domainID *int64) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()
		var prio *int64

		if rc.Type == "MX" {
			target = rc.GetTargetField()

			if rc.MxPreference != 0 {
				p := int64(rc.MxPreference)
				prio = &p
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
			prio = &p
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record := byteplus.CreateRecordRequest{
			Host:   &name,
			Type:   &rc.Type,
			Value:  &target,
			ZID:    domainID,
			Weight: prio,
		}

		if rc.TTL != 0 {
			ttl := int64(rc.TTL)
			record.TTL = &ttl
		}

		_, err := c.client.CreateRecord(context.Background(), &record)

		return err
	}
}

// Returns a function that can be invoked to delete a record in a zone.
func (c *byteplusProvider) deleteRecordFunc(recordID string) func() error {
	return func() error {
		return c.client.DeleteRecord(context.Background(), &byteplus.DeleteRecordRequest{RecordID: &recordID})
	}
}

// Returns a function that can be invoked to update a record in a zone.
func (c *byteplusProvider) updateRecordFunc(record *byteplus.QueryRecordResponse, rc *models.RecordConfig, domainID *int64) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()

		if rc.Type == "MX" {
			target = rc.GetTargetField()

			if rc.MxPreference != 0 {
				p := int64(rc.MxPreference)
				record.Weight = &p
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
			record.Weight = &p
		}

		if rc.Type == "NS" && (name == "@" || name == "") {
			name = "*"
		}

		record.Host = &name
		record.Type = &rc.Type
		record.Value = &target
		if rc.TTL != 0 {
			ttl := int64(rc.TTL)
			record.TTL = &ttl
		}

		newRecord := &byteplus.UpdateRecordRequest{
			Host:     *record.Host,
			Line:     *record.Line,
			RecordID: *record.RecordID,
			TTL:      record.TTL,
			Type:     record.Type,
			Value:    record.Value,
			Weight:   record.Weight,
		}

		_, err := c.client.UpdateRecord(context.Background(), newRecord)

		return err
	}
}

func (c *byteplusProvider) findDomainByName(name string) (*byteplus.TopZoneResponse, error) {
	domains, err := c.client.ListZones(context.Background(), &byteplus.ListZonesRequest{})
	if err != nil {
		return nil, err
	}

	for _, domain := range domains.Zones {
		if domain.ZoneName != nil && domain.ZID != nil && *domain.ZoneName == name {
			return &domain, nil
		}
	}

	return nil, ErrDomainNotFound
}

func defaultNSSUffix(defNS string) bool {
	return (strings.HasSuffix(defNS, ".byteplus.io.") ||
		strings.HasSuffix(defNS, ".byteplus.com.") ||
		strings.HasSuffix(defNS, ".byteplus.net."))
}

// remove all non-byteplus NS records from our desired state.
// if any are found, print a warning
func removeOtherNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" {
			// apex NS inside byteplus are expected.
			if rec.GetLabelFQDN() == dc.Name && defaultNSSUffix(rec.GetTargetField()) {
				continue
			}
			printer.Printf("Warning: byteplus.com(.io, .ch, .net) does not allow NS records to be modified. %s will not be added.\n", rec.GetTargetField())
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
