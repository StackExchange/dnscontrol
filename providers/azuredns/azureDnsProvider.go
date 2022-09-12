package azuredns

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	aauth "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	adns "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/diff"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/pkg/txtutil"
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
	clientID, clientSecret, tenantID := m["ClientID"], m["ClientSecret"], m["TenantID"]
	credential, authErr := aauth.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if authErr != nil {
		return nil, authErr
	}
	zonesClient, zoneErr := adns.NewZonesClient(subID, credential, nil)
	if zoneErr != nil {
		return nil, zoneErr
	}
	recordsClient, recordErr := adns.NewRecordSetsClient(subID, credential, nil)
	if recordErr != nil {
		return nil, recordErr
	}

	api := &azurednsProvider{zonesClient: zonesClient, recordsClient: recordsClient, resourceGroup: to.StringPtr(rg), subscriptionID: to.StringPtr(subID)}
	err := api.getZones()
	if err != nil {
		return nil, err
	}
	return api, nil
}

var features = providers.DocumentationNotes{
	providers.CanGetZones:            providers.Can(),
	providers.CanUseAlias:            providers.Cannot("Azure DNS does not provide a generic ALIAS functionality. Use AZURE_ALIAS instead."),
	providers.CanUseAzureAlias:       providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can("Azure does not permit modifying the existing NS records, only adding/removing additional records."),
	providers.DocOfficiallySupported: providers.Can(),
}

