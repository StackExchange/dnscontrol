package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	baseURL = "https://porkbun.com/api/json/v3"
)

type porkbunProvider struct {
	apiKey    string
	secretKey string
}

type requestParams map[string]any

type errorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type domainRecord struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Content string `json:"content"`
	TTL     string `json:"ttl"`
	Prio    string `json:"prio"`
}

type recordResponse struct {
	Records []domainRecord `json:"records"`
}

type domainListRecord struct {
	Domain string `json:"domain"`
}

type domainListResponse struct {
	Domains []domainListRecord `json:"domains"`
}

type nsResponse struct {
	Nameservers []string `json:"ns"`
}

func (c *porkbunProvider) post(endpoint string, params requestParams) ([]byte, error) {
	params["apikey"] = c.apiKey
	params["secretapikey"] = c.secretKey

	personJSON, err := json.Marshal(params)
	if err != nil {
		return []byte{}, err
	}

	client := &http.Client{}
	req, _ := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(personJSON))

	retrycnt := 0

	// If request sending too fast, the server will fail with the following error:
	// porkbun API error: Create error: We were unable to create the DNS record.
retry:
	time.Sleep(500 * time.Millisecond)
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 202 {
		retrycnt += 1
		if retrycnt == 5 {
			return bodyString, fmt.Errorf("rate limiting exceeded")
		}
		printer.Warnf("Rate limiting.. waiting for %d minute(s)\n", retrycnt)
		time.Sleep(time.Minute * time.Duration(retrycnt))
		goto retry
	}

	// Got error from API ?
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err == nil {
		if errResp.Status == "ERROR" {
			return bodyString, fmt.Errorf("porkbun API error: %s URL:%s%s ", errResp.Message, req.Host, req.URL.RequestURI())
		}
	}

	return bodyString, nil
}

func (c *porkbunProvider) ping() error {
	params := requestParams{}
	_, err := c.post("/ping", params)
	return err
}

func (c *porkbunProvider) createRecord(domain string, rec requestParams) error {
	if _, err := c.post("/dns/create/"+domain, rec); err != nil {
		return fmt.Errorf("failed create record (porkbun): %w", err)
	}
	return nil
}

func (c *porkbunProvider) deleteRecord(domain string, recordID string) error {
	params := requestParams{}
	if _, err := c.post(fmt.Sprintf("/dns/delete/%s/%s", domain, recordID), params); err != nil {
		return fmt.Errorf("failed delete record (porkbun): %w", err)
	}
	return nil
}

func (c *porkbunProvider) modifyRecord(domain string, recordID string, rec requestParams) error {
	if _, err := c.post(fmt.Sprintf("/dns/edit/%s/%s", domain, recordID), rec); err != nil {
		return fmt.Errorf("failed update (porkbun): %w", err)
	}
	return nil
}

func (c *porkbunProvider) getRecords(domain string) ([]domainRecord, error) {
	params := requestParams{}
	var bodyString, err = c.post("/dns/retrieve/"+domain, params)
	if err != nil {
		return nil, fmt.Errorf("failed fetching record list from porkbun: %w", err)
	}

	var dr recordResponse
	err = json.Unmarshal(bodyString, &dr)
	if err != nil {
		return nil, fmt.Errorf("failed parsing record list from porkbun: %w", err)
	}

	var records []domainRecord
	for _, rec := range dr.Records {
		if rec.Name == domain && rec.Type == "NS" {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func (c *porkbunProvider) getNameservers(domain string) ([]string, error) {
	params := requestParams{}
	var bodyString, err = c.post(fmt.Sprintf("/domain/getNs/%s", domain), params)
	if err != nil {
		return nil, fmt.Errorf("failed fetching nameserver list from porkbun: %w", err)
	}

	var ns nsResponse
	err = json.Unmarshal(bodyString, &ns)
	if err != nil {
		return nil, fmt.Errorf("failed parsing nameserver list from porkbun: %w", err)
	}

	sort.Strings(ns.Nameservers)

	var nameservers []string
	for _, nameserver := range ns.Nameservers {
		// Remove the trailing dot only if it exists.
		// This provider seems to add the trailing dot to some domains but not others.
		// The .DE domains seem to always include the dot, others don't.
		nameservers = append(nameservers, strings.TrimSuffix(nameserver, "."))
	}
	return nameservers, nil
}

func (c *porkbunProvider) updateNameservers(ns []string, domain string) error {
	params := requestParams{}
	params["ns"] = ns
	if _, err := c.post(fmt.Sprintf("/domain/updateNs/%s", domain), params); err != nil {
		return fmt.Errorf("failed NS update (porkbun): %w", err)
	}
	return nil
}

func (c *porkbunProvider) listAllDomains() ([]string, error) {
	params := requestParams{}
	var bodyString, err = c.post("/domain/listAll", params)
	if err != nil {
		return nil, fmt.Errorf("failed listing all domains from porkbun: %w", err)
	}

	var dlr domainListResponse
	err = json.Unmarshal(bodyString, &dlr)
	if err != nil {
		return nil, fmt.Errorf("failed parsing domain list from porkbun: %w", err)
	}

	var domains []string
	for _, domain := range dlr.Domains {
		domains = append(domains, domain.Domain)
	}
	sort.Strings(domains)
	return domains, nil
}
