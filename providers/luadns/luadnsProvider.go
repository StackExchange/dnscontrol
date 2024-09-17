package luadns

import (
	"encoding/json"
	"fmt"

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
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
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

// NewLuaDNS creates the provider.
func NewLuaDNS(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	l := &luadnsProvider{}
	l.creds.email, l.creds.apikey = m["email"], m["apikey"]
	if l.creds.email == "" || l.creds.apikey == "" {
		return nil, fmt.Errorf("missing LuaDNS email or apikey")
	}

	// Get a domain to validate authentication
	if err := l.fetchDomainList(); err != nil {
		return nil, err
	}

	return l, nil
}

// GetNameservers returns the nameservers for a domain.
func (l *luadnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if len(l.nameserversNames) == 0 {
		l.fetchAvailableNameservers()
	}
	return models.ToNameserversStripTD(l.nameserversNames)
}

// ListZones returns a list of the DNS zones.
func (l *luadnsProvider) ListZones() ([]string, error) {
	if err := l.fetchDomainList(); err != nil {
		return nil, err
	}
	zones := make([]string, 0, len(l.domainIndex))
	for d := range l.domainIndex {
		zones = append(zones, d)
	}
	return zones, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (l *luadnsProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	domainID, err := l.getDomainID(domain)
	if err != nil {
		return nil, err
	}
	records, err := l.getRecords(domainID)
	if err != nil {
		return nil, err
	}
	existingRecords := make([]*models.RecordConfig, len(records))
	for i := range records {
		existingRecords[i] = nativeToRecord(domain, &records[i])
	}
	return existingRecords, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (l *luadnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, records models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	checkNS(dc)

	domainID, err := l.getDomainID(dc.Name)
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
			corrs = l.makeCreateCorrection(change.New[0], domainID, msg)
		case diff2.CHANGE:
			corrs = l.makeChangeCorrection(change.Old[0], change.New[0], domainID, msg)
		case diff2.DELETE:
			corrs = l.makeDeleteCorrection(change.Old[0], domainID, msg)
		default:
			panic(fmt.Sprintf("unhandled inst.Type %s", change.Type))
		}
		corrections = append(corrections, corrs...)
	}
	return corrections, actualChangeCount, nil
}

func (l *luadnsProvider) makeCreateCorrection(newrec *models.RecordConfig, domainID uint32, msg string) []*models.Correction {
	req := recordsToNative(newrec)
	return []*models.Correction{{
		Msg: msg,
		F: func() error {
			return l.createRecord(domainID, req)
		},
	}}
}

func (l *luadnsProvider) makeChangeCorrection(oldrec *models.RecordConfig, newrec *models.RecordConfig, domainID uint32, msg string) []*models.Correction {
	recordID := oldrec.Original.(*domainRecord).ID
	req := recordsToNative(newrec)
	return []*models.Correction{{
		Msg: fmt.Sprintf("%s, LuaDNS ID: %d", msg, recordID),
		F: func() error {
			return l.modifyRecord(domainID, recordID, req)
		},
	}}
}

func (l *luadnsProvider) makeDeleteCorrection(deleterec *models.RecordConfig, domainID uint32, msg string) []*models.Correction {
	recordID := deleterec.Original.(*domainRecord).ID
	return []*models.Correction{{
		Msg: fmt.Sprintf("%s, LuaDNS ID: %d", msg, recordID),
		F: func() error {
			return l.deleteRecord(domainID, recordID)
		},
	}}
}

// EnsureZoneExists creates a zone if it does not exist
func (l *luadnsProvider) EnsureZoneExists(domain string) error {
	if l.domainIndex == nil {
		if err := l.fetchDomainList(); err != nil {
			return err
		}
	}
	if _, ok := l.domainIndex[domain]; ok {
		return nil
	}
	return l.createDomain(domain)
}
