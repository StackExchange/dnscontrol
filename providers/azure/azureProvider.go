package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/diff"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
)

var features = providers.DocumentationNotes{
	providers.DocCreateDomains:       providers.Cannot(),
	providers.DocDualHost:            providers.Cannot("Azure does not permit modification of the NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUsePTR:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseCAA:              providers.Cannot(),
	providers.CanUseTXTMulti:         providers.Can(),
	providers.CanUseNAPTR:            providers.Cannot(),
	providers.CanUseSSHFP:            providers.Cannot(),
}

var ctx = context.Background()

type azureConfig struct {
	tenantID           string
	subscriptionID     string
	resouceGroupName   string
	clientID           string
	clientSecret       string
	resourceManagerURL string
	resourceManager    *azureResourceManager
	zonesClient        dns.ZonesClient
	recordsClient      dns.RecordSetsClient
}

type azureResourceManager struct {
	GalleryEndpoint string `json:"galleryEndpoint"`
	GraphEndpoint   string `json:"graphEndpoint"`
	PortalEndpoint  string `json:"portalEndpoint"`
	Authentication  struct {
		LoginEndpoint string   `json:"loginEndpoint"`
		Audiences     []string `json:"audiences"`
	} `json:"authentication"`
}

func init() {
	providers.RegisterDomainServiceProviderType("AZURE", New, features)
}

// New creates a new instance of the Azure DNS provider for DNSControl
func New(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	azureConfig := &azureConfig{
		tenantID:           config["tenantID"],
		subscriptionID:     config["subscriptionID"],
		resouceGroupName:   config["resouceGroupName"],
		clientID:           config["clientID"],
		clientSecret:       config["clientSecret"],
		resourceManagerURL: config["resourceManagerURL"],
	}

	// If a resourceManagerURL is provided, then we should use that value. This in theory should support requests for non-public Azure clouds
	// (such as China, Government, etc). However it's fairly untested as we have nothing except public Azure to test against.
	// If no URL is provided, then just use the defaults as provided by Azure
	var rm azureResourceManager
	if azureConfig.resourceManagerURL == "" {
		defaultJSON := "{\"galleryEndpoint\":\"https://gallery.azure.com/\",\"graphEndpoint\":\"https://graph.windows.net/\",\"portalEndpoint\":\"https://portal.azure.com/\",\"authentication\":{\"loginEndpoint\":\"https://login.windows.net/\",\"audiences\":[\"https://management.core.windows.net/\",\"https://management.azure.com/\"]}}"
		err := json.Unmarshal([]byte(defaultJSON), &rm)
		if err != nil {
			return nil, err
		}
	} else {
		resp, err := http.Get(azureConfig.resourceManagerURL)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, &rm)
		if err != nil {
			return nil, err
		}
	}

	azureConfig.resourceManager = &rm

	// Authenticate to Azure using the Oauth flow with a client id & secret. This could be expanded to support managed service
	// identities.
	var token adal.OAuthTokenProvider
	oauthConfig, err := adal.NewOAuthConfig(azureConfig.resourceManager.Authentication.LoginEndpoint, azureConfig.tenantID)
	if err != nil {
		return nil, err
	}
	token, err = adal.NewServicePrincipalToken(
		*oauthConfig,
		azureConfig.clientID,
		azureConfig.clientSecret,
		azureConfig.resourceManager.Authentication.Audiences[0],
	)
	if err != nil {
		return nil, err
	}

	azureConfig.zonesClient = dns.NewZonesClient(azureConfig.subscriptionID)
	azureConfig.zonesClient.Authorizer = autorest.NewBearerAuthorizer(token)
	azureConfig.recordsClient = dns.NewRecordSetsClient(azureConfig.subscriptionID)
	azureConfig.recordsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return azureConfig, nil
}

