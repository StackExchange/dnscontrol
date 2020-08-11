package inwx

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"

	"github.com/nrdcg/goinwx"
)

/*
INWX Registrar and DNS provider

Info required in `creds.json`:
	- username
	- password
	- totp (TOPT code if 2FA is enabled)

Additional settings available in `creds.json`:
	- sandbox (set to 1 to use the sandbox API from INWX)

*/

type InwxApi struct {
	client  *goinwx.Client
	sandbox bool
}

var InwxDefaultNs = []string{"ns.inwx.de", "ns2.inwx.de", "ns3.inwx.eu", "ns4.inwx.com", "ns5.inwx.net"}
var InwxSandboxDefaultNs = []string{"ns.ote.inwx.de", "ns2.ote.inwx.de"}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot("INWX does not support the ALIAS or ANAME record type."),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Unimplemented("DS records require a different API call that hasn't been implemented yet."),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUseSRV:              providers.Can("SRV records with empty targets are not supported."),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.CanUseTXTMulti:         providers.Cannot("INWX only supports a single entry for TXT records"),
	providers.CanAutoDNSSEC:          providers.Unimplemented("Supported by INWX but not implemented yet."),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.DocDualHost:            providers.Can(),
	providers.DocCreateDomains:       providers.Unimplemented("Supported by INWX but not implemented yet."),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAzureAlias:       providers.Cannot(),
}

func init() {
	providers.RegisterRegistrarType("INWX", newInwxReg)
	providers.RegisterDomainServiceProviderType("INWX", newInwxDsp, features)
}

func newInwx(m map[string]string) (*InwxApi, error) {
	for key := range m {
		switch key {
		case "username",
			"password",
			"totp",
			"sandbox",
			"domain":
			continue
		default:
			fmt.Printf("INWX: WARNING: unknown key in `creds.json` (%s)\n", key)
		}
	}

	if m["username"] == "" {
		return nil, fmt.Errorf("INWX: username must be provided.")
	}
	if m["password"] == "" {
		return nil, fmt.Errorf("INWX: password must be provided.")
	}

	sandbox := m["sandbox"] == "1"

	opts := &goinwx.ClientOptions{Sandbox: sandbox}
	client := goinwx.NewClient(m["username"], m["password"], opts)

	_, err := client.Account.Login()
	if err != nil {
		return nil, fmt.Errorf("Unable to login to INWX")
	}

	if m["totp"] != "" {
		err := client.Account.Unlock(m["totp"])
		if err != nil {
			return nil, fmt.Errorf("Could not unlock INWX account - TOTP is probably invalid.")
		}
	}

	api := &InwxApi{client: client, sandbox: sandbox}

	return api, nil
}

func newInwxReg(m map[string]string) (providers.Registrar, error) {
	return newInwx(m)
}

func newInwxDsp(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newInwx(m)
}

func makeNameserverRecordRequest(domain string, rec *models.RecordConfig) *goinwx.NameserverRecordRequest {
	content := rec.GetTargetField()

	req := &goinwx.NameserverRecordRequest{
		Domain:  domain,
		Type:    rec.Type,
		Content: content,
		Name:    rec.GetLabel(),
		TTL:     int(rec.TTL),
	}

	switch rType := rec.Type; rType {
	/* INWX is a little bit special for CNAME,NS,MX and SRV records:
	   The API will not accept any target with a final dot but will
	   instead always add this final dot internally.
	   Records with empty targets (i.e. records with target ".")
	   are not allowed.
	*/
	case "CNAME", "NS":
		req.Content = content[:len(content)-1]
	case "MX":
		req.Priority = int(rec.MxPreference)
		req.Content = content[:len(content)-1]
	case "SRV":
		req.Priority = int(rec.SrvPriority)
		req.Content = fmt.Sprintf("%d %d %v", rec.SrvWeight, rec.SrvPort, content[:len(content)-1])
	default:
		req.Content = rec.GetTargetCombined()
	}

	return req
}

