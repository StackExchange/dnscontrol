package transip

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v3/providers"
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

	domainsMap, _ := n.domains.GetAll()
	for _, domainname := range domainsMap {
		domains = append(domains, domainname.Name)
	}

	sort.Strings(domains)

	return domains, nil
}

func (n *transipProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	curRecords, err := n.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	if err := dc.Punycode(); err != nil {
		return nil, err
	}

	removeOtherNS(dc)

	models.PostProcessRecords(curRecords)

	if !diff2.EnableDiff2 {
		corrections, err := n.getCorrectionsUsingOldDiff(dc, curRecords)
		return corrections, err
	}

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

		if canDirectApplyDNSEntries(change) {
			changeFunction := n.getChangeFunction(change.Type, dc, change)
			corrections = append(corrections, change.CreateCorrection(changeFunction))
		} else {
			deleteFunction := n.getChangeFunction(diff2.DELETE, dc, change)
			createFunction := n.getChangeFunction(diff2.CREATE, dc, change)
			corrections = append(corrections, change.CreateCorrectionWithMessage("[1/2] delete", deleteFunction), change.CreateCorrectionWithMessage("[2/2] create", createFunction))
		}
	}

	return corrections, nil
}

func (n *transipProvider) getChangeFunction(changeType diff2.Verb, dc *models.DomainConfig, change diff2.Change) func() error {
	switch changeType {
	case diff2.DELETE:
		return wrapChangeFunction(change.Old, func(rec domain.DNSEntry) error { return n.domains.RemoveDNSEntry(dc.Name, rec) })

	case diff2.CREATE:
		return wrapChangeFunction(change.New, func(rec domain.DNSEntry) error { return n.domains.AddDNSEntry(dc.Name, rec) })

	case diff2.CHANGE:
		return wrapChangeFunction(change.Old, func(rec domain.DNSEntry) error { return n.domains.UpdateDNSEntry(dc.Name, rec) })
	}

	return nil
}

func wrapChangeFunction(records models.Records, executer func(rec domain.DNSEntry) error) func() error {
	return func() error {
		for _, record := range records {
			nativeRec, err := recordToNative(record)

			if err != nil {
				return err
			}

			if err := executer(nativeRec); err != nil {
				return err
			}
		}

		return nil
	}
}

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

func (n *transipProvider) getCorrectionsUsingOldDiff(dc *models.DomainConfig, records models.Records) ([]*models.Correction, error) {
	var corrections []*models.Correction

	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(records)

	if err != nil {
		return nil, err
	}

	for _, del := range del {
		entry, err := recordToNative(del.Existing)
		if err != nil {
			return nil, err
		}

		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   func() error { return n.domains.RemoveDNSEntry(dc.Name, entry) },
		})
	}

	for _, cre := range create {
		entry, err := recordToNative(cre.Desired)
		if err != nil {
			return nil, err
		}

		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   func() error { return n.domains.AddDNSEntry(dc.Name, entry) },
		})
	}

	for _, mod := range modify {
		targetEntry, err := recordToNative(mod.Desired)
		if err != nil {
			return nil, err
		}

		// TransIP identifies records by (Label, TTL Type), we can only update it if only the contents
		// has changed. Otherwise we delete the old record and create the new one
		if canUpdateDNSEntry(mod.Desired, mod.Existing) {
			corrections = append(corrections, &models.Correction{
				Msg: mod.String(),
				F:   func() error { return n.domains.UpdateDNSEntry(dc.Name, targetEntry) },
			})
		} else {
			oldEntry, err := recordToNative(mod.Existing)
			if err != nil {
				return nil, err
			}

			corrections = append(corrections,
				&models.Correction{
					Msg: mod.String() + "[1/2]",
					F:   func() error { return n.domains.RemoveDNSEntry(dc.Name, oldEntry) },
				},
				&models.Correction{
					Msg: mod.String() + "[2/2]",
					F:   func() error { return n.domains.AddDNSEntry(dc.Name, targetEntry) },
				},
			)
		}

	}

	return corrections, nil
}

func canUpdateDNSEntry(desired *models.RecordConfig, existing *models.RecordConfig) bool {
	return desired.Name == existing.Name && desired.TTL == existing.TTL && desired.Type == existing.Type
}

func (n *transipProvider) GetZoneRecords(domainName string) (models.Records, error) {

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

func recordToNative(config *models.RecordConfig) (domain.DNSEntry, error) {
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
	if err := rc.PopulateFromString(entry.Type, entry.Content, origin); err != nil {
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
		return models.StripQuotes(rc.GetTargetCombined())
	}
}
