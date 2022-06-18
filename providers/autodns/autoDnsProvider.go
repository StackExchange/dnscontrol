package autodns

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
}

type autoDnsProvider struct {
	baseURL        url.URL
	defaultHeaders http.Header
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("AUTODNS", fns, features)
}

// New creates a new API handle.
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &autoDnsProvider{}

	api.baseURL = url.URL{
		Scheme: "https",
		User: url.UserPassword(
			settings["username"],
			settings["password"],
		),
		Host: "api.autodns.com",
		Path: "/v1/",
	}

	api.defaultHeaders = http.Header{
		"Accept":                []string{"application/json; charset=UTF-8"},
		"Content-Type":          []string{"application/json; charset=UTF-8"},
		"X-Domainrobot-Context": []string{settings["context"]},
	}

	return api, nil
}

// GetDomainCorrections returns the corrections for a domain.
func (api *autoDnsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	var changes []*models.RecordConfig

	dc, err := dc.Copy()
	if err != nil {
		return nil, err
	}

	err = dc.Punycode()
	if err != nil {
		return nil, err
	}
	domain := dc.Name

	// Get existing records
	existingRecords, err := api.GetZoneRecords(domain)
	if err != nil {
		return nil, err
	}

	// Normalize
	models.PostProcessRecords(existingRecords)
	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

	differ := diff.New(dc)
	unchanged, create, del, modify, err := differ.IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}

	for _, m := range unchanged {
		changes = append(changes, m.Desired)
	}

	for _, m := range del {
		// Just notify, these records don't have to be deleted explicitly
		printer.Debugf(m.String())
	}

	for _, m := range create {
		printer.Debugf(m.String())
		changes = append(changes, m.Desired)
	}

	for _, m := range modify {
		printer.Debugf("mod")
		printer.Debugf(m.String())
		changes = append(changes, m.Desired)
	}

	var corrections []*models.Correction

	if len(create) > 0 || len(del) > 0 || len(modify) > 0 {
		corrections = append(corrections,
			&models.Correction{
				Msg: "Zone update for " + domain,
				F: func() error {
					zoneTTL := uint32(0)
					nameServers := []*models.Nameserver{}
					resourceRecords := []*ResourceRecord{}

					for _, record := range changes {
						// NS records for the APEX should be handled differently
						if record.Type == "NS" && record.Name == "@" {
							nameServers = append(nameServers, &models.Nameserver{
								Name: strings.TrimSuffix(record.GetTargetField(), "."),
							})

							zoneTTL = record.TTL
						} else {
							resourceRecord := &ResourceRecord{
								Name:  record.Name,
								TTL:   int64(record.TTL),
								Type:  record.Type,
								Value: record.GetTargetField(),
							}

							if resourceRecord.Name == "@" {
								resourceRecord.Name = ""
							}

							if record.Type == "MX" {
								resourceRecord.Pref = int32(record.MxPreference)
							}

							if record.Type == "SRV" {
								resourceRecord.Value = fmt.Sprintf(
									"%d %d %d %s",
									record.SrvPriority,
									record.SrvWeight,
									record.SrvPort,
									record.GetTargetField(),
								)
							}

							resourceRecords = append(resourceRecords, resourceRecord)
						}
					}

					err := api.updateZone(domain, resourceRecords, nameServers, zoneTTL)

					if err != nil {
						printer.Errorf(err.Error())
					}

					return nil
				},
			})
	}

	return corrections, nil
}

// GetNameservers returns the nameservers for a domain.
func (api *autoDnsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := api.getZone(domain)

	if err != nil {
		return nil, err
	}

	return zone.NameServers, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *autoDnsProvider) GetZoneRecords(domain string) (models.Records, error) {
	zone, _ := api.getZone(domain)
	existingRecords := make([]*models.RecordConfig, len(zone.ResourceRecords))
	for i, resourceRecord := range zone.ResourceRecords {
		existingRecords[i] = toRecordConfig(domain, resourceRecord)

		// If TTL is not set for an individual RR AutoDNS defaults to the zone TTL defined in SOA
		if existingRecords[i].TTL == 0 {
			existingRecords[i].TTL = zone.Soa.TTL
		}
	}

	// AutoDNS doesn't respond with APEX nameserver records as regular RR but rather as a zone property
	for _, nameServer := range zone.NameServers {
		nameServerRecord := &models.RecordConfig{
			TTL: zone.Soa.TTL,
		}

		nameServerRecord.SetLabel("", domain)

		// make sure the value for this NS record is suffixed with a dot at the end
		_ = nameServerRecord.PopulateFromString("NS", strings.TrimSuffix(nameServer.Name, ".")+".", domain)

		existingRecords = append(existingRecords, nameServerRecord)
	}

	if zone.MainRecord != nil && zone.MainRecord.Value != "" {
		addressRecord := &models.RecordConfig{
			TTL: uint32(zone.MainRecord.TTL),
		}

		// If TTL is not set for an individual RR AutoDNS defaults to the zone TTL defined in SOA
		if addressRecord.TTL == 0 {
			addressRecord.TTL = zone.Soa.TTL
		}

		addressRecord.SetLabel("", domain)

		_ = addressRecord.PopulateFromString("A", zone.MainRecord.Value, domain)

		existingRecords = append(existingRecords, addressRecord)

		if zone.IncludeWwwForMain {
			prefixedAddressRecord := &models.RecordConfig{
				TTL: uint32(zone.MainRecord.TTL),
			}

			// If TTL is not set for an individual RR AutoDNS defaults to the zone TTL defined in SOA
			if prefixedAddressRecord.TTL == 0 {
				prefixedAddressRecord.TTL = zone.Soa.TTL
			}

			prefixedAddressRecord.SetLabel("www", domain)

			_ = prefixedAddressRecord.PopulateFromString("A", zone.MainRecord.Value, domain)

			existingRecords = append(existingRecords, prefixedAddressRecord)
		}
	}

	return existingRecords, nil
}

func toRecordConfig(domain string, record *ResourceRecord) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     record.Type,
		TTL:      uint32(record.TTL),
		Original: record,
	}
	rc.SetLabel(record.Name, domain)

	_ = rc.PopulateFromString(record.Type, record.Value, domain)

	if record.Type == "MX" {
		rc.MxPreference = uint16(record.Pref)
		rc.SetTarget(record.Value)
	}

	if record.Type == "SRV" {
		rc.SrvPriority = uint16(record.Pref)

		re := regexp.MustCompile(`(\d+) (\d+) (.+)$`)
		found := re.FindStringSubmatch(record.Value)

		weight, _ := strconv.Atoi(found[1])
		rc.SrvWeight = uint16(weight)

		port, _ := strconv.Atoi(found[2])
		rc.SrvPort = uint16(port)

		rc.SetTarget(found[3])
	}

	return rc
}
