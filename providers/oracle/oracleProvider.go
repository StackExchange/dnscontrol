package oracle

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/oracle/oci-go-sdk/v32/common"
	"github.com/oracle/oci-go-sdk/v32/dns"
	"github.com/oracle/oci-go-sdk/v32/example/helpers"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot(),
	providers.CanUseAlias:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDS:               providers.Cannot(), // should be supported, but getting 500s in tests
	providers.CanUseLOC:              providers.Unimplemented(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Can(),
	providers.CanUseTLSA:             providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("ORACLE", fns, features)
}

type oracleProvider struct {
	client      dns.DnsClient
	compartment string
}

// New creates a new provider for Oracle Cloud DNS
func New(settings map[string]string, _ json.RawMessage) (providers.DNSServiceProvider, error) {
	client, err := dns.NewDnsClientWithConfigurationProvider(common.NewRawConfigurationProvider(
		settings["tenancy_ocid"],
		settings["user_ocid"],
		settings["region"],
		settings["fingerprint"],
		settings["private_key"],
		nil,
	))
	if err != nil {
		return nil, err
	}

	return &oracleProvider{
		client:      client,
		compartment: settings["compartment"],
	}, nil
}

// ListZones lists the zones on this account.
func (o *oracleProvider) ListZones() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	listResp, err := o.client.ListZones(ctx, dns.ListZonesRequest{
		CompartmentId: &o.compartment,
	})
	if err != nil {
		return nil, err
	}

	zones := make([]string, len(listResp.Items))
	for i, zone := range listResp.Items {
		zones[i] = *zone.Name
	}
	return zones, nil
}

// EnsureZoneExists creates a zone if it does not exist
func (o *oracleProvider) EnsureZoneExists(domain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	getResp, err := o.client.GetZone(ctx, dns.GetZoneRequest{
		ZoneNameOrId:  &domain,
		CompartmentId: &o.compartment,
	})
	if err == nil {
		return nil
	}
	if getResp.RawResponse.StatusCode != 404 {
		return err
	}

	_, err = o.client.CreateZone(ctx, dns.CreateZoneRequest{
		CreateZoneDetails: dns.CreateZoneDetails{
			CompartmentId: &o.compartment,
			Name:          &domain,
			ZoneType:      dns.CreateZoneDetailsZoneTypePrimary,
		},
	})
	if err != nil {
		return err
	}

	// poll until the zone is ready
	pollUntilAvailable := func(r common.OCIOperationResponse) bool {
		if converted, ok := r.Response.(dns.GetZoneResponse); ok {
			return converted.LifecycleState != dns.ZoneLifecycleStateActive
		}
		return true
	}
	_, err = o.client.GetZone(ctx, dns.GetZoneRequest{
		ZoneNameOrId:    &domain,
		CompartmentId:   &o.compartment,
		RequestMetadata: helpers.GetRequestMetadataWithCustomizedRetryPolicy(pollUntilAvailable),
	})

	return err
}

func (o *oracleProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	getResp, err := o.client.GetZone(ctx, dns.GetZoneRequest{
		ZoneNameOrId:  &domain,
		CompartmentId: &o.compartment,
	})
	if err != nil {
		return nil, err
	}

	nss := make([]string, len(getResp.Zone.Nameservers))
	for i, ns := range getResp.Zone.Nameservers {
		nss[i] = *ns.Hostname
	}

	return models.ToNameservers(nss)
}

