package tencentdns

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	domain "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/domain/v20180808"
)

type tencentCloudClient struct {
	dnspodClient *dnspod.Client
	domainClient *domain.Client
}

func newClient(secretId, secretKey, region string) (*tencentCloudClient, error) {
	credential := common.NewCredential(secretId, secretKey)
	cpf := profile.NewClientProfile()
	
	// DNSPod client
	dpc, err := dnspod.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create dnspod client: %w", err)
	}

	// Domain client
	dmc, err := domain.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("failed to create domain client: %w", err)
	}

	return &tencentCloudClient{
		dnspodClient: dpc,
		domainClient: dmc,
	}, nil
}

func (c *tencentCloudClient) fetchRecords(domainName string) ([]*dnspod.RecordListItem, error) {
	var records []*dnspod.RecordListItem
	var offset uint64 = 0
	var limit uint64 = 1000

	for {
		request := dnspod.NewDescribeRecordListRequest()
		request.Domain = common.StringPtr(domainName)
		request.Offset = common.Uint64Ptr(offset)
		request.Limit = common.Uint64Ptr(limit)

		response, err := c.dnspodClient.DescribeRecordList(request)
		if err != nil {
			return nil, err
		}

		records = append(records, response.Response.RecordList...)

		if uint64(len(records)) >= *response.Response.RecordCountInfo.TotalCount {
			break
		}
		offset += limit
	}

	return records, nil
}

func (c *tencentCloudClient) getNameservers(domainName string) ([]string, error) {
	request := dnspod.NewDescribeDomainRequest()
	request.Domain = common.StringPtr(domainName)

	response, err := c.dnspodClient.DescribeDomain(request)
	if err != nil {
		return nil, err
	}

	var nss []string
	for _, ns := range response.Response.DomainInfo.DnspodNsList {
		nss = append(nss, *ns)
	}
	return nss, nil
}

func (c *tencentCloudClient) getRegistrarNameservers(domainName string) ([]string, error) {
	request := dnspod.NewDescribeDomainWhoisRequest()
	request.Domain = common.StringPtr(domainName)

	response, err := c.dnspodClient.DescribeDomainWhois(request)
	if err != nil {
		return nil, err
	}

	var nss []string
	for _, ns := range response.Response.Info.NameServers {
		nss = append(nss, *ns)
	}
	return nss, nil
}

func (c *tencentCloudClient) updateRegistrarNameservers(domainName string, nss []string) error {
	request := domain.NewModifyDomainDNSBatchRequest()
	request.Domains = common.StringPtrs([]string{domainName})
	request.Dns = common.StringPtrs(nss)

	_, err := c.domainClient.ModifyDomainDNSBatch(request)
	return err
}

func (c *tencentCloudClient) createRecord(domainName string, request *dnspod.CreateRecordRequest) error {
	request.Domain = common.StringPtr(domainName)
	_, err := c.dnspodClient.CreateRecord(request)
	return err
}

func (c *tencentCloudClient) modifyRecord(domainName string, request *dnspod.ModifyRecordRequest) error {
	request.Domain = common.StringPtr(domainName)
	_, err := c.dnspodClient.ModifyRecord(request)
	return err
}

func (c *tencentCloudClient) deleteRecord(domainName string, recordId uint64) error {
	request := dnspod.NewDeleteRecordRequest()
	request.Domain = common.StringPtr(domainName)
	request.RecordId = common.Uint64Ptr(recordId)
	_, err := c.dnspodClient.DeleteRecord(request)
	return err
}
