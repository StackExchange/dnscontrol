package azuredns

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	adns "github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	aauth "github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/providers"
)

type azurednsProvider struct {
	zonesClient    *adns.ZonesClient
	recordsClient  *adns.RecordSetsClient
	zones          map[string]*adns.Zone
	resourceGroup  *string
	subscriptionID *string
}

func newAzureDNSDsp(conf map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newAzureDNS(conf, metadata)
}

func newAzureDNS(m map[string]string, metadata json.RawMessage) (*azurednsProvider, error) {
	subID, rg := m["SubscriptionID"], m["ResourceGroup"]

	zonesClient := adns.NewZonesClient(subID)
	recordsClient := adns.NewRecordSetsClient(subID)
	clientCredentialAuthorizer := aauth.NewClientCredentialsConfig(m["ClientID"], m["ClientSecret"], m["TenantID"])
	authorizer, authErr := clientCredentialAuthorizer.Authorizer()

	if authErr != nil {
		return nil, authErr
	}

	zonesClient.Authorizer = authorizer
	recordsClient.Authorizer = authorizer
	api := &azurednsProvider{zonesClient: &zonesClient, recordsClient: &recordsClient, resourceGroup: to.StringPtr(rg), subscriptionID: to.StringPtr(subID)}
	err := api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Cannot("Azure DNS does not provide a generic ALIAS functionality. Use AZURE_ALIAS instead."),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can("Azure does not permit modifying the existing NS records, only adding/removing additional records."),
	providers.DocOfficiallySupported: providers.Can(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAzureAlias:       providers.Can(),
}

func init() {
	providers.RegisterDomainServiceProviderType("AZURE_DNS", newAzureDNSDsp, features)
	providers.RegisterCustomRecordType("AZURE_ALIAS", "AZURE_DNS", "")
}

func (a *azurednsProvider) getExistingZones() (*adns.ZoneListResult, error) {
	// Please note — this function doesn't work with > 100 zones
	// https://github.com/StackExchange/dnscontrol/issues/792
	// Copied this code to getZones and ListZones and modified it for using a paging
	// As a result getExistingZones is not used anymore
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	zonesIterator, zonesErr := a.zonesClient.ListByResourceGroupComplete(ctx, *a.resourceGroup, to.Int32Ptr(100))
	if zonesErr != nil {
		return nil, zonesErr
	}
	zonesResult := zonesIterator.Response()
	return &zonesResult, nil
}

func (a *azurednsProvider) getZones() error {
	a.zones = make(map[string]*adns.Zone)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	zonesIterator, zonesErr := a.zonesClient.ListByResourceGroup(ctx, *a.resourceGroup, to.Int32Ptr(100))
	if zonesErr != nil {
		return fmt.Errorf("getZones: zonesErr: %w", zonesErr)
	}

	// Check getExistingZones and https://github.com/StackExchange/dnscontrol/issues/792 for the details
	for zonesIterator.NotDone() {
		zonesResult := zonesIterator.Response()
		for _, z := range *zonesResult.Value {
			zone := z
			domain := strings.TrimSuffix(*z.Name, ".")
			a.zones[domain] = &zone
		}
		zonesIterator.NextWithContext(ctx)
	}

	return nil
}

type errNoExist struct {
	domain string
}

func (e errNoExist) Error() string {
	return fmt.Sprintf("Domain %s not found in you Azure account", e.domain)
}

func (a *azurednsProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, ok := a.zones[domain]
	if !ok {
		return nil, errNoExist{domain}
	}

	var nss []string
	if zone.ZoneProperties != nil {
		for _, ns := range *zone.ZoneProperties.NameServers {
			nss = append(nss, ns)
		}
	}
	return models.ToNameserversStripTD(nss)
}

func (a *azurednsProvider) ListZones() ([]string, error) {
	var zones []string

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	zonesIterator, zonesErr := a.zonesClient.ListByResourceGroup(ctx, *a.resourceGroup, to.Int32Ptr(100))
	if zonesErr != nil {
		return nil, fmt.Errorf("ListZones: zonesErr: %w", zonesErr)
	}

	// Check getExistingZones and https://github.com/StackExchange/dnscontrol/issues/792 for the details
	for zonesIterator.NotDone() {
		zonesResult := zonesIterator.Response()
		for _, z := range *zonesResult.Value {
			domain := strings.TrimSuffix(*z.Name, ".")
			zones = append(zones, domain)
		}
		zonesIterator.NextWithContext(ctx)
	}

	return zones, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (a *azurednsProvider) GetZoneRecords(domain string) (models.Records, error) {
	existingRecords, _, _, err := a.getExistingRecords(domain)
	if err != nil {
		return nil, err
	}
	return existingRecords, nil
}

func (a *azurednsProvider) getExistingRecords(domain string) (models.Records, []*adns.RecordSet, string, error) {
	zone, ok := a.zones[domain]
	if !ok {
		return nil, nil, "", errNoExist{domain}
	}
	zoneName := *zone.Name
	records, err := a.fetchRecordSets(zoneName)
	if err != nil {
		return nil, nil, "", err
	}

	var existingRecords models.Records
	for _, set := range records {
		existingRecords = append(existingRecords, nativeToRecords(set, zoneName)...)
	}

	models.PostProcessRecords(existingRecords)
	return existingRecords, records, zoneName, nil
}

func (a *azurednsProvider) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	err := dc.Punycode()

	if err != nil {
		return nil, err
	}

	var corrections []*models.Correction

	existingRecords, records, zoneName, err := a.getExistingRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	differ := diff.New(dc)
	namesToUpdate, err := differ.ChangedGroups(existingRecords)
	if err != nil {
		return nil, err
	}

	if len(namesToUpdate) == 0 {
		return nil, nil
	}

	updates := map[models.RecordKey][]*models.RecordConfig{}

	for k := range namesToUpdate {
		updates[k] = nil
		for _, rc := range dc.Records {
			if rc.Key() == k {
				updates[k] = append(updates[k], rc)
			}
		}
	}

	for k, recs := range updates {
		if len(recs) == 0 {
			var rrset *adns.RecordSet
			for _, r := range records {
				if strings.TrimSuffix(*r.RecordSetProperties.Fqdn, ".") == k.NameFQDN {
					n1, err := nativeToRecordType(r.Type)
					if err != nil {
						return nil, err
					}
					n2, err := nativeToRecordType(to.StringPtr(k.Type))
					if err != nil {
						return nil, err
					}
					if n1 == n2 {
						rrset = r
						break
					}
				}
			}
			if rrset != nil {
				corrections = append(corrections,
					&models.Correction{
						Msg: strings.Join(namesToUpdate[k], "\n"),
						F: func() error {
							ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
							defer cancel()
							rt, err := nativeToRecordType(rrset.Type)
							if err != nil {
								return err
							}
							_, err = a.recordsClient.Delete(ctx, *a.resourceGroup, zoneName, *rrset.Name, rt, "")
							if err != nil {
								return err
							}
							return nil
						},
					})
			} else {
				return nil, fmt.Errorf("no record set found to delete. Name: '%s'. Type: '%s'", k.NameFQDN, k.Type)
			}
		} else {
			rrset, recordType, err := a.recordToNative(k, recs)
			if err != nil {
				return nil, err
			}
			var recordName string
			for _, r := range recs {
				i := int64(r.TTL)
				rrset.TTL = &i // TODO: make sure that ttls are consistent within a set
				recordName = r.Name
			}

			for _, r := range records {
				existingRecordType, err := nativeToRecordType(r.Type)
				if err != nil {
					return nil, err
				}
				changedRecordType, err := nativeToRecordType(to.StringPtr(k.Type))
				if err != nil {
					return nil, err
				}
				if strings.TrimSuffix(*r.RecordSetProperties.Fqdn, ".") == k.NameFQDN && (changedRecordType == adns.CNAME || existingRecordType == adns.CNAME) {
					if existingRecordType == adns.A || existingRecordType == adns.AAAA || changedRecordType == adns.A || changedRecordType == adns.AAAA { //CNAME cannot coexist with an A or AA
						corrections = append(corrections,
							&models.Correction{
								Msg: strings.Join(namesToUpdate[k], "\n"),
								F: func() error {
									ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
									defer cancel()
									_, err := a.recordsClient.Delete(ctx, *a.resourceGroup, zoneName, recordName, existingRecordType, "")
									if err != nil {
										return err
									}
									return nil
								},
							})
					}
				}
			}

			corrections = append(corrections,
				&models.Correction{
					Msg: strings.Join(namesToUpdate[k], "\n"),
					F: func() error {
						ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
						defer cancel()
						_, err := a.recordsClient.CreateOrUpdate(ctx, *a.resourceGroup, zoneName, recordName, recordType, *rrset, "", "")
						if err != nil {
							return err
						}
						return nil
					},
				})
		}
	}

	// Sort the records for cosmetic reasons: It just makes a long list
	// of deletes or adds easier to read if they are in sorted order.
	// That said, it may be risky to sort them (sort key is the text
	// message "Msg") if there are deletes that must happen before adds.
	// Reading the above code it isn't clear that any of the updates are
	// order-dependent.  That said, all the tests pass.
	// If in the future this causes a bug, we can either just remove
	// this next line, or (even better) put any order-dependent
	// operations in a single models.Correction{}.
	sort.Slice(corrections, func(i, j int) bool { return diff.CorrectionLess(corrections, i, j) })

	return corrections, nil
}

func nativeToRecordType(recordType *string) (adns.RecordType, error) {
	recordTypeStripped := strings.TrimPrefix(*recordType, "Microsoft.Network/dnszones/")
	switch recordTypeStripped {
	case "A", "AZURE_ALIAS_A":
		return adns.A, nil
	case "AAAA", "AZURE_ALIAS_AAAA":
		return adns.AAAA, nil
	case "CAA":
		return adns.CAA, nil
	case "CNAME", "AZURE_ALIAS_CNAME":
		return adns.CNAME, nil
	case "MX":
		return adns.MX, nil
	case "NS":
		return adns.NS, nil
	case "PTR":
		return adns.PTR, nil
	case "SRV":
		return adns.SRV, nil
	case "TXT":
		return adns.TXT, nil
	case "SOA":
		return adns.SOA, nil
	default:
		// Unimplemented type. Return adns.A as a decoy, but send an error.
		return adns.A, fmt.Errorf("rc.String rtype %v unimplemented", *recordType)
	}
}

func nativeToRecords(set *adns.RecordSet, origin string) []*models.RecordConfig {
	var results []*models.RecordConfig
	switch rtype := *set.Type; rtype {
	case "Microsoft.Network/dnszones/A":
		if set.ARecords != nil {
			for _, rec := range *set.ARecords {
				rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
				rc.SetLabelFromFQDN(*set.Fqdn, origin)
				rc.Type = "A"
				_ = rc.SetTarget(*rec.Ipv4Address)
				results = append(results, rc)
			}
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.TTL),
				AzureAlias: map[string]string{
					"type": "A",
				},
			}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			_ = rc.SetTarget(*set.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/AAAA":
		if set.AaaaRecords != nil {
			for _, rec := range *set.AaaaRecords {
				rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
				rc.SetLabelFromFQDN(*set.Fqdn, origin)
				rc.Type = "AAAA"
				_ = rc.SetTarget(*rec.Ipv6Address)
				results = append(results, rc)
			}
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.TTL),
				AzureAlias: map[string]string{
					"type": "AAAA",
				},
			}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			_ = rc.SetTarget(*set.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/CNAME":
		if set.CnameRecord != nil {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "CNAME"
			_ = rc.SetTarget(*set.CnameRecord.Cname)
			results = append(results, rc)
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.TTL),
				AzureAlias: map[string]string{
					"type": "CNAME",
				},
			}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			_ = rc.SetTarget(*set.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/NS":
		for _, rec := range *set.NsRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "NS"
			_ = rc.SetTarget(*rec.Nsdname)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/PTR":
		for _, rec := range *set.PtrRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "PTR"
			_ = rc.SetTarget(*rec.Ptrdname)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/TXT":
		if len(*set.TxtRecords) == 0 { // Empty String Record Parsing
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "TXT"
			_ = rc.SetTargetTXT("")
			results = append(results, rc)
		} else {
			for _, rec := range *set.TxtRecords {
				rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
				rc.SetLabelFromFQDN(*set.Fqdn, origin)
				rc.Type = "TXT"
				_ = rc.SetTargetTXTs(*rec.Value)
				results = append(results, rc)
			}
		}
	case "Microsoft.Network/dnszones/MX":
		for _, rec := range *set.MxRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "MX"
			_ = rc.SetTargetMX(uint16(*rec.Preference), *rec.Exchange)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/SRV":
		for _, rec := range *set.SrvRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "SRV"
			_ = rc.SetTargetSRV(uint16(*rec.Priority), uint16(*rec.Weight), uint16(*rec.Port), *rec.Target)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/CAA":
		for _, rec := range *set.CaaRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.TTL)}
			rc.SetLabelFromFQDN(*set.Fqdn, origin)
			rc.Type = "CAA"
			_ = rc.SetTargetCAA(uint8(*rec.Flags), *rec.Tag, *rec.Value)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/SOA":
	default:
		panic(fmt.Errorf("rc.String rtype %v unimplemented", *set.Type))
	}
	return results
}

