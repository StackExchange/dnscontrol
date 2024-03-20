package realtimeregister

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/miekg/dns/dnsutil"
	"golang.org/x/exp/slices"
)

/*
Realtime Register DNS provider

Info required in `creds.json`:
  - apikey
  - premium: (0 for BASIC or 1 for PREMIUM)

Additional settings available in `creds.json`:
  - sandbox (set to 1 to use the sandbox API from realtime register)
*/

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDHCID:            providers.Cannot(),
	providers.CanUseDS:               providers.Cannot("Only for subdomains"),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

// init registers the domain service provider with dnscontrol.
func init() {
	fns := providers.DspFuncs{
		Initializer:   newRtrDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("REALTIMEREGISTER", fns, features)
	providers.RegisterRegistrarType("REALTIMEREGISTER", newRtrReg)
}

func newRtr(config map[string]string, _ json.RawMessage) (*realtimeregisterAPI, error) {
	apikey := config["apikey"]
	sandbox := config["sandbox"] == "1"

	if apikey == "" {
		return nil, fmt.Errorf("realtime register: apikey must be provided")
	}

	api := &realtimeregisterAPI{
		apikey:      apikey,
		endpoint:    getEndpoint(sandbox),
		Zones:       make(map[string]*Zone),
		ServiceType: getServiceType(config["premium"] == "1"),
	}

	return api, nil
}

func newRtrDsp(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newRtr(config, metadata)
}

func newRtrReg(config map[string]string) (providers.Registrar, error) {
	return newRtr(config, nil)
}

// GetNameservers Default name servers should not be included in the update
func (api *realtimeregisterAPI) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return []*models.Nameserver{}, nil
}

func (api *realtimeregisterAPI) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	response, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	records := response.Records
	recordConfigs := make([]*models.RecordConfig, len(records))
	for i := range records {
		recordConfigs[i] = toRecordConfig(domain, &records[i])
	}

	return recordConfigs, nil
}

func (api *realtimeregisterAPI) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	msgs, changes, err := diff2.ByZone(existing, dc, nil)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	if !changes {
		return corrections, nil
	}

	dnssec := api.Zones[dc.Name].Dnssec

	if api.Zones[dc.Name].Dnssec && dc.AutoDNSSEC == "off" {
		dnssec = false
		corrections = append(corrections,
			&models.Correction{
				Msg: "Update DNSSEC on -> off",
				F: func() error {
					return nil
				},
			})
	}

	if !api.Zones[dc.Name].Dnssec && dc.AutoDNSSEC == "on" {
		dnssec = true
		corrections = append(corrections,
			&models.Correction{
				Msg: "Update DNSSEC off -> on",
				F: func() error {
					return nil
				},
			})
	}

	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: strings.Join(msgs, "\n"),
				F: func() error {
					records := make([]Record, len(dc.Records))
					for i, r := range dc.Records {
						records[i] = toRecord(r)
					}
					zone := &Zone{Records: records, Dnssec: dnssec}

					err := api.updateZone(dc.Name, zone)
					if err != nil {
						return err
					}
					return nil
				},
			})
	}

	return corrections, nil
}

func (api *realtimeregisterAPI) ListZones() ([]string, error) {
	zones, err := api.getAllZones()
	if err != nil {
		return nil, err
	}
	return zones, nil
}

func (api *realtimeregisterAPI) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	nameservers, err := api.getDomainNameservers(dc.Name)
	if err != nil {
		return nil, err
	}

	expected := make([]string, len(dc.Nameservers))
	for i, ns := range dc.Nameservers {
		expected[i] = removeTrailingDot(ns.Name)
	}

	sort.Strings(nameservers)
	sort.Strings(expected)

	if !slices.Equal(nameservers, expected) {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s",
					strings.Join(nameservers, ","), strings.Join(expected, ",")),
				F: func() error { return api.updateNameservers(dc.Name, expected) },
			},
		}, nil
	}

	return nil, nil
}