func init() {
	fns := providers.DspFuncs{
		Initializer:   newAzureDNSDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType("AZURE_DNS", fns, features)
	providers.RegisterCustomRecordType("AZURE_ALIAS", "AZURE_DNS", "")
}

func (a *azurednsProvider) getExistingZones() ([]*adns.Zone, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()
	zonesPager := a.zonesClient.NewListByResourceGroupPager(*a.resourceGroup, nil)
	var zones []*adns.Zone
	for zonesPager.More() {
		nextResult, zonesErr := zonesPager.NextPage(ctx)
		if zonesErr != nil {
			return nil, zonesErr
		}
		zones = append(zones, nextResult.Value...)
	}
	return zones, nil
}

func (a *azurednsProvider) getZones() error {
	a.zones = make(map[string]*adns.Zone)

	zones, err := a.getExistingZones()
	if err != nil {
		return err
	}

	for _, z := range zones {
		zone := z
		domain := strings.TrimSuffix(*z.Name, ".")
		a.zones[domain] = zone
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
	if zone.Properties != nil {
		for _, ns := range zone.Properties.NameServers {
			nss = append(nss, *ns)
		}
	}

	return models.ToNameserversStripTD(nss)
}

func (a *azurednsProvider) ListZones() ([]string, error) {
	zonesResult, err := a.getExistingZones()
	if err != nil {
		return nil, err
	}
	var zones []string

	for _, z := range zonesResult {
		domain := strings.TrimSuffix(*z.Name, ".")
		zones = append(zones, domain)
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

	// FIXME(tlim): PostProcessRecords is usually called in GetDomainCorrections.
	models.PostProcessRecords(existingRecords)

	// FIXME(tlim): The "records" return value is usually stored in RecordConfig.Original.
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

	txtutil.SplitSingleLongTxt(dc.Records) // Autosplit long TXT records

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
				if strings.TrimSuffix(*r.Properties.Fqdn, ".") == k.NameFQDN {
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
							_, err = a.recordsClient.Delete(ctx, *a.resourceGroup, zoneName, *rrset.Name, rt, nil)
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
				rrset.Properties.TTL = &i // TODO: make sure that ttls are consistent within a set
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
				if strings.TrimSuffix(*r.Properties.Fqdn, ".") == k.NameFQDN && (changedRecordType == adns.RecordTypeCNAME || existingRecordType == adns.RecordTypeCNAME) {
					if existingRecordType == adns.RecordTypeA || existingRecordType == adns.RecordTypeAAAA || changedRecordType == adns.RecordTypeA || changedRecordType == adns.RecordTypeAAAA { //CNAME cannot coexist with an A or AA
						corrections = append(corrections,
							&models.Correction{
								Msg: strings.Join(namesToUpdate[k], "\n"),
								F: func() error {
									ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
									defer cancel()
									_, err := a.recordsClient.Delete(ctx, *a.resourceGroup, zoneName, recordName, existingRecordType, nil)
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
						_, err := a.recordsClient.CreateOrUpdate(ctx, *a.resourceGroup, zoneName, recordName, recordType, *rrset, nil)
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
		return adns.RecordTypeA, nil
	case "AAAA", "AZURE_ALIAS_AAAA":
		return adns.RecordTypeAAAA, nil
	case "CAA":
		return adns.RecordTypeCAA, nil
	case "CNAME", "AZURE_ALIAS_CNAME":
		return adns.RecordTypeCNAME, nil
	case "MX":
		return adns.RecordTypeMX, nil
	case "NS":
		return adns.RecordTypeNS, nil
	case "PTR":
		return adns.RecordTypePTR, nil
	case "SRV":
		return adns.RecordTypeSRV, nil
	case "TXT":
		return adns.RecordTypeTXT, nil
	case "SOA":
		return adns.RecordTypeSOA, nil
	default:
		// Unimplemented type. Return adns.A as a decoy, but send an error.
		return adns.RecordTypeA, fmt.Errorf("rc.String rtype %v unimplemented", *recordType)
	}
}

func nativeToRecords(set *adns.RecordSet, origin string) []*models.RecordConfig {
	var results []*models.RecordConfig
	switch rtype := *set.Type; rtype {
	case "Microsoft.Network/dnszones/A":
		if set.Properties.ARecords != nil {
			for _, rec := range set.Properties.ARecords {
				rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
				rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
				rc.Type = "A"
				_ = rc.SetTarget(*rec.IPv4Address)
				results = append(results, rc)
			}
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.Properties.TTL),
				AzureAlias: map[string]string{
					"type": "A",
				},
			}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			_ = rc.SetTarget(*set.Properties.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/AAAA":
		if set.Properties.AaaaRecords != nil {
			for _, rec := range set.Properties.AaaaRecords {
				rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
				rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
				rc.Type = "AAAA"
				_ = rc.SetTarget(*rec.IPv6Address)
				results = append(results, rc)
			}
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.Properties.TTL),
				AzureAlias: map[string]string{
					"type": "AAAA",
				},
			}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			_ = rc.SetTarget(*set.Properties.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/CNAME":
		if set.Properties.CnameRecord != nil {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "CNAME"
			_ = rc.SetTarget(*set.Properties.CnameRecord.Cname)
			results = append(results, rc)
		} else {
			rc := &models.RecordConfig{
				Type: "AZURE_ALIAS",
				TTL:  uint32(*set.Properties.TTL),
				AzureAlias: map[string]string{
					"type": "CNAME",
				},
			}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			_ = rc.SetTarget(*set.Properties.TargetResource.ID)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/NS":
		for _, rec := range set.Properties.NsRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "NS"
			_ = rc.SetTarget(*rec.Nsdname)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/PTR":
		for _, rec := range set.Properties.PtrRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "PTR"
			_ = rc.SetTarget(*rec.Ptrdname)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/TXT":
		if len(set.Properties.TxtRecords) == 0 { // Empty String Record Parsing
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "TXT"
			_ = rc.SetTargetTXT("")
			results = append(results, rc)
		} else {
			for _, rec := range set.Properties.TxtRecords {
				rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
				rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
				rc.Type = "TXT"
				var txts []string
				for _, txt := range rec.Value {
					txts = append(txts, *txt)
				}
				_ = rc.SetTargetTXTs(txts)
				results = append(results, rc)
			}
		}
	case "Microsoft.Network/dnszones/MX":
		for _, rec := range set.Properties.MxRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "MX"
			_ = rc.SetTargetMX(uint16(*rec.Preference), *rec.Exchange)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/SRV":
		for _, rec := range set.Properties.SrvRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
			rc.Type = "SRV"
			_ = rc.SetTargetSRV(uint16(*rec.Priority), uint16(*rec.Weight), uint16(*rec.Port), *rec.Target)
			results = append(results, rc)
		}
	case "Microsoft.Network/dnszones/CAA":
		for _, rec := range set.Properties.CaaRecords {
			rc := &models.RecordConfig{TTL: uint32(*set.Properties.TTL)}
			rc.SetLabelFromFQDN(*set.Properties.Fqdn, origin)
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
	recordSet := &adns.RecordSet{Type: to.StringPtr(recordKey.Type), Properties: &adns.RecordSetProperties{}}
	for _, rec := range recordConfig {
		switch recordKey.Type {
		case "A":
			if recordSet.Properties.ARecords == nil {
				recordSet.Properties.ARecords = []*adns.ARecord{}
			}
			recordSet.Properties.ARecords = append(recordSet.Properties.ARecords, &adns.ARecord{IPv4Address: to.StringPtr(rec.GetTargetField())})
		case "AAAA":
			if recordSet.Properties.AaaaRecords == nil {
				recordSet.Properties.AaaaRecords = []*adns.AaaaRecord{}
			}
			recordSet.Properties.AaaaRecords = append(recordSet.Properties.AaaaRecords, &adns.AaaaRecord{IPv6Address: to.StringPtr(rec.GetTargetField())})
		case "CNAME":
			recordSet.Properties.CnameRecord = &adns.CnameRecord{Cname: to.StringPtr(rec.GetTargetField())}
		case "NS":
			if recordSet.Properties.NsRecords == nil {
				recordSet.Properties.NsRecords = []*adns.NsRecord{}
			}
			recordSet.Properties.NsRecords = append(recordSet.Properties.NsRecords, &adns.NsRecord{Nsdname: to.StringPtr(rec.GetTargetField())})
		case "PTR":
			if recordSet.Properties.PtrRecords == nil {
				recordSet.Properties.PtrRecords = []*adns.PtrRecord{}
			}
			recordSet.Properties.PtrRecords = append(recordSet.Properties.PtrRecords, &adns.PtrRecord{Ptrdname: to.StringPtr(rec.GetTargetField())})
		case "TXT":
			if recordSet.Properties.TxtRecords == nil {
				recordSet.Properties.TxtRecords = []*adns.TxtRecord{}
			}
			// Empty TXT record needs to have no value set in it's properties
			if !(len(rec.TxtStrings) == 1 && rec.TxtStrings[0] == "") {
				var txts []*string
				for _, txt := range rec.TxtStrings {
					txts = append(txts, to.StringPtr(txt))
				}
				recordSet.Properties.TxtRecords = append(recordSet.Properties.TxtRecords, &adns.TxtRecord{Value: txts})
			}
		case "MX":
			if recordSet.Properties.MxRecords == nil {
				recordSet.Properties.MxRecords = []*adns.MxRecord{}
			}
			recordSet.Properties.MxRecords = append(recordSet.Properties.MxRecords, &adns.MxRecord{Exchange: to.StringPtr(rec.GetTargetField()), Preference: to.Int32Ptr(int32(rec.MxPreference))})
		case "SRV":
			if recordSet.Properties.SrvRecords == nil {
				recordSet.Properties.SrvRecords = []*adns.SrvRecord{}
			}
			recordSet.Properties.SrvRecords = append(recordSet.Properties.SrvRecords, &adns.SrvRecord{Target: to.StringPtr(rec.GetTargetField()), Port: to.Int32Ptr(int32(rec.SrvPort)), Weight: to.Int32Ptr(int32(rec.SrvWeight)), Priority: to.Int32Ptr(int32(rec.SrvPriority))})
		case "CAA":
			if recordSet.Properties.CaaRecords == nil {
				recordSet.Properties.CaaRecords = []*adns.CaaRecord{}
			}
			recordSet.Properties.CaaRecords = append(recordSet.Properties.CaaRecords, &adns.CaaRecord{Value: to.StringPtr(rec.GetTargetField()), Tag: to.StringPtr(rec.CaaTag), Flags: to.Int32Ptr(int32(rec.CaaFlag))})
		case "AZURE_ALIAS_A", "AZURE_ALIAS_AAAA", "AZURE_ALIAS_CNAME":
			*recordSet.Type = rec.AzureAlias["type"]
			recordSet.Properties.TargetResource = &adns.SubResource{ID: to.StringPtr(rec.GetTargetField())}
		default:
			return nil, adns.RecordTypeA, fmt.Errorf("rc.String rtype %v unimplemented", recordKey.Type) // ands.A is a placeholder
		}
	}

	rt, err := nativeToRecordType(to.StringPtr(*recordSet.Type))
	if err != nil {
		return nil, adns.RecordTypeA, err // adns.A is a placeholder
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
	recordsPager := a.recordsClient.NewListAllByDNSZonePager(*a.resourceGroup, zoneName, nil)

	for recordsPager.More() {
		nextResult, recordsErr := recordsPager.NextPage(ctx)
		if recordsErr != nil {
			return nil, recordsErr
		}
		records = append(records, nextResult.Value...)
	}

	return records, nil
}

func (a *azurednsProvider) EnsureDomainExists(domain string) error {
	if _, ok := a.zones[domain]; ok {
		return nil
	}
	printer.Printf("Adding zone for %s to Azure dns account\n", domain)

	ctx, cancel := context.WithTimeout(context.Background(), 6000*time.Second)
	defer cancel()

	_, err := a.zonesClient.CreateOrUpdate(ctx, *a.resourceGroup, domain, adns.Zone{Location: to.StringPtr("global")}, nil)
	if err != nil {
		return err
	}
	return nil
}
