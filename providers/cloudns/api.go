package cloudns

import (
	"encoding/json"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

// LinodeApi is the handle for this provider.

type Credentials struct {
	id       string
	password string
}

type CloudnsApi struct {
	domainIndex map[string]string
	creds       Credentials
}

func (c *CloudnsApi) fetchDomainList() error {
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
			return errors.Errorf("Error fetching domain list from ClouDNS: %s", err)
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

func (c *CloudnsApi) createDomain(domain string) error {
	params := requestParams{
		"domain-name": domain,
		"zone-type":   "master",
	}
	if _, err := c.get("/dns/register.json", params); err != nil {
		return errors.Errorf("Error create domain  ClouDNS: %s", err)
	}
	return nil
}

func (c *CloudnsApi) createRecord(domainID string, rec requestParams) error {
	rec["domain-name"] = domainID
	if _, err := c.get("/dns/add-record.json", rec); err != nil {
		return errors.Errorf("Error create record  ClouDNS: %s", err)
	}
	return nil
}

func (c *CloudnsApi) deleteRecord(domainID string, recordID string) error {
	params := requestParams{
		"domain-name": domainID,
		"record-id":   recordID,
	}
	if _, err := c.get("/dns/delete-record.json", params); err != nil {
		return errors.Errorf("Error delete record  ClouDNS: %s", err)
	}
	return nil
}

func (c *CloudnsApi) modifyRecord(domainID string, recordID string, rec requestParams) error {
	rec["domain-name"] = domainID
	rec["record-id"] = recordID
	if _, err := c.get("/dns/mod-record.json", rec); err != nil {
		return errors.Errorf("Error create update ClouDNS: %s", err)
	}
	return nil
}

func (c *CloudnsApi) getRecords(id string) ([]domainRecord, error) {
	params := requestParams{"domain-name": id}

	var bodyString, err = c.get("/dns/records.json", params)
	if err != nil {
		return nil, errors.Errorf("Error fetching record list from ClouDNS: %s", err)
	}

	var dr recordResponse
	json.Unmarshal(bodyString, &dr)

	var records []domainRecord
	for _, rec := range dr {
		records = append(records, rec)
	}
	return records, nil
}

func (c *CloudnsApi) get(endpoint string, params requestParams) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.cloudns.net"+endpoint, nil)
	q := req.URL.Query()

	//TODO: Suport  sub-auth-id / sub-auth-user https://asia.cloudns.net/wiki/article/42/
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
	if errResp.Status == "Failed" {
		return bodyString, errors.Errorf("ClouDNS API error: %s URL:%s%s ", errResp.Description, req.Host, req.URL.RequestURI())
	}

	return bodyString, nil
}

type requestParams map[string]string

type errorResponse struct {
	Status      string `json:"status"`
	Description string `json:"statusDescription"`
}

type zoneRecord struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Zone   string `json:"zone"`
}

type zoneResponse []zoneRecord

type domainRecord struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Host     string `json:"host"`
	Target   string `json:"record"`
	Priority string `json:"priority"`
	Weight   string `json:"weight"`
	Port     string `json:"port"`
	Service  string `json:"service"`
	Protocol string `json:"protocol"`
	TTL      string `json:"ttl"`
	Status   int8   `json:"status"`
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
