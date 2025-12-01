package luadns

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	api "github.com/luadns/luadns-go"
	"golang.org/x/time/rate"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

/*

LuaDNS API DNS provider:

Info required in `creds.json`:
   - email
   - apikey
*/

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseHTTPS:            providers.Can(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "LUADNS"
	const providerMaintainer = "@riku22"
	fns := providers.DspFuncs{
		Initializer:   NewLuaDNS,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

type luadnsProvider struct {
	provider    *api.Client
	ctx         context.Context
	rateLimiter *rate.Limiter
	nameServers []string
	zones       []*api.Zone
}

// NewLuaDNS creates the provider.
func NewLuaDNS(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["email"] == "" || m["apikey"] == "" {
		return nil, errors.New("missing LuaDNS email or apikey")
	}
	ctx := context.Background()
	rateLimiter := rate.NewLimiter(4, 1)
	provider := api.NewClient(m["email"], m["apikey"])
	user, err := provider.Me(ctx)
	if err != nil {
		return nil, err
	}
	l := &luadnsProvider{
		provider:    provider,
		ctx:         ctx,
		rateLimiter: rateLimiter,
		nameServers: user.NameServers,
	}
	return l, nil
}

// GetNameservers returns the nameservers for a domain.
func (l *luadnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return models.ToNameserversStripTD(l.nameServers)
}

// ListZones returns a list of the DNS zones.
func (l *luadnsProvider) ListZones() ([]string, error) {
	if err := l.fetchDomainList(); err != nil {
		return nil, err
	}
	zoneList := make([]string, 0, len(l.zones))
	for _, d := range l.zones {
		zoneList = append(zoneList, d.Name)
	}
	return zoneList, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (l *luadnsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	zone, err := l.getZone(domain)
	if err != nil {
		return nil, err
	}
	records, err := l.getRecords(zone)
	if err != nil {
		return nil, err
	}
	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		newr, err := nativeToRecord(domain, records[i])
		if err != nil {
			return nil, err
		}
		existingRecords[i] = newr
	}
	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (l *luadnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	checkNS(dc)

	zone, err := l.getZone(dc.Name)
	if err != nil {
		return nil, 0, err
	}

	var corrs []*models.Correction

	changes, actualChangeCount, err := diff2.ByRecord(records, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, change := range changes {
		msg := change.Msgs[0]
		switch change.Type {
		case diff2.REPORT:
			corrs = []*models.Correction{{Msg: change.MsgsJoined}}
		case diff2.CREATE:
			corrs = l.makeCreateCorrection(change.New[0], zone, msg)
		case diff2.CHANGE:
			corrs = l.makeChangeCorrection(change.Old[0], change.New[0], zone, msg)
		case diff2.DELETE:
			corrs = l.makeDeleteCorrection(change.Old[0], zone, msg)
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", change.Type))
		}
		corrections = append(corrections, corrs...)
	}
	return corrections, actualChangeCount, nil
}

func (l *luadnsProvider) makeCreateCorrection(newrec *models.RecordConfig, zone *api.Zone, msg string) []*models.Correction {
	req := recordsToNative(newrec)
	return []*models.Correction{{
		Msg: msg,
		F: func() error {
			if err := l.rateLimiter.Wait(l.ctx); err != nil {
				return err
			}
			_, err := l.provider.CreateRecord(l.ctx, zone, req)
			if err != nil {
				return err
			}
			return nil
		},
	}}
}

func (l *luadnsProvider) makeChangeCorrection(oldrec *models.RecordConfig, newrec *models.RecordConfig, zone *api.Zone, msg string) []*models.Correction {
	recordID := oldrec.Original.(*api.Record).ID
	req := recordsToNative(newrec)
	return []*models.Correction{{
		Msg: fmt.Sprintf("%s, LuaDNS ID: %d", msg, recordID),
		F: func() error {
			if err := l.rateLimiter.Wait(l.ctx); err != nil {
				return err
			}
			_, err := l.provider.UpdateRecord(l.ctx, zone, recordID, req)
			if err != nil {
				return err
			}
			return nil
		},
	}}
}

func (l *luadnsProvider) makeDeleteCorrection(deleterec *models.RecordConfig, zone *api.Zone, msg string) []*models.Correction {
	recordID := deleterec.Original.(*api.Record).ID
	return []*models.Correction{{
		Msg: fmt.Sprintf("%s, LuaDNS ID: %d", msg, recordID),
		F: func() error {
			if err := l.rateLimiter.Wait(l.ctx); err != nil {
				return err
			}
			_, err := l.provider.DeleteRecord(l.ctx, zone, recordID)
			if err != nil {
				return err
			}
			return nil
		},
	}}
}

// EnsureZoneExists creates a zone if it does not exist
func (l *luadnsProvider) EnsureZoneExists(domain string, metadata map[string]string) error {
	if l.zones == nil {
		if err := l.fetchDomainList(); err != nil {
			return err
		}
	}
	if err := l.rateLimiter.Wait(l.ctx); err != nil {
		return err
	}
	zone, err := l.provider.CreateZone(l.ctx, &api.Zone{Name: domain})
	if err != nil {
		return err
	}
	l.zones = append(l.zones, zone)
	return nil
}

func (l *luadnsProvider) fetchDomainList() error {
	if err := l.rateLimiter.Wait(l.ctx); err != nil {
		return err
	}
	zones, err := l.provider.ListZones(l.ctx, &api.ListParams{})
	if err != nil {
		return err
	}
	l.zones = zones
	return nil
}

func (l *luadnsProvider) getZone(name string) (*api.Zone, error) {
	if l.zones == nil {
		if err := l.fetchDomainList(); err != nil {
			return nil, err
		}
	}
	for i := range l.zones {
		if l.zones[i].Name == name {
			return l.zones[i], nil
		}
	}
	return nil, fmt.Errorf("'%s' not a zone in luadns account", name)
}

func (l *luadnsProvider) getRecords(zone *api.Zone) ([]*api.Record, error) {
	if err := l.rateLimiter.Wait(l.ctx); err != nil {
		return nil, err
	}
	records, err := l.provider.ListRecords(l.ctx, zone, &api.ListParams{})
	if err != nil {
		return nil, err
	}
	var newRecords []*api.Record
	for _, rec := range records {
		if rec.Type == "SOA" {
			continue
		}
		newRecords = append(newRecords, rec)
	}
	return newRecords, nil
}

func nativeToRecord(domain string, r *api.Record) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      r.TTL,
		Original: r,
	}
	rc.SetLabelFromFQDN(r.Name, domain)
	var err error
	switch rtype := rc.Type; rtype {
	case "TXT":
		err = rc.SetTargetTXT(r.Content)
	default:
		err = rc.PopulateFromString(rtype, r.Content, domain)
	}
	return rc, err
}

func recordsToNative(rc *models.RecordConfig) *api.Record {
	r := &api.Record{
		Name: rc.GetLabelFQDN() + ".",
		Type: rc.Type,
		TTL:  rc.TTL,
	}
	switch rtype := rc.Type; rtype {
	case "TXT":
		r.Content = rc.GetTargetTXTJoined()
	case "HTTPS":
		content := fmt.Sprintf("%d %s %s", rc.SvcPriority, rc.GetTargetField(), rc.SvcParams)
		if rc.SvcParams == "" {
			content = content[:len(content)-1]
		}
		r.Content = content
	default:
		r.Content = rc.GetTargetCombined()
	}
	return r
}

func checkNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		// LuaDNS does not support changing the TTL of the default nameservers, so forcefully change the TTL to 86400.
		if rec.Type == "NS" && strings.HasSuffix(rec.GetTargetField(), ".luadns.net.") && rec.TTL != 86400 {
			rec.TTL = 86400
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