func (a *azurednsProvider) recordToNative(recordKey models.RecordKey, recordConfig []*models.RecordConfig) (*adns.RecordSet, adns.RecordType, error) {
	recordSet := &adns.RecordSet{Type: to.StringPtr(recordKey.Type), RecordSetProperties: &adns.RecordSetProperties{}}
	for _, rec := range recordConfig {
		switch recordKey.Type {
		case "A":
			if recordSet.ARecords == nil {
				recordSet.ARecords = &[]adns.ARecord{}
			}
			*recordSet.ARecords = append(*recordSet.ARecords, adns.ARecord{Ipv4Address: to.StringPtr(rec.Target)})
		case "AAAA":
			if recordSet.AaaaRecords == nil {
				recordSet.AaaaRecords = &[]adns.AaaaRecord{}
			}
			*recordSet.AaaaRecords = append(*recordSet.AaaaRecords, adns.AaaaRecord{Ipv6Address: to.StringPtr(rec.Target)})
		case "CNAME":
			recordSet.CnameRecord = &adns.CnameRecord{Cname: to.StringPtr(rec.Target)}
		case "NS":
			if recordSet.NsRecords == nil {
				recordSet.NsRecords = &[]adns.NsRecord{}
			}
			*recordSet.NsRecords = append(*recordSet.NsRecords, adns.NsRecord{Nsdname: to.StringPtr(rec.Target)})
		case "PTR":
			if recordSet.PtrRecords == nil {
				recordSet.PtrRecords = &[]adns.PtrRecord{}
			}
			*recordSet.PtrRecords = append(*recordSet.PtrRecords, adns.PtrRecord{Ptrdname: to.StringPtr(rec.Target)})
		case "TXT":
			if recordSet.TxtRecords == nil {
				recordSet.TxtRecords = &[]adns.TxtRecord{}
			}
			// Empty TXT record needs to have no value set in it's properties
			if !(len(rec.TxtStrings) == 1 && rec.TxtStrings[0] == "") {
				*recordSet.TxtRecords = append(*recordSet.TxtRecords, adns.TxtRecord{Value: &rec.TxtStrings})
			}
		case "MX":
			if recordSet.MxRecords == nil {
				recordSet.MxRecords = &[]adns.MxRecord{}
			}
			*recordSet.MxRecords = append(*recordSet.MxRecords, adns.MxRecord{Exchange: to.StringPtr(rec.Target), Preference: to.Int32Ptr(int32(rec.MxPreference))})
		case "SRV":
			if recordSet.SrvRecords == nil {
				recordSet.SrvRecords = &[]adns.SrvRecord{}
			}
			*recordSet.SrvRecords = append(*recordSet.SrvRecords, adns.SrvRecord{Target: to.StringPtr(rec.Target), Port: to.Int32Ptr(int32(rec.SrvPort)), Weight: to.Int32Ptr(int32(rec.SrvWeight)), Priority: to.Int32Ptr(int32(rec.SrvPriority))})
		case "CAA":
			if recordSet.CaaRecords == nil {
				recordSet.CaaRecords = &[]adns.CaaRecord{}
			}
			*recordSet.CaaRecords = append(*recordSet.CaaRecords, adns.CaaRecord{Value: to.StringPtr(rec.Target), Tag: to.StringPtr(rec.CaaTag), Flags: to.Int32Ptr(int32(rec.CaaFlag))})
		case "AZURE_ALIAS_A", "AZURE_ALIAS_AAAA", "AZURE_ALIAS_CNAME":
			*recordSet.Type = rec.AzureAlias["type"]
			recordSet.TargetResource = &adns.SubResource{ID: to.StringPtr(rec.Target)}
		default:
			return nil, adns.A, fmt.Errorf("rc.String rtype %v unimplemented", recordKey.Type) // ands.A is a placeholder
		}
	}

	rt, err := nativeToRecordType(to.StringPtr(*recordSet.Type))
	if err != nil {
		return nil, adns.A, err // adns.A is a placeholder
	}
	return recordSet, rt, nil
}

func (a *azurednsProvider) fetchRecordSets(zoneName string) ([]*adns.RecordSet, error) {
	if zoneName == "" {
		return nil, nil
	}
	var records []*adns.RecordSet
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	recordsIterator, recordsErr := a.recordsClient.ListAllByDNSZone(ctx, *a.resourceGroup, zoneName, to.Int32Ptr(1000), "")
	if recordsErr != nil {
		return nil, recordsErr
	}

	for recordsIterator.NotDone() {
		recordsResult := recordsIterator.Response()
		for _, r := range *recordsResult.Value {
			record := r
			records = append(records, &record)
		}
		recordsIterator.NextWithContext(ctx)
	}

	return records, nil
}

func (a *azurednsProvider) EnsureDomainExists(domain string) error {
	if _, ok := a.zones[domain]; ok {
		return nil
	}
	fmt.Printf("Adding zone for %s to Azure dns account\n", domain)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	_, err := a.zonesClient.CreateOrUpdate(ctx, *a.resourceGroup, domain, adns.Zone{Location: to.StringPtr("global")}, "", "")
	if err != nil {
		return err
	}
	return nil
}
