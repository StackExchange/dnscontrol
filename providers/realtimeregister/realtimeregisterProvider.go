package realtimeregister

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/miekg/dns/dnsutil"
	"strconv"
	"strings"
)

/*
Realtime Register DNS provider

Info required in `creds.json`:
  - apikey

Additional settings available in `creds.json`:
  - sandbox (set to 1 to use the sandbox API from realtime register)
*/

var features = providers.DocumentationNotes{
	providers.CanAutoDNSSEC:          providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot("Only for subdomains"),
	providers.CanUseDSForChildren:    providers.Can(),
	providers.CanUseLOC:              providers.Can("Getting invalid LOC format from API"),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Cannot(),
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
}

func newRtrDsp(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	apikey := config["apikey"]
	sandbox := config["sandbox"] == "1"

	if apikey == "" {
		return nil, fmt.Errorf("realtime register: apikey must be provided")
	}

	api := &realtimeregisterApi{apikey: apikey, endpoint: getEndpoint(sandbox)}

	return api, nil
}

// GetNameservers Default name servers can not be changed, and should not be included in the update
func (api *realtimeregisterApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	return []*models.Nameserver{}, nil
}

func (api *realtimeregisterApi) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	response, err := api.get(domain)
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

func (api *realtimeregisterApi) GetZoneRecordsCorrections(dc *models.DomainConfig, existing models.Records) ([]*models.Correction, error) {
	msgs, changes, err := diff2.ByZone(existing, dc, nil)
	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction
	if changes {
		corrections = append(corrections,
			&models.Correction{
				Msg: strings.Join(msgs, "\n"),
				F: func() error {
					records := make([]Record, len(dc.Records))
					for i, r := range dc.Records {
						records[i] = toRecord(r)
					}
					zone := &Zone{Records: records}

					err := api.post(dc.Name, zone)
					if err != nil {
						return err
					}
					return nil
				},
			})
	}

	return corrections, nil
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
		_ = recordConfig.SetTargetTXT(fixBackslashesAndDoubleQuotes(record.Content, true))
	case "NS", "ALIAS", "CNAME", "MX":
		_ = recordConfig.SetTarget(dnsutil.AddOrigin(addTrailingDot(record.Content), domain))
	case "NAPTR":
		parts := strings.Split(record.Content, " ")
		order, _ := strconv.ParseUint(parts[0], 10, 16)
		preference, _ := strconv.ParseUint(parts[1], 10, 16)
		_ = recordConfig.SetTargetNAPTR(
			uint16(order),
			uint16(preference),
			strings.Trim(parts[2], "\""),
			strings.Trim(parts[3], "\""),
			strings.Trim(parts[4], "\""),
			strings.Trim(parts[5], "\""),
		)
	case "SRV":
		parts := strings.Fields(record.Content)
		weight, _ := strconv.ParseUint(parts[0], 10, 16)
		port, _ := strconv.ParseUint(parts[1], 10, 16)
		_ = recordConfig.SetTargetSRV(uint16(record.Priority), uint16(weight), uint16(port), addTrailingDot(parts[2]))
	case "CAA":
		parts := strings.Fields(record.Content)
		caaFlag, _ := strconv.ParseUint(parts[0], 10, 8)
		_ = recordConfig.SetTargetCAA(uint8(caaFlag), parts[1], strings.Trim(strings.Join(parts[2:], " "), "\""))
	case "SSHFP":
		parts := strings.Fields(record.Content)
		algorithm, _ := strconv.ParseUint(parts[0], 10, 8)
		fingerprint, _ := strconv.ParseUint(parts[1], 10, 8)
		_ = recordConfig.SetTargetSSHFP(uint8(algorithm), uint8(fingerprint), parts[2])
	case "TLSA":
		parts := strings.Fields(record.Content)
		usage, _ := strconv.ParseUint(parts[0], 10, 8)
		selector, _ := strconv.ParseUint(parts[1], 10, 8)
		matchingtype, _ := strconv.ParseUint(parts[2], 10, 8)
		_ = recordConfig.SetTargetTLSA(uint8(usage), uint8(selector), uint8(matchingtype), parts[3])
	case "DS":
		parts := strings.Fields(record.Content)
		keytag, _ := strconv.ParseUint(parts[0], 10, 16)
		algorithm, _ := strconv.ParseUint(parts[1], 10, 8)
		digesttype, _ := strconv.ParseUint(parts[2], 10, 8)
		_ = recordConfig.SetTargetDS(uint16(keytag), uint8(algorithm), uint8(digesttype), parts[3])
	case "LOC":
		_ = recordConfig.SetTargetLOCString(domain, record.Content)
	default:
		_ = recordConfig.SetTarget(record.Content)
	}
	return recordConfig
}

func toRecord(recordConfig *models.RecordConfig) Record {
	record := &Record{
		Type:     recordConfig.Type,
		Name:     recordConfig.NameFQDN,
		Content:  removeTrailingDot(recordConfig.GetTargetField()),
		TTL:      int(recordConfig.TTL),
		Priority: int(recordConfig.MxPreference),
	}

	switch rtype := recordConfig.Type; rtype {
	case "SRV":
		record.Priority = int(recordConfig.SrvPriority)
		if record.Content == "" {
			record.Content = "."
		}
		record.Content = fmt.Sprintf("%d %d %s", recordConfig.SrvWeight, recordConfig.SrvPort, record.Content)
	case "CAA":
		record.Content = fmt.Sprintf("%d %s \"%s\"", recordConfig.CaaFlag, recordConfig.CaaTag, record.Content)
	case "NAPTR", "SSHFP", "TLSA":
		record.Content = recordConfig.GetTargetCombined()
	case "TXT":
		record.Content = fixBackslashesAndDoubleQuotes(record.Content, false)
	case "DS":
		record.Content = fmt.Sprintf("%d %d %d %s", recordConfig.DsKeyTag, recordConfig.DsAlgorithm,
			recordConfig.DsDigestType, strings.ToUpper(recordConfig.DsDigest))
	case "MX":
		if record.Content == "" {
			record.Content = "."
			record.Priority = -1
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

func removeTrailingDot(record string) string {
	return strings.TrimSuffix(record, ".")
}

func fixBackslashesAndDoubleQuotes(name string, inverse bool) string {
	if inverse {
		return strings.Replace(strings.Replace(name, "\\\"", "\"", -1), "\\\\", "\\", -1)
	}
	return strings.Replace(strings.Replace(name, "\\", "\\\\", -1), "\"", "\\\"", -1)
}

func addTrailingDot(record string) string {
	if strings.HasSuffix(record, ".") {
		return record
	}
	return record + "."
}

func getEndpoint(sandbox bool) string {
	if sandbox {
		return endpointSandbox
	}
	return endpoint
}
