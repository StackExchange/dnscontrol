package alidns

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanAutoDNSSEC:          providers.Cannot(),
	providers.CanConcur:              providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocCreateDomains:       providers.Cannot(),
	providers.CanUseRoute53Alias:     providers.Cannot(),
}

func init() {
	const providerName = "ALIDNS"
	const providerMaintainer = "@bytemain"
	fns := providers.DspFuncs{
		Initializer:   newAliDnsDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	// https://www.alibabacloud.com/help/en/dns/pubz-add-parsing-record#45347620b7mi9
	// Explicit URL forwarding uses 301 (permanent redirect) or 302 (temporary redirect)
	// redirection technology. The browser's address bar displays the target address, and the content displayed is from the target website.
	providers.RegisterCustomRecordType("EXPLICIT_URL_FORWARDING", providerName, "")
	// Implicit URL forwarding: Implicit URL Forwarding forwarding uses iframe technology.
	// The domain name in the browser's address bar does not change, but the content displayed is from the target website.
	providers.RegisterCustomRecordType("IMPLICIT_URL_FORWARDING", providerName, "")
	providers.RegisterMaintainer(providerName, providerMaintainer)
	// Register default TTL of 600 seconds (10 minutes) for Alibaba Cloud DNS
	// This is the minimum TTL for free/personal edition domains
	providers.RegisterDefaultTTL(providerName, 600)

}

type aliDnsDsp struct {
	client             *alidns.Client
	domainVersionCache map[string]*domainVersionInfo
	cacheMu            sync.Mutex
}

type domainVersionInfo struct {
	versionCode string
	minTTL      uint32
	maxTTL      uint32
}

func newAliDnsDsp(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	accessKeyID := config["access_key_id"]
	if accessKeyID == "" {
		return nil, fmt.Errorf("creds.json: access_key_id must not be empty")
	}

	accessKeySecret := config["access_key_secret"]
	if accessKeySecret == "" {
		return nil, fmt.Errorf("creds.json: access_key_secret must not be empty")
	}

	// Region ID defaults to "cn-hangzhou". The region value does not affect
	// DNS management (DNS is global) but Alibaba's SDK/examples require a
	// region to be provided — their docs/examples use Hangzhou:
	// https://www.alibabacloud.com/help/en/dns/quick-start-1
	region := config["region_id"]
	if region == "" {
		region = "cn-hangzhou"
	}

	client, err := alidns.NewClientWithAccessKey(
		region,
		accessKeyID,
		accessKeySecret,
	)
	if err != nil {
		return nil, err
	}
	return &aliDnsDsp{
		client:             client,
		domainVersionCache: make(map[string]*domainVersionInfo),
	}, nil
}

// GetZoneRecords returns an array of RecordConfig structs for a zone.
func (a *aliDnsDsp) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	// Fetch all pages of domain records.
	records, err := a.describeDomainRecordsAll(domain)
	if err != nil {
		return nil, err
	}

	out := models.Records{}
	for _, r := range records {
		if r.Status != "ENABLE" {
			continue
		}

		rc, err := nativeToRecord(r, domain)
		if err != nil {
			return nil, err
		}

		out = append(out, rc)
	}

	return out, nil
}

func (a *aliDnsDsp) ListZones() ([]string, error) {
	return a.describeDomainsAll()
}

func removeTrailingDot(record string) string {
	return strings.TrimSuffix(record, ".")
}

func deduplicateNameServerTargets(newRecs models.Records) models.Records {
	dedupedMap := make(map[string]bool)
	var deduped models.Records
	for _, rec := range newRecs {
		if !dedupedMap[rec.GetTargetField()] {
			dedupedMap[rec.GetTargetField()] = true
			deduped = append(deduped, rec)
		}
	}
	return deduped
}

// PrepDesiredRecords munges any records to best suit this provider.
func (a *aliDnsDsp) PrepDesiredRecords(dc *models.DomainConfig) {
	versionInfo, err := a.getDomainVersionInfo(dc.Name)
	if err != nil {
		return
	}

	recordsToKeep := make([]*models.RecordConfig, 0, len(dc.Records))

	for _, rec := range dc.Records {
		// If TTL is 0 (not set), use the minimum TTL as default
		if rec.TTL == 0 {
			rec.TTL = versionInfo.minTTL
		}

		if rec.TTL < versionInfo.minTTL {
			printer.Warnf("record %s has TTL %d which is below the minimum %d for this domain version (%s)\n",
				rec.GetLabelFQDN(), rec.TTL, versionInfo.minTTL, versionInfo.versionCode)
			rec.TTL = versionInfo.minTTL
		}
		if rec.TTL > versionInfo.maxTTL {
			printer.Warnf("record %s has TTL %d which exceeds the maximum %d\n",
				rec.GetLabelFQDN(), rec.TTL, versionInfo.maxTTL)
			rec.TTL = versionInfo.maxTTL
		}
		recordsToKeep = append(recordsToKeep, rec)
	}

	dc.Records = recordsToKeep
}

func (a *aliDnsDsp) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	a.PrepDesiredRecords(dc)

	var corrections []*models.Correction

	// Azure is a "ByRecordSet" API.
	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, change := range changes {
		// Copy all param values to local variables to avoid overwrites
		msgs := change.MsgsJoined
		dcn := dc.Name
		chaKey := change.Key

		if change.Type == diff2.CHANGE || change.Type == diff2.CREATE {
			if chaKey.Type == "NS" && dcn == removeTrailingDot(change.Key.NameFQDN) {
				change.New = deduplicateNameServerTargets(change.New)
			}
		}

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE:
			changeNew := change.New
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return a.createRecordset(changeNew, dcn)
				},
			})
		case diff2.CHANGE:
			changeNew := change.New
			changeExisting := change.Old
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return a.updateRecordset(changeExisting, changeNew, dcn)
				},
			})
		case diff2.DELETE:
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return a.deleteRecordset(change.Old, dcn)
				},
			})
		default:
			panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	return corrections, actualChangeCount, nil
}