func toRecordConfig(domain string, record *Record) *models.RecordConfig {

	recordConfig := &models.RecordConfig{
		Type:         record.Type,
		TTL:          uint32(record.TTL),
		MxPreference: uint16(record.Priority),
		SrvWeight:    uint16(0),
		SrvPort:      uint16(0),
		Original:     record,
	}

	recordConfig.SetLabelFromFQDN(record.Name, domain)

	switch rtype := record.Type; rtype { // #rtype_variations
	case "TXT":
		_ = recordConfig.SetTargetTXT(removeEscapeChars(record.Content))
	case "NS", "ALIAS", "CNAME":
		_ = recordConfig.SetTarget(dnsutil.AddOrigin(addTrailingDot(record.Content), domain))
	case "MX":
		content := record.Content
		if content != "." {
			content = addTrailingDot(content)
		}
		_ = recordConfig.SetTarget(dnsutil.AddOrigin(content, domain))
	case "NAPTR":
		_ = recordConfig.SetTargetNAPTRString(record.Content)
	case "SRV":
		parts := strings.Fields(record.Content)
		weight, _ := strconv.ParseUint(parts[0], 10, 16)
		port, _ := strconv.ParseUint(parts[1], 10, 16)
		content := parts[2]
		if content != "." {
			content = addTrailingDot(content)
		}
		_ = recordConfig.SetTargetSRV(uint16(record.Priority), uint16(weight), uint16(port), content)
	case "CAA":
		_ = recordConfig.SetTargetCAAString(record.Content)
	case "SSHFP":
		_ = recordConfig.SetTargetSSHFPString(record.Content)
	case "TLSA":
		_ = recordConfig.SetTargetTLSAString(record.Content)
	case "DS":
		_ = recordConfig.SetTargetDSString(record.Content)
	case "LOC":
		_ = recordConfig.SetTargetLOCString(domain, record.Content)
	default:
		_ = recordConfig.SetTarget(record.Content)
	}
	return recordConfig
}

func toRecord(recordConfig *models.RecordConfig) Record {
	record := &Record{
		Type:    recordConfig.Type,
		Name:    recordConfig.NameFQDN,
		Content: removeTrailingDot(recordConfig.GetTargetField()),
		TTL:     int(recordConfig.TTL),
	}

	switch rtype := recordConfig.Type; rtype {
	case "SRV":
		if record.Content == "" {
			record.Content = "."
		}
		record.Priority = int(recordConfig.SrvPriority)
		record.Content = fmt.Sprintf("%d %d %s", recordConfig.SrvWeight, recordConfig.SrvPort, record.Content)
	case "NAPTR", "SSHFP", "TLSA", "CAA":
		record.Content = recordConfig.GetTargetCombined()
	case "TXT":
		record.Content = addEscapeChars(record.Content)
	case "DS":
		record.Content = fmt.Sprintf("%d %d %d %s", recordConfig.DsKeyTag, recordConfig.DsAlgorithm,
			recordConfig.DsDigestType, strings.ToUpper(recordConfig.DsDigest))
	case "MX":
		if record.Content == "" {
			record.Content = "."
			record.Priority = -1
		} else {
			record.Priority = int(recordConfig.MxPreference)
		}
	case "LOC":
		parts := strings.Fields(recordConfig.GetTargetCombined())
		degrees1, _ := strconv.ParseUint(parts[0], 10, 32)
		minutes1, _ := strconv.ParseUint(parts[1], 10, 32)
		degrees2, _ := strconv.ParseUint(parts[4], 10, 32)
		minutes2, _ := strconv.ParseUint(parts[5], 10, 32)
		altitude, _ := strconv.ParseFloat(strings.Split(parts[8], "m")[0], 64)
		size, _ := strconv.ParseFloat(strings.Split(parts[9], "m")[0], 64)
		hp, _ := strconv.ParseFloat(strings.Split(parts[10], "m")[0], 64)
		vp, _ := strconv.ParseFloat(strings.Split(parts[11], "m")[0], 64)
		record.Content = fmt.Sprintf("%d %d %s %s %d %d %s %s %.2fm %.2fm %.2fm %.2fm",
			degrees1, minutes1, parts[2], parts[3], degrees2, minutes2,
			parts[6], parts[7], altitude, size, hp, vp,
		)
	}

	return *record
}

func (api *realtimeregisterAPI) EnsureZoneExists(domain string) error {
	exists, err := api.zoneExists(domain)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return api.createZone(domain)
}

func removeTrailingDot(record string) string {
	return strings.TrimSuffix(record, ".")
}

func addTrailingDot(record string) string {
	return record + "."
}

func removeEscapeChars(name string) string {
	return strings.Replace(strings.Replace(name, "\\\"", "\"", -1), "\\\\", "\\", -1)
}

func addEscapeChars(name string) string {
	return strings.Replace(strings.Replace(name, "\\", "\\\\", -1), "\"", "\\\"", -1)
}

func getEndpoint(sandbox bool) string {
	if sandbox {
		return endpointSandbox
	}
	return endpoint
}

func getServiceType(premium bool) string {
	if premium {
		return "PREMIUM"
	}
	return "BASIC"
}
