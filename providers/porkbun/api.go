package porkbun

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL = "https://porkbun.com/api/json/v3"
)

type porkbunProvider struct {
	apiKey    string
	secretKey string
}

type requestParams map[string]string

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
	Notes   string `json:"notes"`
}

type recordResponse struct {
	Status  string         `json:"status"`
	Records []domainRecord `json:"records"`
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

	// If request sending too fast, the server will fail with the following error:
	// porkbun API error: Create error: We were unable to create the DNS record.
	time.Sleep(500 * time.Millisecond)
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := io.ReadAll(resp.Body)

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
		return fmt.Errorf("failed create record (porkbun): %s", err)
	}
	return nil
}

func (c *porkbunProvider) deleteRecord(domain string, recordID string) error {
	params := requestParams{}
	if _, err := c.post(fmt.Sprintf("/dns/delete/%s/%s", domain, recordID), params); err != nil {
		return fmt.Errorf("failed delete record (porkbun): %s", err)
	}
	return nil
}

func (c *porkbunProvider) modifyRecord(domain string, recordID string, rec requestParams) error {
	if _, err := c.post(fmt.Sprintf("/dns/edit/%s/%s", domain, recordID), rec); err != nil {
		return fmt.Errorf("failed update (porkbun): %s", err)
	}
	return nil
}

func (c *porkbunProvider) getRecords(domain string) ([]domainRecord, error) {
	params := requestParams{}
	var bodyString, err = c.post("/dns/retrieve/"+domain, params)
	if err != nil {
		return nil, fmt.Errorf("failed fetching record list from porkbun: %s", err)
	}

	var dr recordResponse
	json.Unmarshal(bodyString, &dr)

	var records []domainRecord
	for _, rec := range dr.Records {
		if rec.Name == domain && rec.Type == "NS" {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}
