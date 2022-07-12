package domainnameshop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var rootAPIURI = "https://api.domeneshop.no/v0"

//TODO: CHECK that domains match
func (api *domainNameShopProvider) getDomains(domainName string) ([]DomainResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodGet, rootAPIURI+"/domains?domain="+domainName, nil)
	if err != nil {
		// handle error
		return nil, err
	}
	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)

	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()
	domainResponse := make([]DomainResponse, 1)
	err = json.NewDecoder(resp.Body).Decode(&domainResponse)
	if err != nil {
		return nil, err
	}
	return domainResponse, nil
}

func (api *domainNameShopProvider) getDomainID(domainName string) (string, error) {
	domainResp, err := api.getDomains(domainName)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(domainResp[0].ID), nil
}

func (api *domainNameShopProvider) getNS(domainName string) ([]string, error) {
	domainResp, err := api.getDomains(domainName)
	if err != nil {
		return nil, err
	}
	return domainResp[0].Nameservers, nil
}

func (api *domainNameShopProvider) getDNS(domainName string) ([]DomainNameShopRecord, error) {
	domainID, err := api.getDomainID(domainName)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, rootAPIURI+"/domains/"+domainID+"/dns", nil)
	if err != nil {
		// handle error
		return nil, err
	}
	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)

	if err != nil {
		// handle error
		return nil, err
	}
	defer resp.Body.Close()
	domainResponse := make([]DomainNameShopRecord, 1)
	err = json.NewDecoder(resp.Body).Decode(&domainResponse)
	if err != nil {
		return nil, err
	}

	for i := range domainResponse {
		// Fix priority
		record := &domainResponse[i]
		priority, err := strconv.ParseUint(record.Priority, 10, 16)
		if err != nil {
			record.ActualPriority = 0
		}
		record.ActualPriority = uint16(priority)

		// Fix CAA flags
		if record.Type == "CAA" {
			CaaFlag, err := strconv.ParseUint(record.ActualCAAFlag, 10, 8)
			if err != nil {
				record.CAAFlag = 0
			}
			record.CAAFlag = CaaFlag
		}

		// Add domain id
		(&domainResponse[i]).DomainID = domainID
	}

	ns, err := api.getNS(domainName)
	if err != nil {
		return nil, err
	}

	// Adds NS as records
	for _, nameserver := range ns {
		domainResponse = append(domainResponse, DomainNameShopRecord{
			ID:       0,
			Host:     "@",
			TTL:      300,
			Type:     "NS",
			Data:     nameserver,
			DomainID: domainID,
		})
	}

	return domainResponse, nil
}

func (api *domainNameShopProvider) deleteRecord(domainID string, recordID string) error {
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, rootAPIURI+"/domains/"+domainID+"/dns/"+recordID, nil)
	if err != nil {
		// handle error
		return err
	}
	req.SetBasicAuth(api.Token, api.Secret)
	resp, err := client.Do(req)
	if err != nil {
		// handle error
		return err
	}

	switch resp.StatusCode {
	case 204:
		//Record is deleted
		return nil
	case 403:
		return fmt.Errorf("not authorized")
	case 404:
		return fmt.Errorf("DNS record does not exist")
	default:
		return fmt.Errorf("unknown statuscode: %v", resp.StatusCode)
	}
}