func (c *azureConfig) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	if err := dc.Punycode(); err != nil {
		return nil, err
	}

	existingRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	var corrections = []*models.Correction{}

	checkNSModifications(dc)
	models.PostProcessRecords(existingRecords)
	differ := diff.New(dc)

	namesToUpdate := differ.ChangedGroups(existingRecords)
	if len(namesToUpdate) == 0 {
		return nil, nil
	}

	//fmt.Println("----", namesToUpdate)

	updates := map[models.RecordKey][]*models.RecordConfig{}
	// for each name we need to update, collect relevant records from our desired domain state
	for k := range namesToUpdate {
		if k.Type == "NS" && k.NameFQDN == dc.Name {
			//printer.Warnf("azure does not support modifying NS records on base domain. @ will not be modified.\n")
		} else {
			updates[k] = nil
			for _, rc := range dc.Records {
				if rc.Key() == k {
					updates[k] = append(updates[k], rc)
				}
			}
		}
	}

	dels := make(map[string]string)

	for k, recs := range updates {
		if len(recs) == 0 {
			dels[k.NameFQDN] = k.Type
		}
		for _, ex := range existingRecords {
			if k.NameFQDN == ex.NameFQDN && (k.Type == "CNAME" || ex.Type == "CNAME") {
				if ex.Type == "A" || ex.Type == "AAAA" || k.Type == "A" || k.Type == "AAAA" { //Cnames cannot coexist with an A or AA
					fmt.Println("---- found mismatched type to delete", k.NameFQDN, k.Type, ex.Type)
					dels[k.NameFQDN] = ex.Type
				}
			}
		}

	}

	for name, recordtype := range dels {
		localname := strings.Replace(name, fmt.Sprintf(".%s", dc.Name), "", -1)
		if name == dc.Name {
			localname = "@"
		}

		if recordtype == "NS" && localname == "@" {
			//printer.Warnf("azure does not support modifying NS records on base domain. @ will not be deleted.\n")
		} else {
			msg := fmt.Sprintf("delete %s %s", recordtype, localname)
			fmt.Println("----", msg)

			corr := &models.Correction{
				Msg: msg,
				F: func() error {
					_, err := c.recordsClient.Delete(ctx, c.resouceGroupName, dc.Name, localname, dns.RecordType(recordtype), "")
					// Artifically slow things down after a delete, as the API can take time to register it. The tests fail if we delete and then recheck too quickly.
					time.Sleep(2 * time.Second)
					return err
				},
			}

			corrections = append(corrections, corr)
		}
	}

	for _, recs := range updates {
		azureRecord := RStoAZRecord(recs)
		if azureRecord == nil {
			continue
		}

		var msg string
		for _, rec := range recs {
			msg += fmt.Sprintf("update %s %s %s\n", rec.Name, rec.Type, rec.GetTargetCombined())
		}

		fmt.Println("----", msg)

		recordType := dns.RecordType(*azureRecord.Type)
		corr := &models.Correction{
			Msg: msg,
			F: func() error {
				_, err := c.recordsClient.CreateOrUpdate(ctx, c.resouceGroupName, dc.Name, *azureRecord.Name, recordType, *azureRecord, "", "")
				return err
			},
		}

		corrections = append(corrections, corr)
	}

	return corrections, nil
}

func (c *azureConfig) GetNameservers(domain string) ([]*models.Nameserver, error) {
	zone, err := c.zonesClient.Get(ctx, c.resouceGroupName, domain)
	if err != nil {
		return nil, err
	}

	var nsList []*models.Nameserver
	for _, ns := range *zone.NameServers {
		nsList = append(nsList, &models.Nameserver{
			Name: ns,
		})
	}

	return nsList, nil
}

func (c *azureConfig) GetZoneRecords(zoneName string) (models.Records, error) {
	itemLimit := int32(1000) //this is an int32 but Azure only supports a maximum of 1,000 records. Hope you don't have more than that.
	list, err := c.recordsClient.ListByDNSZone(ctx, c.resouceGroupName, zoneName, &itemLimit, "")
	if err != nil {
		return nil, err
	}

	var records models.Records

	for list.NotDone() == false {
		list.NextWithContext(ctx)
	}

	for _, record := range list.Values() {
		recordType := strings.Replace(*record.Type, "Microsoft.Network/dnszones/", "", -1)

		switch recordType {
		case "A":
			for _, a := range *record.ARecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *a.Ipv4Address, zoneName)
				records = append(records, thisRecord)
			}
		case "AAAA":
			for _, aaaa := range *record.AaaaRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *aaaa.Ipv6Address, zoneName)
				records = append(records, thisRecord)
			}
		case "CNAME":
			thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
			thisRecord.PopulateFromString(thisRecord.Type, *record.CnameRecord.Cname, zoneName)
			records = append(records, thisRecord)
		case "NS":
			for _, ns := range *record.NsRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *ns.Nsdname, zoneName)
				records = append(records, thisRecord)
			}
		case "TXT":
			for _, txt := range *record.TxtRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.SetTargetTXTs(*txt.Value)
				records = append(records, thisRecord)
			}
		case "MX":
			for _, mx := range *record.MxRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.SetTargetMX(uint16(*mx.Preference), *mx.Exchange)
				records = append(records, thisRecord)
			}
		case "PTR":
			for _, ptr := range *record.PtrRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *ptr.Ptrdname, zoneName)
				records = append(records, thisRecord)
			}
		case "SOA":
			continue
		case "SRV":
			for _, srv := range *record.SrvRecords {
				thisRecord := newRecord(recordType, *record.Fqdn, zoneName, uint32(*record.TTL))
				thisRecord.SetTargetSRV(uint16(*srv.Priority), uint16(*srv.Weight), uint16(*srv.Port), *srv.Target)
				records = append(records, thisRecord)
			}
		}
	}

	return records, nil
}

