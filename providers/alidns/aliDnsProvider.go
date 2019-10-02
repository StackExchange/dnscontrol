package alidns

import (
	"encoding/json"
	"fmt"

	aerrs "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	adns "github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/pkg/errors"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"
)

// Api documentation: https://www.alibabacloud.com/help/doc-detail/34272.htm

const (
	metaLine   = "alidns_line"
	defaultTTL = 600
)

type aliDnsProvider struct {
	client *adns.Client
	zones  map[string]adns.Domain
}

func newAliDnsDsp(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	keyID, secretKey, token := m["KeyId"], m["SecretKey"], m["Token"]

	const region = "cn-hangzhou" // TODO(timonwong): Is this configurable?
	var client *adns.Client
	var err error
	if token == "" {
		client, err = adns.NewClientWithAccessKey(region, keyID, secretKey)
	} else {
		client, err = adns.NewClientWithStsToken(region, keyID, secretKey, token)
	}

	if err != nil {
		return nil, err
	}

	api := &aliDnsProvider{client: client}
	err = api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:    providers.Cannot(),
	providers.CanUseCAA:      providers.Can(),
	providers.CanUsePTR:      providers.Cannot(),
	providers.CanUseNAPTR:    providers.Cannot(),
	providers.CanUseSRV:      providers.Can(),
	providers.CanUseSSHFP:    providers.Cannot(),
	providers.CanUseTLSA:     providers.Cannot(),
	providers.CanUseTXTMulti: providers.Can(), // TODO(timonwong): Confirm this feature
	providers.CantUseNOPURGE: providers.Can(),

	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Can(),
	providers.DocCreateDomains:       providers.Can(),

	providers.CanUseRoute53Alias: providers.Cannot(),
}

func init() {
	const providerName = "ALI_DNS"
	providers.RegisterDomainServiceProviderType(providerName, newAliDnsDsp, features)
	providers.RegisterCustomRecordType("REDIRECT_URL", providerName, "") // HTTP 302
	providers.RegisterCustomRecordType("FORWARD_URL", providerName, "")  // iframe
}

func (a *aliDnsProvider) getZones() error {
	a.zones = make(map[string]adns.Domain)
	for pageNumber := 1; ; pageNumber++ {
		listInput := adns.CreateDescribeDomainsRequest()
		listInput.PageNumber = requests.NewInteger(pageNumber)
		listInput.PageSize = "100"
		list, err := a.client.DescribeDomains(listInput)
		if err != nil {
			return err
		}

		if len(list.Domains.Domain) == 0 {
			break
		}

		for _, z := range list.Domains.Domain {
			domain := z.DomainName
			a.zones[domain] = z
		}
	}
	return nil
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in your alidns account", e.domain)
}

func (a *aliDnsProvider) EnsureDomainExists(domain string) error {
	if _, ok := a.zones[domain]; ok {
		return nil
	}

	fmt.Printf("Adding zone for %s to alidns account\n", domain)
	req := adns.CreateAddDomainRequest()
	req.DomainName = domain
	_, err := a.client.AddDomain(req)
	return err
}

func (a *aliDnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, ok := a.zones[domain]
	if !ok {
		return nil, errNoExist{domain}
	}

	var ns []*models.Nameserver
	for _, nsName := range zone.DnsServers.DnsServer {
		ns = append(ns, &models.Nameserver{Name: nsName})
	}
	return ns, nil
}

func (a *aliDnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	_, ok := a.zones[dc.Name]
	// add zone if it doesn't exist
	if !ok {
		return nil, errNoExist{dc.Name}
	}

	for _, rec := range dc.Records {
		// NOTE: The "default" line is required when using custom line (regional, isp, etc)
		// TODO(timonwong): Add validation about missing line?
		if rec.Metadata[metaLine] == "" {
			rec.Metadata[metaLine] = "default"
		}
	}

	records, err := a.fetchRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	var existingRecords models.Records
	for i := range records {
		r := &records[i]
		existingRecords = append(existingRecords, nativeToRecord(r, dc.Name))
	}

	// Normalize
	models.PostProcessRecords(existingRecords)

	// diff
	differ := diff.New(dc, getRecordMetadata)
	_, create, del, mod := differ.IncrementalDiff(existingRecords)

	var corrections []*models.Correction
	for _, cre := range create {
		rec := cre.Desired
		corrections = append(corrections, &models.Correction{
			Msg: cre.String(),
			F:   func() error { return a.createRecord(rec, dc.Name) },
		})
	}

	for _, del := range del {
		rec := del.Existing.Original.(*adns.Record)
		corrections = append(corrections, &models.Correction{
			Msg: del.String(),
			F:   func() error { return a.deleteRecord(rec.RecordId) },
		})
	}

	for _, mod := range mod {
		old := mod.Existing.Original.(*adns.Record)
		rec := mod.Desired

		corrections = append(corrections, &models.Correction{
			Msg: mod.String(),
			F:   func() error { return a.updateRecord(old, rec, dc.Name) },
		})
	}

	return corrections, nil
}

