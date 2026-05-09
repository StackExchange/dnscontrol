package tencentdns

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/DNSControl/dnscontrol/v4/models"
	"github.com/DNSControl/dnscontrol/v4/pkg/diff2"
	"github.com/DNSControl/dnscontrol/v4/pkg/providers"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var features = providers.DocumentationNotes{
	providers.CanUseAlias:            providers.Can(),
	providers.CanGetZones:            providers.Can(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseSRV:              providers.Can(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Can("Tencent Cloud allows full management of apex NS records"),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "TENCENTDNS"
	const providerMaintainer = ""
	fns := providers.DspFuncs{
		Initializer:   newTencentDNSDsp,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterRegistrarType(providerName, newTencentDNSReg)
	providers.RegisterMaintainer(providerName, providerMaintainer)
	// Default TTL for Tencent Cloud DNSPod is 600 for free domains.
	providers.RegisterDefaultTTL(providerName, 600)
}

type tencentdnsProvider struct {
	client *tencentCloudClient
}

func newTencentDNSDsp(config map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	return newTencentDNS(config)
}

func newTencentDNSReg(config map[string]string) (providers.Registrar, error) {
	return newTencentDNS(config)
}

func newTencentDNS(config map[string]string) (*tencentdnsProvider, error) {
	secretId := config["secret_id"]
	secretKey := config["secret_key"]
	if secretId == "" || secretKey == "" {
		return nil, fmt.Errorf("missing tencent cloud credentials (secret_id, secret_key)")
	}

	region := config["region"]
	if region == "" {
		region = "ap-guangzhou" // Default region
	}

	client, err := newClient(secretId, secretKey, region)
	if err != nil {
		return nil, err
	}

	return &tencentdnsProvider{
		client: client,
	}, nil
}

func (p *tencentdnsProvider) ListZones() ([]string, error) {
	// For simplicity, we just use the API to list all domains.
	// In a real implementation, we might want to handle pagination better.
	request := dnspod.NewDescribeDomainListRequest()
	response, err := p.client.dnspodClient.DescribeDomainList(request)
	if err != nil {
		return nil, err
	}

	var zones []string
	for _, domain := range response.Response.DomainList {
		zones = append(zones, *domain.Name)
	}
	return zones, nil
}

func (p *tencentdnsProvider) GetNameservers(domainName string) ([]*models.Nameserver, error) {
	nss, err := p.client.getNameservers(domainName)
	if err != nil {
		if strings.Contains(err.Error(), "DomainNotExists") || strings.Contains(err.Error(), "域名有误") {
			return nil, nil
		}
		return nil, err
	}
	return models.ToNameservers(nss)
}

func (p *tencentdnsProvider) GetZoneRecords(dc *models.DomainConfig) (models.Records, error) {
	records, err := p.client.fetchRecords(dc.Name)
	if err != nil {
		if strings.Contains(err.Error(), "DomainNotExists") {
			return nil, nil
		}
		return nil, err
	}

	existingRecords := models.Records{}
	for _, r := range records {
		if *r.Status != "ENABLE" {
			continue
		}
		rc, err := nativeToRecord(r, dc.Name)
		if err != nil {
			return nil, err
		}
		existingRecords = append(existingRecords, rc)
	}
	return existingRecords, nil
}

func (p *tencentdnsProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	// Tencent Cloud is a "ByRecord" API.
	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, change := range changes {
		msgs := change.MsgsJoined
		domainName := dc.Name

		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: msgs})
		case diff2.CREATE:
			rc := change.New[0]
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return p.client.createRecord(domainName, recordToCreateRequest(rc))
				},
			})
		case diff2.CHANGE:
			rc := change.New[0]
			recordId := *(change.Old[0].Original.(*dnspod.RecordListItem).RecordId)
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return p.client.modifyRecord(domainName, recordToModifyRequest(rc, recordId))
				},
			})
		case diff2.DELETE:
			recordId := *(change.Old[0].Original.(*dnspod.RecordListItem).RecordId)
			corrections = append(corrections, &models.Correction{
				Msg: msgs,
				F: func() error {
					return p.client.deleteRecord(domainName, recordId)
				},
			})
		}
	}

	return corrections, actualChangeCount, nil
}

func (p *tencentdnsProvider) GetRegistrarCorrections(dc *models.DomainConfig) ([]*models.Correction, error) {
	actualSet, err := p.client.getRegistrarNameservers(dc.Name)
	if err != nil {
		return nil, err
	}
	sort.Strings(actualSet)
	actual := strings.Join(actualSet, ",")

	expectedSet := []string{}
	for _, ns := range dc.Nameservers {
		expectedSet = append(expectedSet, ns.Name)
	}
	sort.Strings(expectedSet)
	expected := strings.Join(expectedSet, ",")

	if actual != expected {
		return []*models.Correction{
			{
				Msg: fmt.Sprintf("Update nameservers %s -> %s", actual, expected),
				F: func() error {
					return p.client.updateRegistrarNameservers(dc.Name, expectedSet)
				},
			},
		}, nil
	}

	return nil, nil
}

func (p *tencentdnsProvider) EnsureZoneExists(domainName string, metadata map[string]string) error {
	request := dnspod.NewCreateDomainRequest()
	request.Domain = &domainName
	_, err := p.client.dnspodClient.CreateDomain(request)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return nil
		}
		return err
	}
	return nil
}
