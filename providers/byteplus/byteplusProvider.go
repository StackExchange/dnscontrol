// byteplus support is done thanks to the authors of exoscale, porkbun, ovh providers!

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
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

const (
	minimumTTL = 600
)

var defaultNS = []string{
	"ns1.byteplusdns.com",
	"ns2.byteplusdns.net",
}

// ErrDomainNotFound error indicates domain name is not managed by Byteplus.
var ErrDomainNotFound = errors.New("domain not found")

type byteplusProvider struct {
	client *byteplus.Client
}

type domainStruct struct {
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

// GetNameservers returns the nameservers for a domain.
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

	// warning!! inconsistency within byteplus go sdk.
	// ListRecords later below is demanding string version if ZID (domain id)
	// while TopZoneResponse (findDomainByName) above returns in int64.
	domainIDString := strconv.Itoa(int(domainID))

	// Create the struct, setting the pointer to the string
	domainStr := domainStruct{
		DomainID: &domainIDString, // Take the address of domainIDStr
	}

	ctx := context.Background()
	pageSize := "500" // arbitrary limit. max records size per page is 500.
	records, err := c.client.ListRecords(ctx, &byteplus.ListRecordsRequest{ZID: domainStr.DomainID, PageSize: &pageSize})
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
			sp := strings.Split(*record.Value, " ")       // received combined value from byteplus "5 domain.com"
			rcontent = strings.Join(sp[1:], " ") + "."    // re-add trailing dot "5 domain.com."
			rprio, err := strconv.ParseInt(sp[0], 10, 64) // split get priority value

			var prio uint16
			if err != nil {
				return nil, err
			}
			prio = uint16(rprio)

			rc.SetTargetMX(prio, rcontent)
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
func (c *byteplusProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, actual models.Records) ([]*models.Correction, error) {
	corrections, err := c.getDiff2DomainCorrections(dc, actual)
	if err != nil {
		return nil, err
	}

	return corrections, nil
}

func (c *byteplusProvider) getDiff2DomainCorrections(dc *models.DomainConfig, actual models.Records) ([]*models.Correction, error) {
	domain, err := c.findDomainByName(dc.Name)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction
	instructions, err := diff2.ByRecord(actual, dc, nil)
	if err != nil {
		return nil, err
	}

	for _, inst := range instructions {
		switch inst.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: inst.MsgsJoined})
		case diff2.CHANGE:
			corrections = append(corrections, &models.Correction{
				Msg: inst.Msgs[0],
				F:   c.updateRecordFunc(inst.Old[0].Original.(*byteplus.QueryRecordResponse), inst.New[0], domain.ZID),
			})
		case diff2.CREATE:
			corrections = append(corrections, &models.Correction{
				Msg: inst.Msgs[0],
				F:   c.createRecordFunc(inst.New[0], domain.ZID),
			})
		case diff2.DELETE:
			rec := inst.Old[0].Original.(*byteplus.QueryRecordResponse)
			corrections = append(corrections, &models.Correction{
				Msg: inst.Msgs[0],
				F:   c.deleteRecordFunc(*rec.RecordID),
			})
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", inst.Type))
		}
	}
	return corrections, nil
}

// Returns a function that can be invoked to create a record in a zone.
func (c *byteplusProvider) createRecordFunc(rc *models.RecordConfig, domainID *int64) func() error {
	return func() error {
		target := rc.GetTargetCombined()
		name := rc.GetLabel()
		var prio *int64

		// byteplus have kinda(?) non-compliant spec for MX
		// the Weight value will be combined with domain name in "Value" key
		// instead of its own Weight key.
		// below combines MX weight + domain.
		if rc.Type == "MX" {
			prioStr := strconv.FormatInt(int64(rc.MxPreference), 10)
			target = prioStr + " " + rc.GetTargetField()
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

		// byteplus have kinda(?) non-compliant spec for MX
		// the Weight value will be combined with domain name in "Value" key
		// instead of its own Weight key.
		// below combines MX weight + domain.
		if rc.Type == "MX" {
			prioStr := strconv.FormatInt(int64(rc.MxPreference), 10)
			target = prioStr + " " + rc.GetTargetField()
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