func (api *InwxApi) createRecord(domain string, rec *models.RecordConfig) error {
	req := makeNameserverRecordRequest(domain, rec)
	_, err := api.client.Nameservers.CreateRecord(req)
	return err
}

func (api *InwxApi) updateRecord(RoId int, rec *models.RecordConfig) error {
	req := makeNameserverRecordRequest("", rec)
	err := api.client.Nameservers.UpdateRecord(RoId, req)
	return err
}

func (api *InwxApi) deleteRecord(RoId int) error {
	return api.client.Nameservers.DeleteRecord(RoId)
}

func (api *InwxApi) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	dc.Punycode()

	foundRecords, err := api.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	models.PostProcessRecords(foundRecords)

	differ := diff.New(dc)
	_, create, del, mod := differ.IncrementalDiff(foundRecords)
	corrections := []*models.Correction{}

	for _, d := range create {
		des := d.Desired
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.createRecord(dc.Name, des) },
		})
	}
	for _, d := range del {
		existingId := d.Existing.Original.(goinwx.NameserverRecord).ID
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.deleteRecord(existingId) },
		})
	}
	for _, d := range mod {
		rec := d.Desired
		existingId := d.Existing.Original.(goinwx.NameserverRecord).ID
		corrections = append(corrections, &models.Correction{
			Msg: d.String(),
			F:   func() error { return api.updateRecord(existingId, rec) },
		})
	}

	return corrections, nil
}

func (api *InwxApi) GetNameservers(domain string) ([]*models.Nameserver, error) {
	if api.sandbox {
		return models.ToNameservers(InwxSandboxDefaultNs)
	} else {
		return models.ToNameservers(InwxDefaultNs)
	}
}

func (api *InwxApi) GetZoneRecords(domain string) (models.Records, error) {
	info, err := api.client.Nameservers.Info(&goinwx.NameserverInfoRequest{Domain: domain})
	if err != nil {
		return nil, err
	}

	var records = []*models.RecordConfig{}

	for _, record := range info.Records {
		if record.Type == "SOA" {
			continue
		}

		/* INWX is a little bit special for CNAME,NS,MX and SRV records:
		   The API will not accept any target with a final dot but will
		   instead always add this final dot internally.
		   Records with empty targets (i.e. records with target ".")
		   are not allowed.
		*/
		if record.Type == "CNAME" || record.Type == "MX" || record.Type == "NS" || record.Type == "SRV" {
			record.Content = record.Content + "."
		}

		rc := &models.RecordConfig{
			TTL:      uint32(record.TTL),
			Original: record,
		}
		rc.SetLabelFromFQDN(record.Name, domain)

		switch rType := record.Type; rType {
		case "MX":
			err = rc.SetTargetMX(uint16(record.Priority), record.Content)
		case "SRV":
			err = rc.SetTargetSRVPriorityString(uint16(record.Priority), record.Content)
		default:
			err = rc.PopulateFromString(rType, record.Content, domain)
		}

		if err != nil {
			panic(fmt.Errorf("INWX: unparsable record received: %w", err))
		}

		records = append(records, rc)
	}

	return records, nil
}

func (api *InwxApi) updateNameservers(ns []string, domain string) func() error {
	return func() error {
		request := &goinwx.DomainUpdateRequest{
			Domain:      domain,
			Nameservers: ns,
		}

		_, err := api.client.Domains.Update(request)
		return err
	}
}

func (api *InwxApi) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	info, err := api.client.Domains.Info(dc.Name, 0)
	if err != nil {
		return nil, err
	}

	sort.Strings(info.Nameservers)
	foundNameservers := strings.Join(info.Nameservers, ",")
	expected := []string{}
	for _, ns := range dc.Nameservers {
		expected = append(expected, ns.Name)
	}
	sort.Strings(expected)
	expectedNameservers := strings.Join(expected, ",")

	if foundNameservers != expectedNameservers {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", foundNameservers, expectedNameservers),
				F:   api.updateNameservers(expected, dc.Name),
			},
		}, nil
	}
	return nil, nil
}
