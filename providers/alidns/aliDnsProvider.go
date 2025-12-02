package alidns

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:   providers.Cannot(),
	providers.CanUseCAA:     providers.Can(),
	providers.CanUsePTR:     providers.Cannot(),
	providers.CanUseNAPTR:   providers.Cannot(),
	providers.CanUseSRV:     providers.Can(),
	providers.CanUseSSHFP:   providers.Cannot(),
	providers.CanUseTLSA:    providers.Cannot(),
	providers.CanAutoDNSSEC: providers.Can(),
	providers.CanConcur:     providers.Cannot(),

	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Can(),
	providers.DocCreateDomains:       providers.Can(),

	providers.CanUseRoute53Alias: providers.Cannot(),
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

}

type aliDnsDsp struct {
	client *alidns.Client
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
	return &aliDnsDsp{client}, nil
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

func (a *aliDnsDsp) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	keysToUpdate, toReport, actualChangeCount, err := diff.NewCompat(dc).ChangedGroups(existingRecords)
	if err != nil {
		return nil, 0, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	existingRecordsMap := make(map[models.RecordKey][]*models.RecordConfig)
	for _, r := range existingRecords {
		key := models.RecordKey{NameFQDN: r.NameFQDN, Type: r.Type}
		existingRecordsMap[key] = append(existingRecordsMap[key], r)
	}

	desiredRecordsMap := dc.Records.GroupedByKey()

	// Deletes must occur first. For example, if replacing a existing CNAME with an A of the same name:
	//    DELETE CNAME foo.example.net
	// must occur before
	//    CREATE A foo.example.net
	// because both an A and a CNAME for the same name is not allowed.

	lastCorrections := []*models.Correction{} // creates and replaces last

	for key, msg := range keysToUpdate {
		existing, okExisting := existingRecordsMap[key]
		desired, okDesired := desiredRecordsMap[key]

		if okExisting && !okDesired {
			// In the existing map but not in the desired map: Delete
			corrections = append(corrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return a.deleteRecordset(existing, dc.Name)
				},
			})
			printer.Debugf("deleteRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range existing {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		} else if !okExisting && okDesired {
			// Not in the existing map but in the desired map: Create
			lastCorrections = append(lastCorrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return a.createRecordset(desired, dc.Name)
				},
			})
			printer.Debugf("createRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range desired {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		} else if okExisting && okDesired {
			// In the existing map and in the desired map: Replace
			lastCorrections = append(lastCorrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return a.updateRecordset(existing, desired, dc.Name)
				},
			})
			printer.Debugf("updateRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range desired {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		}
	}

	// Append creates and updates after deletes
	corrections = append(corrections, lastCorrections...)

	printer.Debugf("Found %d corrections (actualChangeCount=%d)\n", len(corrections), actualChangeCount)
	return corrections, actualChangeCount, nil
}
