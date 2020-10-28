package cloudns

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Api layer for CloDNS
type cloudnsProvider struct {
	domainIndex      map[string]string
	nameserversNames []string
	creds            struct {
		id       string
		password string
	}
}

type requestParams map[string]string

type errorResponse struct {
	Status      string `json:"status"`
	Description string `json:"statusDescription"`
}

type nameserverRecord struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type nameserverResponse []nameserverRecord

type zoneRecord struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Zone   string `json:"zone"`
}

type zoneResponse []zoneRecord

type domainRecord struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Host             string `json:"host"`
	Target           string `json:"record"`
	Priority         string `json:"priority"`
	Weight           string `json:"weight"`
	Port             string `json:"port"`
	Service          string `json:"service"`
	Protocol         string `json:"protocol"`
	TTL              string `json:"ttl"`
	Status           int8   `json:"status"`
	CaaFlag          string `json:"caa_flag,omitempty"`
	CaaTag           string `json:"caa_type,omitempty"`
	CaaValue         string `json:"caa_value,omitempty"`
	TlsaUsage        string `json:"tlsa_usage,omitempty"`
	TlsaSelector     string `json:"tlsa_selector,omitempty"`
	TlsaMatchingType string `json:"tlsa_matching_type,omitempty"`
	SshfpAlgorithm   string `json:"algorithm,omitempty"`
	SshfpFingerprint string `json:"fp_type,omitempty"`
}

type recordResponse map[string]domainRecord

var allowedTTLValues = []uint32{
	60,      // 1 minute
	300,     // 5 minutes
	900,     // 15 minutes
	1800,    // 30 minutes
	3600,    // 1 hour
	21600,   // 6 hours
	43200,   // 12 hours
	86400,   // 1 day
	172800,  // 2 days
	259200,  // 3 days
	604800,  // 1 week
	1209600, // 2 weeks
	2419200, // 4 weeks
}

func (c *cloudnsProvider) fetchAvailableNameservers() error {
	c.nameserversNames = nil

	var bodyString, err = c.get("/dns/available-name-servers.json", requestParams{})
	if err != nil {
		return fmt.Errorf("failed fetching available nameservers list from ClouDNS: %s", err)
	}

	var nr nameserverResponse
	json.Unmarshal(bodyString, &nr)

	for _, nameserver := range nr {
		if nameserver.Type == "premium" {
			c.nameserversNames = append(c.nameserversNames, nameserver.Name)
		}

	}
	return nil
}

func (c *cloudnsProvider) fetchDomainList() error {
	c.domainIndex = map[string]string{}
	rowsPerPage := 100
	page := 1
	for {
		var dr zoneResponse
		params := requestParams{
			"page":          strconv.Itoa(page),
			"rows-per-page": strconv.Itoa(rowsPerPage),
		}
		endpoint := "/dns/list-zones.json"
		var bodyString, err = c.get(endpoint, params)
		if err != nil {
			return fmt.Errorf("failed fetching domain list from ClouDNS: %s", err)
		}
		json.Unmarshal(bodyString, &dr)

		for _, domain := range dr {
			c.domainIndex[domain.Name] = domain.Name
		}
		if len(dr) < rowsPerPage {
			break
		}
		page++
	}
	return nil
}

func (c *cloudnsProvider) createDomain(domain string) error {
	params := requestParams{
		"domain-name": domain,
		"zone-type":   "master",
	}
	if _, err := c.get("/dns/register.json", params); err != nil {
		return fmt.Errorf("failed create domain (ClouDNS): %s", err)
	}
	return nil
}

func (c *cloudnsProvider) createRecord(domainID string, rec requestParams) error {
	rec["domain-name"] = domainID
	if _, err := c.get("/dns/add-record.json", rec); err != nil {
		return fmt.Errorf("failed create record (ClouDNS): %s", err)
	}
	return nil
}

func (c *cloudnsProvider) deleteRecord(domainID string, recordID string) error {
	params := requestParams{
		"domain-name": domainID,
		"record-id":   recordID,
	}
	if _, err := c.get("/dns/delete-record.json", params); err != nil {
		return fmt.Errorf("failed delete record (ClouDNS): %s", err)
	}
	return nil
}

func (c *cloudnsProvider) modifyRecord(domainID string, recordID string, rec requestParams) error {
	rec["domain-name"] = domainID
	rec["record-id"] = recordID
	if _, err := c.get("/dns/mod-record.json", rec); err != nil {
		return fmt.Errorf("failed update (ClouDNS): %s", err)
	}
	return nil
}

func (c *cloudnsProvider) getRecords(id string) ([]domainRecord, error) {
	params := requestParams{"domain-name": id}

	var bodyString, err = c.get("/dns/records.json", params)
	if err != nil {
		return nil, fmt.Errorf("failed fetching record list from ClouDNS: %s", err)
	}

	var dr recordResponse
	json.Unmarshal(bodyString, &dr)

	var records []domainRecord
	for _, rec := range dr {
		records = append(records, rec)
	}
	return records, nil
}

func (c *cloudnsProvider) get(endpoint string, params requestParams) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.cloudns.net"+endpoint, nil)
	q := req.URL.Query()

	//TODO: Support  sub-auth-id / sub-auth-user https://asia.cloudns.net/wiki/article/42/
	// Add auth params
	q.Add("auth-id", c.creds.id)
	q.Add("auth-password", c.creds.password)

	for pName, pValue := range params {
		q.Add(pName, pValue)
	}

	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)

	// Got error from API ?
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err == nil {
		if errResp.Status == "Failed" {
			return bodyString, fmt.Errorf("ClouDNS API error: %s URL:%s%s ", errResp.Description, req.Host, req.URL.RequestURI())
		}
	}

	return bodyString, nil
}

func fixTTL(ttl uint32) uint32 {
	// if the TTL is larger than the largest allowed value, return the largest allowed value
	if ttl > allowedTTLValues[len(allowedTTLValues)-1] {
		return allowedTTLValues[len(allowedTTLValues)-1]
	}

	for _, v := range allowedTTLValues {
		if v >= ttl {
			return v
		}
	}

	return allowedTTLValues[0]
}
