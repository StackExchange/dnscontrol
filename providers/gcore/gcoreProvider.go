package gcore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"

	dnssdk "github.com/G-Core/gcore-dns-sdk-go"
)

/*
G-Core API DNS provider:
Info required in `creds.json`:
   - api-key
*/

type gcoreProvider struct {
	provider *dnssdk.Client
	ctx      context.Context
	apiKey   string
}

// NewGCore creates the provider.
func NewGCore(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["api-key"] == "" {
		return nil, fmt.Errorf("missing G-Core API key")
	}

	c := &gcoreProvider{
		provider: dnssdk.NewClient(dnssdk.PermanentAPIKeyAuth(m["api-key"])),
		ctx:      context.TODO(),
		apiKey:   m["api-key"],
	}

	return c, nil
}

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can("G-Core doesn't support SRV records with empty targets"),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

var defaultNameServerNames = []string{
	"ns1.gcorelabs.net",
	"ns2.gcdn.services",
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewGCore,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("GCORE", fns, features)
}

// GetNameservers returns the nameservers for a domain.
func (c *gcoreProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameservers(defaultNameServerNames)
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (c *gcoreProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	zone, err := c.provider.Zone(c.ctx, domain)
	if err != nil {
		return nil, err
	}

	// Convert RRsets to DNSControl format on the fly
	existingRecords := []*models.RecordConfig{}

	// We cannot directly use Zone's ShortAnswers, they aren't complete for CAA & SRV

	rrsets, err := c.dnssdkRRSets(domain)
	if err != nil {
		return nil, err
	}

	for _, rec := range rrsets.RRSets {
		nativeRecords, err := nativeToRecords(rec, zone.Name)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, nativeRecords...)
	}

	return existingRecords, nil
}

// EnsureZoneExists creates a zone if it does not exist
func (c *gcoreProvider) EnsureZoneExists(domain string) error {
	zones, err := c.provider.Zones(c.ctx)
	if err != nil {
		return err
	}

	for _, zone := range zones {
		if zone.Name == domain {
			return nil
		}
	}

	_, err = c.provider.CreateZone(c.ctx, domain)
	return err
}

func generateChangeMsg(updates []string) string {
	return strings.Join(updates, "\n")
}

// GenerateDomainCorrections takes the desired and existing records
// and produces a Correction list.  The correction list is simply
// a list of functions to call to actually make the desired
// correction, and a message to output to the user when the change is
// made.
func (c *gcoreProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {

	// Make delete happen earlier than creates & updates.
	var corrections []*models.Correction
	var deletions []*models.Correction
	var reports []*models.Correction

	changes, err := diff2.ByRecordSet(existing, dc, nil)
	if err != nil {
		return nil, err
	}

	for _, change := range changes {
		record := recordsToNative(change.New, change.Key)

		// Copy all params to avoid overwrites
		zone := dc.Name
		name := change.Key.NameFQDN
		typ := change.Key.Type
		msg := generateChangeMsg(change.Msgs)

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE:
			corrections = append(corrections, &models.Correction{
				Msg: msg,
				F: func() error {
					return c.provider.CreateRRSet(c.ctx, zone, name, typ, *record)
				},
			})
		case diff2.CHANGE:
			corrections = append(corrections, &models.Correction{
				Msg: msg,
				F: func() error {
					return c.provider.UpdateRRSet(c.ctx, zone, name, typ, *record)
				},
			})
		case diff2.DELETE:
			deletions = append(deletions, &models.Correction{
				Msg: msg,
				F: func() error {
					return c.provider.DeleteRRSet(c.ctx, zone, name, typ)
				},
			})
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	result := append(reports, deletions...)
	result = append(result, corrections...)
	return result, nil
}