func (a *aliDnsProvider) createRecord(rc *models.RecordConfig, domain string) error {
	req := adns.CreateAddDomainRecordRequest()
	req.DomainName = domain
	req.RR = rc.GetLabel()
	req.Type = rc.Type
	req.Value = rc.GetTargetField()
	req.TTL = requests.NewInteger(int(rc.TTL))
	req.Line = rc.Metadata[metaLine]

	if rc.Type == "MX" {
		req.Priority = requests.NewInteger(int(rc.MxPreference))
	}

	_, err := a.client.AddDomainRecord(req)
	return wrapSDKError(err)
}

func (a *aliDnsProvider) deleteRecord(recID string) error {
	req := adns.CreateDeleteDomainRecordRequest()
	req.RecordId = recID
	_, err := a.client.DeleteDomainRecord(req)
	return wrapSDKError(err)
}

func (a *aliDnsProvider) updateRecord(old *adns.Record, rc *models.RecordConfig, domainName string) error {
	req := adns.CreateUpdateDomainRecordRequest()
	req.RecordId = old.RecordId
	req.RR = rc.GetLabel()
	req.Type = rc.Type
	req.Value = rc.GetTargetField()
	req.TTL = requests.NewInteger(int(rc.TTL))
	req.Line = rc.Metadata[metaLine]

	if rc.Type == "MX" {
		req.Priority = requests.NewInteger(int(rc.MxPreference))
	}

	_, err := a.client.UpdateDomainRecord(req)
	return wrapSDKError(err)
}

func wrapSDKError(err error) error {
	if serverErr, ok := err.(*aerrs.ServerError); ok {
		if serverErr.ErrorCode() == "QuotaExceeded.TTL" {
			return errors.New(serverErr.Message())
		}
	}

	return err
}

func (a *aliDnsProvider) fetchRecords(domainName string) ([]adns.Record, error) {
	var records []adns.Record
	for pageNumber := 1; ; pageNumber++ {
		listInput := adns.CreateDescribeDomainRecordsRequest()
		listInput.DomainName = domainName
		listInput.PageNumber = requests.NewInteger(pageNumber)
		listInput.PageSize = "100"
		list, err := a.client.DescribeDomainRecords(listInput)
		if err != nil {
			return nil, err
		}

		if len(list.DomainRecords.Record) == 0 {
			break
		}

		records = append(records, list.DomainRecords.Record...)
	}

	// Make target into a FQDN if it is a CNAME, NS, MX, or SRV.
	for i := range records {
		recPtr := &records[i]
		if recPtr.Type == "CNAME" || recPtr.Type == "SRV" || recPtr.Type == "MX" || recPtr.Type == "NS" {
			recPtr.Value = recPtr.Value + "."
		}
	}
	return records, nil
}

func nativeToRecord(r *adns.Record, domain string) *models.RecordConfig {
	rec := &models.RecordConfig{
		TTL:      uint32(r.TTL),
		Metadata: map[string]string{},
		Original: r,
	}
	rec.SetLabel(r.RR, domain)

	switch rtype := r.Type; rtype {
	case "MX":
		err := rec.SetTargetMX(uint16(r.Priority), r.Value)
		if err != nil {
			panic(errors.Wrap(err, "unparsable MX record received from alidns"))
		}
	case "REDIRECT_URL", "FORWARD_URL":
		rec.Type = rtype
		rec.SetTarget(r.Value)
	default:
		err := rec.PopulateFromString(r.Type, r.Value, domain)
		if err != nil {
			panic(errors.Wrap(err, "unparsable record received from alidns"))
		}
	}

	if rec.TTL == 0 {
		rec.TTL = defaultTTL
	}

	rec.Metadata[metaLine] = r.Line
	return rec
}

func getRecordMetadata(r *models.RecordConfig) map[string]string {
	var line string
	if r.Original != nil {
		line = r.Original.(*adns.Record).Line
	} else {
		line = r.Metadata[metaLine]
	}

	return map[string]string{
		"line": line,
	}
}
