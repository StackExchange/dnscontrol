package transip

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/transip/gotransip/v6"
	"github.com/transip/gotransip/v6/domain"
	"github.com/transip/gotransip/v6/repository"
)

/*

TransIP DNS Provider (transip.nl)

Info required in `creds.json`
	- AccessToken

*/

type transipProvider struct {
	client  *repository.Client
	domains *domain.Repository
}

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Can(),
	providers.CanUseAKAMAICDN:        providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseAzureAlias:       providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDHCID:            providers.Cannot(),
	providers.CanUseDNAME:            providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseHTTPS:            providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseRoute53Alias:     providers.Cannot(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSVCB:             providers.Cannot(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseDNSKEY:           providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// NewTransip creates a new TransIP provider.
func NewTransip(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {

	if m["AccessToken"] == "" && m["PrivateKey"] == "" {
		return nil, fmt.Errorf("no TransIP AccessToken or PrivateKey provided")
	}

	if m["PrivateKey"] != "" && m["AccountName"] == "" {
		return nil, fmt.Errorf("no AccountName given, required for authenticating with PrivateKey")
	}

	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		Token:            m["AccessToken"],
		AccountName:      m["AccountName"],
		PrivateKeyReader: strings.NewReader(m["PrivateKey"]),
	})

	if err != nil {
		return nil, fmt.Errorf("TransIP client fail %s", err.Error())
	}

	api := &transipProvider{}
	api.client = &client
	api.domains = &domain.Repository{Client: client}

	return api, nil
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   NewTransip,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("TRANSIP", fns, features)
}

func (n *transipProvider) ListZones() ([]string, error) {
	var domains []string

	domainsMap, err := n.domains.GetAll()
	if err != nil {
		return nil, err
	}
	for _, domainname := range domainsMap {
		domains = append(domains, domainname.Name)
	}

	sort.Strings(domains)

	return domains, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (n *transipProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, curRecords models.Records) ([]*models.Correction, error) {

	removeOtherNS(dc)

	corrections, err := n.getCorrectionsUsingDiff2(dc, curRecords)
	return corrections, err
}

func (n *transipProvider) getCorrectionsUsingDiff2(dc *models.DomainConfig, records models.Records) ([]*models.Correction, error) {
	var corrections []*models.Correction

	instructions, err := diff2.ByRecordSet(records, dc, nil)
	if err != nil {
		return nil, err
	}

	for _, change := range instructions {

		switch change.Type {
		case diff2.DELETE:
			oldEntries, err := recordsToNative(change.Old, true)
			if err != nil {
				return corrections, err
			}
			correction := change.CreateCorrection(
				wrapChangeFunction(
					oldEntries,
					func(rec domain.DNSEntry) error { return n.domains.RemoveDNSEntry(dc.Name, rec) },
				),
			)
			corrections = append(corrections, correction)

		case diff2.CREATE:
			newEntries, err := recordsToNative(change.New, false)
			if err != nil {
				return corrections, err
			}
			correction := change.CreateCorrection(
				wrapChangeFunction(
					newEntries,
					func(rec domain.DNSEntry) error { return n.domains.AddDNSEntry(dc.Name, rec) },
				),
			)
			corrections = append(corrections, correction)

		case diff2.CHANGE:
			if canDirectApplyDNSEntries(change) {
				newEntries, err := recordsToNative(change.New, false)
				if err != nil {
					return corrections, err
				}
				correction := change.CreateCorrection(
					wrapChangeFunction(
						newEntries,
						func(rec domain.DNSEntry) error { return n.domains.UpdateDNSEntry(dc.Name, rec) },
					),
				)
				corrections = append(corrections, correction)
			} else {
				corrections = append(
					corrections,
					n.recreateRecordSet(dc, change)...,
				)
			}
		case diff2.REPORT:
			corrections = append(corrections, change.CreateMessage())
		}

	}

	return corrections, nil
}

func (n *transipProvider) recreateRecordSet(dc *models.DomainConfig, change diff2.Change) []*models.Correction {
	var corrections []*models.Correction

	for _, rec := range change.Old {
		if existsInRecords(rec, change.New) {
			continue
		}

		nativeRec, _ := recordToNative(rec, true)
		createCorrection := change.CreateCorrectionWithMessage("[1/2] delete", func() error { return n.domains.RemoveDNSEntry(dc.Name, nativeRec) })
		corrections = append(corrections, createCorrection)
	}

	for _, rec := range change.New {
		if existsInRecords(rec, change.Old) {
			continue
		}

		nativeRec, _ := recordToNative(rec, false)
		createCorrection := change.CreateCorrectionWithMessage("[2/2] create", func() error { return n.domains.AddDNSEntry(dc.Name, nativeRec) })
		corrections = append(corrections, createCorrection)
	}

	return corrections
}

func existsInRecords(rec *models.RecordConfig, set models.Records) bool {
	for _, existing := range set {
		if rec.ToComparableNoTTL() == existing.ToComparableNoTTL() && rec.TTL == existing.TTL {
			return true
		}
	}

	return false
}

func recordsToNative(records models.Records, useOriginal bool) ([]domain.DNSEntry, error) {
	entries := make([]domain.DNSEntry, len(records))

	for iX, record := range records {
		entry, err := recordToNative(record, useOriginal)

		if err != nil {
			return nil, err
		}

		entries[iX] = entry
	}

	return entries, nil
}

func wrapChangeFunction(entries []domain.DNSEntry, executer func(rec domain.DNSEntry) error) func() error {
	return func() error {
		for _, entry := range entries {

			if err := executer(entry); err != nil {
				return err
			}
		}

		return nil
	}
}

// canDirectApplyDNSEntries determines if a change can be done in a single API call or
// if we must remove the old records and re-create them.  TransIP is unable to do certain
// changes in a single call. As we learn those situations, add them here.
func canDirectApplyDNSEntries(change diff2.Change) bool {
	desired, existing := change.New, change.Old

	if change.Type != diff2.CHANGE {
		return true
	}

	if len(desired) != len(existing) {
		return false
	}

	if len(desired) > 1 {
		return false
	}

	for i := 0; i < len(desired); i++ {
		if !canUpdateDNSEntry(desired[i], existing[i]) {
			return false
		}
	}

	return true
}

func canUpdateDNSEntry(desired *models.RecordConfig, existing *models.RecordConfig) bool {
	return desired.Name == existing.Name && desired.TTL == existing.TTL && desired.Type == existing.Type
}

func (n *transipProvider) GetZoneRecords(domainName string, meta map[string]string) (models.Records, error) {

	entries, err := n.domains.GetDNSEntries(domainName)
	if err != nil {
		return nil, err
	}

	var existingRecords = []*models.RecordConfig{}
	for _, entry := range entries {
		rts, err := nativeToRecord(entry, domainName)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rts)
	}

	return existingRecords, nil
}

func (n *transipProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	var nss []string

	entries, err := n.domains.GetNameservers(domainName)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		nss = append(nss, entry.Hostname)
	}

	return models.ToNameservers(nss)
}

// recordToNative convrts RecordConfig TO Native.
func recordToNative(config *models.RecordConfig, useOriginal bool) (domain.DNSEntry, error) {
	if useOriginal && config.Original != nil {
		return config.Original.(domain.DNSEntry), nil
	}

	return domain.DNSEntry{
		Name:    config.Name,
		Expire:  int(config.TTL),
		Type:    config.Type,
		Content: config.GetTargetCombinedFunc(nil),
	}, nil
}

// nativeToRecord converts native to RecordConfig.
func nativeToRecord(entry domain.DNSEntry, origin string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		TTL:      uint32(entry.Expire),
		Type:     entry.Type,
		Original: entry,
	}
	rc.SetLabel(entry.Name, origin)
	if err := rc.PopulateFromStringFunc(entry.Type, entry.Content, origin, nil); err != nil {
		return nil, fmt.Errorf("unparsable record received from TransIP: %w", err)
	}

	return rc, nil
}

func removeOtherNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && (strings.HasPrefix(rec.GetTargetField(), "ns0.transip") ||
			strings.HasPrefix(rec.GetTargetField(), "ns1.transip") ||
			strings.HasPrefix(rec.GetTargetField(), "ns2.transip")) {
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
