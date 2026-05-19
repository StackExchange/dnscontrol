package oracle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DNSControl/dnscontrol/v4/models"
	"github.com/DNSControl/dnscontrol/v4/pkg/diff2"
	"github.com/DNSControl/dnscontrol/v4/pkg/printer"
	"github.com/DNSControl/dnscontrol/v4/pkg/providers"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/dns"
	"github.com/oracle/oci-go-sdk/v65/example/helpers"
)

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanConcur:              providers.Unimplemented(),
	providers.CanGetZones:            providers.Can(),
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
	const providerName = "ORACLE"
	const providerMaintainer = "@kallsyms"
	fns := providers.DspFuncs{
		Initializer:   New,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

type oracleProvider struct {
	client      dns.DnsClient
	compartment string
}

// New creates a new provider for Oracle Cloud DNS.
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

	// Set default retry policy to handle 429 automatically
	defaultRetryPolicy := common.DefaultRetryPolicy()
	client.SetCustomClientConfiguration(common.CustomClientConfiguration{
		RetryPolicy: &defaultRetryPolicy,
	})

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

	for listResp.OpcNextPage != nil {
		listResp, err = o.client.ListZones(ctx, dns.ListZonesRequest{
			CompartmentId: &o.compartment,
			Page:          listResp.OpcNextPage,
		})
		if err != nil {
			return nil, err
		}

		for _, zone := range listResp.Items {
			zones = append(zones, *zone.Name)
		}
	}

	return zones, nil
}

// EnsureZoneExists creates a zone if it does not exist.
func (o *oracleProvider) EnsureZoneExists(domain string, metadata map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	getResp, err := o.client.GetZone(ctx, dns.GetZoneRequest{
		ZoneNameOrId:  &domain,
		CompartmentId: &o.compartment,
	})
	if err == nil {
		return nil
	}
	if getResp.RawResponse.StatusCode != http.StatusNotFound {
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
	getResp, err = o.client.GetZone(ctx, dns.GetZoneRequest{
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

	nss := make([]string, len(getResp.Nameservers))
	for i, ns := range getResp.Nameservers {
		nss[i] = *ns.Hostname
	}

	nssNoStrip, err := models.ToNameservers(nss)
	if err != nil {
		nssStrip, err := models.ToNameserversStripTD(nss)
		if err != nil {
			return nil, errors.New("could not determine if trailing dots should be stripped or not")
		}

		return nssStrip, nil
	}

	return nssNoStrip, nil
}

func (o *oracleProvider) GetZoneRecords(dc *models.DomainConfig) (models.Records, error) {
	zone := dc.Name

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
func (o *oracleProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
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

		if rec.GetLabel() == "@" && rec.TTL != 86400 {
			// printer.Warnf("Oracle Cloud forces TTL=86400 for NS records. Ignoring configured TTL of %d for %s\n", rec.TTL, recNS)
			rec.TTL = 86400
		}
	}

	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}
	if changes == nil {
		return nil, 0, nil
	}

	/*
		Oracle's API doesn't have a way to update an existing record.
		You can either update an existing RRSet, Domain (FQDN), or Zone in which you have to supply
		the entire desired state, or you can patch specifying ADD/REMOVE actions.
		Oracle's API is also increadibly slow, so updating individual RRSets is unbearably slow
		for any size zone.

		Using this method means we need to handle the add/delete functions manually in this function
		rather than passing a function attached to the correction like every other provider.
		This cannot be the most elegant way to handle this issue, but I have not come up with better yet...
	*/

	var corrections []*models.Correction

	ops := make([]dns.RecordOperation, 0, actualChangeCount)

	for _, change := range changes {
		switch change.Type {
			case diff2.REPORT:
				corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
			case diff2.CREATE:
				ops = append(ops, convertToRecordOperation(change.New[0], dns.RecordOperationOperationAdd))
				corrections = append(corrections, &models.Correction{
					Msg: change.MsgsJoined,
					F: func() error { return nil},
				})
			case diff2.DELETE:
				ops = append(ops, convertToRecordOperation(change.Old[0], dns.RecordOperationOperationRemove))
				corrections = append(corrections, &models.Correction{
					Msg: change.MsgsJoined,
					F: func() error { return nil},
				})
			case diff2.CHANGE:
				ops = append(ops, convertToRecordOperation(change.Old[0], dns.RecordOperationOperationRemove))
				ops = append(ops, convertToRecordOperation(change.New[0], dns.RecordOperationOperationAdd))
				corrections = append(corrections, &models.Correction{
					Msg: change.MsgsJoined,
					F: func() error { return nil},
				})
			default:
				panic(fmt.Sprintf("unhandled change.Type %s", change.Type))
		}
	}

	// Prepare batched operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	patchReq := dns.PatchZoneRecordsRequest{
		ZoneNameOrId:  &dc.Name,
		CompartmentId: &o.compartment,
	}

	// Send batched corrections
	for batchStart := 0; batchStart < len(ops); batchStart += 100 {
		batchEnd := min(batchStart+100, len(ops))
		patchReq.Items = ops[batchStart:batchEnd]

		_, err := o.client.PatchZoneRecords(ctx, patchReq)
		if err != nil {
			return nil, 0, err
		}
	}

	return corrections, actualChangeCount, nil
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