func (o *oracleProvider) GetZoneRecords(zone string, meta map[string]string) (models.Records, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	records := models.Records{}

	request := dns.GetZoneRecordsRequest{
		ZoneNameOrId:  &zone,
		CompartmentId: &o.compartment,
	}

	for {
		getResp, err := o.client.GetZoneRecords(ctx, request)
		if err != nil {
			return nil, err
		}

		for _, record := range getResp.Items {
			// Hide SOAs
			if *record.Rtype == "SOA" {
				continue
			}

			rc := &models.RecordConfig{
				Type:     *record.Rtype,
				TTL:      uint32(*record.Ttl),
				Original: record,
			}
			rc.SetLabelFromFQDN(*record.Domain, zone)

			switch rc.Type {
			case "ALIAS":
				err = rc.SetTarget(*record.Rdata)
			default:
				err = rc.PopulateFromString(*record.Rtype, *record.Rdata, zone)
			}

			if err != nil {
				return nil, err
			}

			records = append(records, rc)
		}

		if getResp.OpcNextPage == nil {
			break
		}

		request.Page = getResp.OpcNextPage
	}

	return records, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (o *oracleProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, error) {
	var err error

	// Ensure we don't emit changes for attempted modification of built-in apex NSs
	for _, rec := range dc.Records {
		if rec.Type != "NS" {
			continue
		}

		recNS := rec.GetTargetField()
		if rec.GetLabel() == "@" && strings.HasSuffix(recNS, "dns.oraclecloud.com.") {
			printer.Warnf("Oracle Cloud does not allow changes to built-in apex NS records. Ignoring change to %s...\n", recNS)
			continue
		}

		if rec.TTL != 86400 {
			printer.Warnf("Oracle Cloud forces TTL=86400 for NS records. Ignoring configured TTL of %d for %s\n", rec.TTL, recNS)
			rec.TTL = 86400
		}
	}

	toReport, create, dels, modify, err := diff.NewCompat(dc).IncrementalDiff(existingRecords)
	if err != nil {
		return nil, err
	}
	// Start corrections with the reports
	corrections := diff.GenerateMessageCorrections(toReport)

	/*
		Oracle's API doesn't have a way to update an existing record.
		You can either update an existing RRSet, Domain (FQDN), or Zone in which you have to supply
		the entire desired state, or you can patch specifying ADD/REMOVE actions.
		Oracle's API is also increadibly slow, so updating individual RRSets is unbearably slow
		for any size zone.
	*/

	desc := ""
	createRecords := models.Records{}
	deleteRecords := models.Records{}

	if len(create) > 0 {
		for _, rec := range create {
			createRecords = append(createRecords, rec.Desired)
			desc += rec.String() + "\n"
		}
		desc = desc[:len(desc)-1]
	}

	if len(dels) > 0 {
		for _, rec := range dels {
			deleteRecords = append(deleteRecords, rec.Existing)
			desc += rec.String() + "\n"
		}
		desc = desc[:len(desc)-1]
	}

	if len(modify) > 0 {
		for _, rec := range modify {
			createRecords = append(createRecords, rec.Desired)
			deleteRecords = append(deleteRecords, rec.Existing)
			desc += rec.String() + "\n"
		}
		desc = desc[:len(desc)-1]
	}

	// There were corrections. Send them as one big batch:
	if len(createRecords) > 0 || len(deleteRecords) > 0 {
		corrections = append(corrections, &models.Correction{
			Msg: desc,
			F: func() error {
				return o.patch(createRecords, deleteRecords, dc.Name)
			},
		})
	}

	return corrections, nil
}

func (o *oracleProvider) patch(createRecords, deleteRecords models.Records, domain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	patchReq := dns.PatchZoneRecordsRequest{
		ZoneNameOrId:  &domain,
		CompartmentId: &o.compartment,
	}

	ops := make([]dns.RecordOperation, 0, len(createRecords)+len(deleteRecords))

	for _, rec := range deleteRecords {
		ops = append(ops, convertToRecordOperation(rec, dns.RecordOperationOperationRemove))
	}
	for _, rec := range createRecords {
		ops = append(ops, convertToRecordOperation(rec, dns.RecordOperationOperationAdd))
	}

	for batchStart := 0; batchStart < len(ops); batchStart += 100 {
		batchEnd := batchStart + 100
		if batchEnd > len(ops) {
			batchEnd = len(ops)
		}
		patchReq.Items = ops[batchStart:batchEnd]
		_, err := o.client.PatchZoneRecords(ctx, patchReq)
		if err != nil {
			return err
		}
	}

	return nil
}

func convertToRecordOperation(rec *models.RecordConfig, op dns.RecordOperationOperationEnum) dns.RecordOperation {
	if rec.Original != nil {
		return dns.RecordOperation{
			RecordHash: rec.Original.(dns.Record).RecordHash,
			Operation:  op,
		}
	}

	fqdn := rec.GetLabelFQDN()
	rtype := rec.Type
	rdata := rec.GetTargetCombined()
	ttl := int(rec.TTL)

	return dns.RecordOperation{
		Domain:    &fqdn,
		Rtype:     &rtype,
		Rdata:     &rdata,
		Ttl:       &ttl,
		Operation: op,
	}
}
