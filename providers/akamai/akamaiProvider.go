package akamai

/*
  ###TBD### Some comment about this wonderful provider
*/

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"strings"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilties.
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CantUseNOPURGE:         providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(), // ###TBD### "split horizon"
	providers.CanUseSOA:              providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAKAMAICDN:        providers.Can(),
}

type akamaiProvider struct {
	contractId string
	groupId    string
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newAkamaiDSP,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("AKAMAI", fns, features)
	providers.RegisterCustomRecordType("AKAMAICDN", "AKAMAI", "")
}

// DnsServiceProvider
func newAkamaiDSP(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	clientSecret := config["client_secret"]
	host := config["host"]
	accessToken := config["access_token"]
	clientToken := config["client_token"]
	contractId_ := config["contract_id"]
	groupId_ := config["group_id"]

	if clientSecret == "" {
		return nil, fmt.Errorf("creds.json: client_secret must not be empty")
	}
	if host == "" {
		return nil, fmt.Errorf("creds.json: host must not be empty")
	}
	if accessToken == "" {
		return nil, fmt.Errorf("creds.json: accessToken must not be empty")
	}
	if clientToken == "" {
		return nil, fmt.Errorf("creds.json: clientToken must not be empty")
	}
	if contractId_ == "" {
		return nil, fmt.Errorf("creds.json: contractId must not be empty")
	}
	if groupId_ == "" {
		return nil, fmt.Errorf("creds.json: groupId must not be empty")
	}

	AkaInitialize(clientSecret, host, accessToken, clientToken)

	api := &akamaiProvider{
		contractId: contractId_,
		groupId:    groupId_,
	}
	return api, nil
}

// AuditRecords returns an error if any records are not supportable by this provider.
func AuditRecords(records []*models.RecordConfig) error {
	return nil
}

// EnsureDomainExists configures a new zone if the zone does not already exist.
func (a *akamaiProvider) EnsureDomainExists(domain string) error {
	if AkaZoneDoesExist(domain) {
		printer.Debugf("Zone %s already exists\n", domain)
		return nil
	}
	return AkaCreateZone(domain, a.contractId, a.groupId)
}

// GetDomainCorrections return a list of corrections. Each correction is a text string describing the change
// and a function that, if called, will make the change.
// “dnscontrol preview” simply prints the text strings.
// "dnscontrol push" prints the strings and calls the functions.
func (a *akamaiProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	err := dc.Punycode()
	if err != nil {
		return nil, err
	}

	existingRecords, err := AkaGetRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	models.PostProcessRecords(existingRecords)
	txtutil.SplitSingleLongTxt(dc.Records)

	keysToUpdate, err := (diff.New(dc)).ChangedGroups(existingRecords)
	if err != nil {
		return nil, err
	}

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

	corrections := []*models.Correction{}     // deletes first
	lastCorrections := []*models.Correction{} // creates and replaces last

	for key, msg := range keysToUpdate {
		existing, okExisting := existingRecordsMap[key]
		desired, okDesired := desiredRecordsMap[key]

		if okExisting && !okDesired {
			// In the existing map but not in the desired map: Delete
			corrections = append(corrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return AkaDeleteRecordset(existing, dc.Name)
				},
			})
			printer.Debugf("AkaDeleteRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range existing {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		} else if !okExisting && okDesired {
			// Not in the existing map but in the desired map: Create
			lastCorrections = append(lastCorrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return AkaCreateRecordset(desired, dc.Name)
				},
			})
			printer.Debugf("AkaCreateRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range desired {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		} else if okExisting && okDesired {
			// In the existing map and in the desired map: Replace
			lastCorrections = append(lastCorrections, &models.Correction{
				Msg: strings.Join(msg, "\n   "),
				F: func() error {
					return AkaReplaceRecordset(desired, dc.Name)
				},
			})
			printer.Debugf("AkaReplaceRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range desired {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		}
	}

	// Deletes first, then creates and replaces
	corrections = append(corrections, lastCorrections...)

	// AutoDnsSec correction
	existingAutoDnsSecEnabled, err := AkaIsAutoDnsSecEnabled(dc.Name)
	if err != nil {
		return nil, err
	}

	desiredAutoDnsSecEnabled := dc.AutoDNSSEC == "on"

	if !existingAutoDnsSecEnabled && desiredAutoDnsSecEnabled {
		// Existing false (disabled), Desired true (enabled)
		corrections = append(corrections, &models.Correction{
			Msg: "Enable AutoDnsSec\n",
			F: func() error {
				return AkaAutoDnsSecEnable(true, dc.Name)
			},
		})
		printer.Debugf("AkaAutoDnsSecEnable: Enable AutoDnsSec for zone %s\n", dc.Name)
	} else if existingAutoDnsSecEnabled && !desiredAutoDnsSecEnabled {
		// Existing true (enabled), Desired false (disabled)
		corrections = append(corrections, &models.Correction{
			Msg: "Disable AutoDnsSec\n",
			F: func() error {
				return AkaAutoDnsSecEnable(false, dc.Name)
			},
		})
		printer.Debugf("AkaAutoDnsSecEnable: Disable AutoDnsSec for zone %s\n", dc.Name)
	}

	return corrections, nil
}

// GetNameservers returns the nameservers for a domain.
func (a *akamaiProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	authorities, err := AkaGetAuthorities(a.contractId)
	if err != nil {
		return nil, err
	}
	return models.ToNameserversStripTD(authorities)
}

// GetZoneRecords returns an array of RecordConfig structs for a zone.
func (a *akamaiProvider) GetZoneRecords(domain string) (models.Records, error) {
	records, err := AkaGetRecords(domain)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// ListZones returns all DNS zones managed by this provider.
func (a *akamaiProvider) ListZones() ([]string, error) {
	zones, err := AkaListZones(a.contractId)
	if err != nil {
		return nil, err
	}
	return zones, nil
}
