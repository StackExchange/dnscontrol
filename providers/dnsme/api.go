package dnsme

import (
	"fmt"
	"net/http"
	"time"
)

type dnsmeProvider struct {
	restAPI *dnsmeRestAPI
	domains map[string]multiDomainResponseDataEntry
}

func newDnsmeProvider(apiKey string, secretKey string, sandbox bool, debug bool) *dnsmeProvider {
	fmt.Println("creating DNSME provider for sandbox")

	baseURL := baseURLV2_0
	if sandbox {
		baseURL = sandboxBaseURLV2_0
	}

	return &dnsmeProvider{
		restAPI: &dnsmeRestAPI{
			apiKey:    apiKey,
			secretKey: secretKey,
			baseURL:   baseURL,
			httpClient: &http.Client{
				Timeout: time.Minute,
			},
			dumpHTTPRequest:  debug,
			dumpHTTPResponse: debug,
		},
	}
}

func (api *dnsmeProvider) loadDomains() error {
	if api.domains != nil {
		return nil
	}

	domains := map[string]multiDomainResponseDataEntry{}

	res, err := api.restAPI.multiDomainGet()
	if err != nil {
		return fmt.Errorf("fetching domains from DNSME failed: %w", err)
	}

	for _, domain := range res.Data {
		if domain.GtdEnabled {
			return fmt.Errorf("fetching domains from DNSME failed: domains with GTD enabled are not supported")
		}

		domains[domain.Name] = domain
	}

	api.domains = domains

	return nil
}

func (api *dnsmeProvider) domainExists(name string) (bool, error) {
	if err := api.loadDomains(); err != nil {
		return false, err
	}

	_, ok := api.domains[name]

	return ok, nil
}

func (api *dnsmeProvider) findDomain(name string) (*multiDomainResponseDataEntry, error) {
	if err := api.loadDomains(); err != nil {
		return nil, err
	}

	domain, ok := api.domains[name]
	if !ok {
		return nil, fmt.Errorf("domain not found on this DNSME account: %q", name)
	}

	return &domain, nil
}

func (api *dnsmeProvider) fetchDomainRecords(domainName string) ([]recordResponseDataEntry, error) {
	domain, err := api.findDomain(domainName)
	if err != nil {
		return nil, err
	}

	res, err := api.restAPI.recordGet(domain.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching records failed: %w", err)
	}

	records := make([]recordResponseDataEntry, 0)
	for _, record := range res.Data {
		if record.GtdLocation != "DEFAULT" {
			return nil, fmt.Errorf("fetching records from DNSME failed: only records with DEFAULT GTD location are supported")
		}

		records = append(records, record)
	}

	return records, nil
}

func (api *dnsmeProvider) fetchDomainNameServers(domainName string) ([]string, error) {
	domain, err := api.findDomain(domainName)
	if err != nil {
		return nil, err
	}

	res, err := api.restAPI.singleDomainGet(domain.ID)
	if err != nil {
		return nil, fmt.Errorf("fetching domain from DNSME failed: %w", err)
	}

	var nameServers []string
	for i := range res.NameServers {
		nameServers = append(nameServers, res.NameServers[i].Fqdn)
	}

	return nameServers, nil
}

func (api *dnsmeProvider) createDomain(domain string) error {
	_, err := api.restAPI.singleDomainCreate(singleDomainRequestData{Name: domain})

	if err != nil {
		return err
	}

	// reset cached domains after adding a new one, they will be refetched when needed
	api.domains = nil

	return nil
}

func (api *dnsmeProvider) deleteRecord(domainID int, record recordResponseDataEntry) error {
	err := api.restAPI.recordDelete(domainID, record.ID)

	return err
}

func (api *dnsmeProvider) deleteRecords(domainID int, recordIds []int) error {
	err := api.restAPI.multiRecordDelete(domainID, recordIds)

	return err
}

func (api *dnsmeProvider) updateRecords(domainID int, records []recordRequestData) error {
	err := api.restAPI.multiRecordUpdate(domainID, records)

	return err
}

func (api *dnsmeProvider) createRecords(domainID int, records []recordRequestData) error {
	_, err := api.restAPI.multiRecordCreate(domainID, records)

	return err
}
