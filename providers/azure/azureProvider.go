package azure

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

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
	authorizer         *autorest.BearerAuthorizer
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

	azureConfig.authorizer = autorest.NewBearerAuthorizer(token)
	azureConfig.zonesClient = dns.NewZonesClient(azureConfig.subscriptionID)
	azureConfig.zonesClient.Authorizer = azureConfig.authorizer
	azureConfig.recordsClient = dns.NewRecordSetsClient(azureConfig.subscriptionID)
	azureConfig.recordsClient.Authorizer = azureConfig.authorizer

	return azureConfig, nil
}

func (c *azureConfig) GetDomainCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	existingRecords, err := c.GetZoneRecords(dc.Name)
	if err != nil {
		return nil, err
	}

	models.PostProcessRecords(existingRecords)
	differ := diff.New(dc)
	_, create, _, _ := differ.IncrementalDiff(existingRecords)

	var corrections = []*models.Correction{}

	for _, cs := range create {
		record := RCtoAZRecord(cs.Desired)

		corr := &models.Correction{
			Msg: cs.String(),
			F: func() error {
				_, err := c.recordsClient.CreateOrUpdate(ctx, c.resouceGroupName, dc.Name, "", dns.RecordType(*record.Type), *record, "", "")
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
	list, err := c.recordsClient.ListAllByDNSZoneComplete(ctx, c.resouceGroupName, zoneName, &itemLimit, "")
	if err != nil {
		return nil, err
	}

	var records models.Records

	for list.NextWithContext(ctx) == nil {
		if list.Value().Name == nil {
			break
		}

		recordType := strings.Replace(*list.Value().Type, "Microsoft.Network/dnszones/", "", -1)

		switch recordType {
		case "A":
			for _, a := range *list.Value().ARecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *a.Ipv4Address, zoneName)
				records = append(records, thisRecord)
			}
		case "AAAA":
			for _, aaaa := range *list.Value().AaaaRecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *aaaa.Ipv6Address, zoneName)
				records = append(records, thisRecord)
			}
		case "CNAME":
			thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
			thisRecord.PopulateFromString(thisRecord.Type, *list.Value().CnameRecord.Cname, zoneName)
			records = append(records, thisRecord)
		case "NS":
			for _, ns := range *list.Value().NsRecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
				thisRecord.PopulateFromString(thisRecord.Type, *ns.Nsdname, zoneName)
				records = append(records, thisRecord)
			}
		case "TXT":
			for _, txt := range *list.Value().TxtRecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
				thisRecord.SetTargetTXTs(*txt.Value)
				records = append(records, thisRecord)
			}
		case "MX":
			for _, mx := range *list.Value().MxRecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
				thisRecord.SetTargetMX(uint16(*mx.Preference), *mx.Exchange)
				records = append(records, thisRecord)
			}
		case "SOA":
			continue
		case "SRV":
			for _, srv := range *list.Value().SrvRecords {
				thisRecord := newRecord(recordType, *list.Value().Fqdn, zoneName, uint32(*list.Value().TTL))
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

// RCtoAZRecord converts a DNSControl RecordConfig record to an Azure RecordSet record
func RCtoAZRecord(rc *models.RecordConfig) *dns.RecordSet {
	thisRecord := dns.RecordSet{
		RecordSetProperties: &dns.RecordSetProperties{
			TTL: to.Int64Ptr(int64(rc.TTL)),
		},
	}

	switch rc.Type {
	case "A":
		thisRecord.ARecords = &[]dns.ARecord{
			dns.ARecord{Ipv4Address: to.StringPtr(rc.Target)},
		}
	case "AAAA":
		thisRecord.AaaaRecords = &[]dns.AaaaRecord{
			dns.AaaaRecord{Ipv6Address: to.StringPtr(rc.Target)},
		}
	case "CNAME":
		thisRecord.CnameRecord = &dns.CnameRecord{
			Cname: to.StringPtr(rc.Target),
		}
	case "NS":
		thisRecord.NsRecords = &[]dns.NsRecord{
			dns.NsRecord{Nsdname: to.StringPtr(rc.Target)},
		}
	case "TXT":
		thisRecord.TxtRecords = &[]dns.TxtRecord{
			dns.TxtRecord{Value: to.StringSlicePtr([]string{rc.Target})},
		}
	case "MX":
		thisRecord.MxRecords = &[]dns.MxRecord{
			dns.MxRecord{
				Preference: to.Int32Ptr(int32(rc.MxPreference)),
				Exchange:   to.StringPtr(rc.Target),
			},
		}
	case "SOA":
		return nil
	case "SRV":
		thisRecord.SrvRecords = &[]dns.SrvRecord{
			dns.SrvRecord{
				Port:     to.Int32Ptr(int32(rc.SrvPort)),
				Priority: to.Int32Ptr(int32(rc.SrvPriority)),
				Weight:   to.Int32Ptr(int32(rc.SrvWeight)),
				Target:   to.StringPtr(rc.Target),
			},
		}
	}

	return &thisRecord
}
