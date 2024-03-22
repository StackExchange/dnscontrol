package akamaiedgedns

/*
Akamai Edge DNS provider

For information about Akamai Edge DNS, see:
https://www.akamai.com/us/en/products/security/edge-dns.jsp
https://learn.akamai.com/en-us/products/cloud_security/edge_dns.html
https://www.akamai.com/us/en/multimedia/documents/product-brief/edge-dns-product-brief.pdf
*/

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAKAMAICDN:        providers.Can(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

type edgeDNSProvider struct {
	contractID string
	groupID    string
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newEdgeDNSDSP,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("AKAMAIEDGEDNS", fns, features)
	providers.RegisterCustomRecordType("AKAMAICDN", "AKAMAIEDGEDNS", "")
}

// DnsServiceProvider
func newEdgeDNSDSP(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	clientSecret := config["client_secret"]
	host := config["host"]
	accessToken := config["access_token"]
	clientToken := config["client_token"]
	contractID := config["contract_id"]
	groupID := config["group_id"]

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
	if contractID == "" {
		return nil, fmt.Errorf("creds.json: contractID must not be empty")
	}
	if groupID == "" {
		return nil, fmt.Errorf("creds.json: groupID must not be empty")
	}

	initialize(clientSecret, host, accessToken, clientToken)

	api := &edgeDNSProvider{
		contractID: contractID,
		groupID:    groupID,
	}
	return api, nil
}

// EnsureZoneExists creates a zone if it does not exist
func (a *edgeDNSProvider) EnsureZoneExists(domain string) error {
	if zoneDoesExist(domain) {
		printer.Debugf("Zone %s already exists\n", domain)
		return nil
	}
	return createZone(domain, a.contractID, a.groupID)
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (a *edgeDNSProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {
	keysToUpdate, toReport, err := diff.NewCompat(dc).ChangedGroups(existingRecords)
	if err != nil {
		return nil, err
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
					return deleteRecordset(existing, dc.Name)
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
					return createRecordset(desired, dc.Name)
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
					return replaceRecordset(desired, dc.Name)
				},
			})
			printer.Debugf("replaceRecordset: %s %s\n", key.NameFQDN, key.Type)
			for _, rdata := range desired {
				printer.Debugf("  Rdata: %s\n", rdata.GetTargetCombined())
			}
		}
	}

	// Deletes first, then creates and replaces
	corrections = append(corrections, lastCorrections...)

	// AutoDnsSec correction
	existingAutoDNSSecEnabled, err := isAutoDNSSecEnabled(dc.Name)
	if err != nil {
		return nil, err
	}

	desiredAutoDNSSecEnabled := dc.AutoDNSSEC == "on"

	if !existingAutoDNSSecEnabled && desiredAutoDNSSecEnabled {
		// Existing false (disabled), Desired true (enabled)
		corrections = append(corrections, &models.Correction{
			Msg: "Enable AutoDnsSec\n",
			F: func() error {
				return autoDNSSecEnable(true, dc.Name)
			},
		})
		printer.Debugf("autoDNSSecEnable: Enable AutoDnsSec for zone %s\n", dc.Name)
	} else if existingAutoDNSSecEnabled && !desiredAutoDNSSecEnabled {
		// Existing true (enabled), Desired false (disabled)
		corrections = append(corrections, &models.Correction{
			Msg: "Disable AutoDnsSec\n",
			F: func() error {
				return autoDNSSecEnable(false, dc.Name)
			},
		})
		printer.Debugf("autoDNSSecEnable: Disable AutoDnsSec for zone %s\n", dc.Name)
	}

	return corrections, nil
}

// GetNameservers returns the nameservers for a domain.
func (a *edgeDNSProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	authorities, err := getAuthorities(a.contractID)
	if err != nil {
		return nil, err
	}
	return models.ToNameserversStripTD(authorities)
}

// GetZoneRecords returns an array of RecordConfig structs for a zone.
func (a *edgeDNSProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	records, err := getRecords(domain)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// ListZones returns all DNS zones managed by this provider.
func (a *edgeDNSProvider) ListZones() ([]string, error) {
	zones, err := listZones(a.contractID)
	if err != nil {
		return nil, err
	}
	return zones, nil
}