func newRecord(recordType, fqdn, zoneName string, ttl uint32) *models.RecordConfig {
	thisRecord := &models.RecordConfig{
		Type: recordType,
		TTL:  ttl,
	}
	thisRecord.SetLabelFromFQDN(fqdn, zoneName)
	return thisRecord
}

// RStoAZRecord converts a DNS Control RecordSet to an Azure RecordSet
func RStoAZRecord(rs []*models.RecordConfig) *dns.RecordSet {
	if len(rs) == 0 {
		return nil
	}

	thisRecord := dns.RecordSet{
		Name: to.StringPtr(rs[0].Name),
		Type: to.StringPtr(rs[0].Type),
		RecordSetProperties: &dns.RecordSetProperties{
			TTL: to.Int64Ptr(int64(rs[0].TTL)),
		},
	}

	for _, cs := range rs {
		switch *thisRecord.Type {
		case "A":
			if thisRecord.ARecords == nil {
				thisRecord.ARecords = &[]dns.ARecord{}
			}
			*thisRecord.ARecords = append(*thisRecord.ARecords, dns.ARecord{Ipv4Address: to.StringPtr(cs.Target)})
		case "AAAA":
			if thisRecord.AaaaRecords == nil {
				thisRecord.AaaaRecords = &[]dns.AaaaRecord{}
			}
			*thisRecord.AaaaRecords = append(*thisRecord.AaaaRecords, dns.AaaaRecord{Ipv6Address: to.StringPtr(cs.Target)})
		case "CNAME":
			thisRecord.CnameRecord = &dns.CnameRecord{
				Cname: to.StringPtr(cs.Target),
			}
		case "NS":
			if thisRecord.NsRecords == nil {
				thisRecord.NsRecords = &[]dns.NsRecord{}
			}
			*thisRecord.NsRecords = append(*thisRecord.NsRecords, dns.NsRecord{Nsdname: to.StringPtr(cs.Target)})
		case "TXT":
			if thisRecord.NsRecords == nil {
				thisRecord.NsRecords = &[]dns.NsRecord{}
			}
			*thisRecord.TxtRecords = append(*thisRecord.TxtRecords, dns.TxtRecord{Value: to.StringSlicePtr([]string{cs.Target})})
		case "MX":
			if thisRecord.MxRecords == nil {
				thisRecord.MxRecords = &[]dns.MxRecord{}
			}
			*thisRecord.MxRecords = append(*thisRecord.MxRecords, dns.MxRecord{
				Preference: to.Int32Ptr(int32(cs.MxPreference)),
				Exchange:   to.StringPtr(cs.Target),
			})
		case "PTR":
			if thisRecord.PtrRecords == nil {
				thisRecord.PtrRecords = &[]dns.PtrRecord{}
			}
			*thisRecord.PtrRecords = append(*thisRecord.PtrRecords, dns.PtrRecord{
				Ptrdname: to.StringPtr(cs.Target),
			})
		case "SOA":
			return nil
		case "SRV":
			if thisRecord.SrvRecords == nil {
				thisRecord.SrvRecords = &[]dns.SrvRecord{}
			}
			*thisRecord.SrvRecords = append(*thisRecord.SrvRecords, dns.SrvRecord{
				Port:     to.Int32Ptr(int32(cs.SrvPort)),
				Priority: to.Int32Ptr(int32(cs.SrvPriority)),
				Weight:   to.Int32Ptr(int32(cs.SrvWeight)),
				Target:   to.StringPtr(cs.Target),
			})
		}
	}

	return &thisRecord
}

// Stolen from the Cloudflare provider
func checkNSModifications(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		if rec.Type == "NS" && rec.GetLabelFQDN() == dc.Name {
			//printer.Warnf("azure does not support modifying NS records on base domain. %s will not be modified.\n", rec.GetTargetField())
			continue
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
