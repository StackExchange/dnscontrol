package cscglobal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
)

const apiBase = "https://apis.cscglobal.com/dbs/api/v2"

// Api layer for CSC Global

type cscglobalProvider struct {
	key          string
	token        string
	notifyEmails []string
}

type requestParams map[string]string

type errorResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Value       string `json:"value,omitempty"`
}

type nsModRequest struct {
	Domain        string   `json:"qualifiedDomainName"`
	NameServers   []string `json:"nameServers"`
	DNSType       string   `json:"dnsType,omitempty"`
	Notifications struct {
		Enabled bool     `json:"enabled,omitempty"`
		Emails  []string `json:"additionalNotificationEmails,omitempty"`
	} `json:"notifications"`
	ShowPrice    bool     `json:"showPrice,omitempty"`
	CustomFields []string `json:"customFields,omitempty"`
}

type nsModRequestResult struct {
	Result struct {
		Domain string `json:"qualifiedDomainName"`
		Status struct {
			Code                  string `json:"code"`
			Message               string `json:"message"`
			AdditionalInformation string `json:"additionalInformation"`
			UUID                  string `json:"uuid"`
		} `json:"status"`
	} `json:"result"`
}

type domainRecord struct {
	Nameserver []string `json:"nameservers"`
}

func (c *cscglobalProvider) getNameservers(domain string) ([]string, error) {
	var bodyString, err = c.get("/domains/" + domain)
	if err != nil {
		return nil, err
	}

	var dr domainRecord
	json.Unmarshal(bodyString, &dr)
	ns := []string{}
	for _, nameserver := range dr.Nameserver {
		ns = append(ns, nameserver)
	}
	sort.Strings(ns)
	return ns, nil
}

func (c *cscglobalProvider) updateNameservers(ns []string, domain string) error {
	req := nsModRequest{
		Domain:      domain,
		NameServers: ns,
		DNSType:     "OTHER_DNS",
		ShowPrice:   false,
	}
	if c.notifyEmails != nil {
		req.Notifications.Enabled = true
		req.Notifications.Emails = c.notifyEmails
	}
	req.CustomFields = []string{}

	requestBody, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	bodyString, err := c.put("/domains/nsmodification", requestBody)
	if err != nil {
		return fmt.Errorf("CSC Global: Error update NS : %w", err)
	}

	var res nsModRequestResult
	json.Unmarshal(bodyString, &res)
	if res.Result.Status.Code != "SUBMITTED" {
		return fmt.Errorf("CSC Global: Error update NS Code: %s Message: %s AdditionalInfo: %s", res.Result.Status.Code, res.Result.Status.Message, res.Result.Status.AdditionalInformation)
	}

	return nil
}

func (c *cscglobalProvider) put(endpoint string, requestBody []byte) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("PUT", apiBase+endpoint, bytes.NewReader(requestBody))

	// Add headers
	req.Header.Add("apikey", c.key)
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return bodyString, nil
	}

	// Got a error response from API, see if it's json format
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		// Some error messages are plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}

func (c *cscglobalProvider) get(endpoint string) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiBase+endpoint, nil)

	// Add headers
	req.Header.Add("apikey", c.key)
	req.Header.Add("Authorization", "Bearer "+c.token)
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return bodyString, nil
	}

	if resp.StatusCode == 400 {
		// 400, error message is in the body as plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}

	// Got a json error response from API
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}
