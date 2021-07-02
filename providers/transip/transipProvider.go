package transip

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/transip/gotransip/v6"
	"github.com/transip/gotransip/v6/domain"
	"github.com/transip/gotransip/v6/repository"
)

type transipProvider struct {
	client  *repository.Client
	domains *domain.Repository
}

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Can("Semicolons not supported in issue/issuewild fields.", "https://www.digitalocean.com/docs/networking/dns/how-to/create-caa-records"),
	providers.CanGetZones:            providers.Can(),
}

func NewTransip(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	if m["AccessToken"] == "" {
		return nil, fmt.Errorf("no TransIP token provided")
	}

	client, err := gotransip.NewClient(gotransip.ClientConfiguration{
		Token: m["AccessToken"],
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

func (n *transipProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	var corrections []*models.Correction

	// get current zone records
	curRecords, err := n.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	// post-process records
	if err := dc.Punycode(); err != nil {
		return nil, err
	}
	models.PostProcessRecords(curRecords)

	// create record diff by group
	differ := diff.New(dc)
	_, create, del, modify, err := differ.IncrementalDiff(curRecords)
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
			F:   func() error { return n.domains.UpdateDNSEntry(dc.Name, entry) },
		})
	}

	for _, mod := range modify {
		entry, err := recordToNative(mod.Desired)
		if err != nil {
			return nil, err
		}
		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   func() error { return n.domains.UpdateDNSEntry(dc.Name, entry) },
		})
	}

	return corrections, nil
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
	rc := &models.RecordConfig{TTL: uint32(*&entry.Expire)}
	rc.SetLabelFromFQDN(entry.Name, origin)
	if err := rc.PopulateFromString(entry.Type, entry.Content, origin); err != nil {
		return nil, fmt.Errorf("unparsable record received from TransIP: %w", err)
	}

	return rc, nil
}

func getTargetRecordContent(rc *models.RecordConfig) string {
	switch rtype := rc.Type; rtype {
	case "CAA":
		return rc.GetTargetCombined()
	case "SSHFP":
		return fmt.Sprintf("%d %d %s", rc.SshfpAlgorithm, rc.SshfpFingerprint, rc.GetTargetField())
	case "DS":
		return fmt.Sprintf("%d %d %d %s", rc.DsKeyTag, rc.DsAlgorithm, rc.DsDigestType, rc.DsDigest)
	case "SRV":
		return fmt.Sprintf("%d %d %s", rc.SrvWeight, rc.SrvPort, rc.GetTargetField())
	case "TXT":
		quoted := make([]string, len(rc.TxtStrings))
		for i := range rc.TxtStrings {
			quoted[i] = quoteDNSString(rc.TxtStrings[i])
		}
		return strings.Join(quoted, " ")
	case "NAPTR":
		return fmt.Sprintf("%d %d %s %s %s %s",
			rc.NaptrOrder, rc.NaptrPreference,
			quoteDNSString(rc.NaptrFlags), quoteDNSString(rc.NaptrService),
			quoteDNSString(rc.NaptrRegexp),
			rc.GetTargetField())
	default:
		return rc.GetTargetField()
	}
}
