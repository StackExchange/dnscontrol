package transip

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/txtutil"
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
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot(),
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

				oldEntries, err := recordsToNative(change.Old, true)
				if err != nil {
					return corrections, err
				}
				newEntries, err := recordsToNative(change.New, false)
				if err != nil {
					return corrections, err
				}

				deleteCorrection := wrapChangeFunction(oldEntries, func(rec domain.DNSEntry) error { return n.domains.RemoveDNSEntry(dc.Name, rec) })
				createCorrection := wrapChangeFunction(newEntries, func(rec domain.DNSEntry) error { return n.domains.AddDNSEntry(dc.Name, rec) })
				corrections = append(
					corrections,
					change.CreateCorrectionWithMessage("[1/2] delete", deleteCorrection),
					change.CreateCorrectionWithMessage("[2/2] create", createCorrection),
				)
			}
		case diff2.REPORT:
			corrections = append(corrections, change.CreateMessage())
		}

	}

	return corrections, nil
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

func recordToNative(config *models.RecordConfig, useOriginal bool) (domain.DNSEntry, error) {
	if useOriginal && config.Original != nil {
		return config.Original.(domain.DNSEntry), nil
	}

	return domain.DNSEntry{
		Name:    config.Name,
		Expire:  int(config.TTL),
		Type:    config.Type,
		Content: getTargetRecordContent(config),
	}, nil
}

func nativeToRecord(entry domain.DNSEntry, origin string) (*models.RecordConfig, error) {
	rc := &models.RecordConfig{
		TTL:      uint32(entry.Expire),
		Type:     entry.Type,
		Original: entry,
	}
	rc.SetLabel(entry.Name, origin)
	if err := rc.PopulateFromStringFunc(entry.Type, entry.Content, origin, txtutil.ParseQuoted); err != nil {
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

func getTargetRecordContent(rc *models.RecordConfig) string {
	switch rtype := rc.Type; rtype {
	case "SSHFP":
		return fmt.Sprintf("%d %d %s", rc.SshfpAlgorithm, rc.SshfpFingerprint, rc.GetTargetField())
	case "DS":
		return fmt.Sprintf("%d %d %d %s", rc.DsKeyTag, rc.DsAlgorithm, rc.DsDigestType, rc.DsDigest)
	case "SRV":
		return fmt.Sprintf("%d %d %d %s", rc.SrvPriority, rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	default:
		return models.StripQuotes(rc.GetTargetCombinedFunc(txtutil.EncodeQuoted))
	}
}
