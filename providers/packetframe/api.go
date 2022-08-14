package packetframe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	mediaType      = "application/json"
	defaultBaseURL = "https://packetframe.com/api/"
)

var defaultNameServerNames = []string{
	"ns1.packetframe.com",
	"ns2.packetframe.com",
}

type zone struct {
	ID         string   `json:"id"`
	Zone       string   `json:"zone"`
	Users      []string `json:"users"`
	UserEmails []string `json:"user_emails"`
}

type domainResponse struct {
	Data struct {
		Zones []zone `json:"zones"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type deleteRequest struct {
	Record string `json:"record"`
	Zone   string `json:"zone"`
}

type recordResponse struct {
	Data struct {
		Records []domainRecord `json:"records"`
	} `json:"data"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type domainRecord struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
	Proxy bool   `json:"proxy"`
	Zone  string `json:"zone"`
}

func (api *packetframeProvider) fetchDomainList() error {
	api.domainIndex = map[string]zone{}
	dr := &domainResponse{}
	endpoint := "dns/zones"
	if err := api.get(endpoint, dr); err != nil {
		return fmt.Errorf("failed fetching domain list (Packetframe): %w", err)
	}
	for _, zone := range dr.Data.Zones {
		api.domainIndex[zone.Zone] = zone
	}

	return nil
}

func (api *packetframeProvider) getRecords(zoneID string) ([]domainRecord, error) {
	var records []domainRecord
	dr := &recordResponse{}
	endpoint := "dns/records/" + zoneID
	if err := api.get(endpoint, dr); err != nil {
		return records, fmt.Errorf("failed fetching domain list (Packetframe): %w", err)
	}
	records = append(records, dr.Data.Records...)

	for i := range defaultNameServerNames {
		records = append(records, domainRecord{
			Type:  "NS",
			TTL:   86400,
			Value: defaultNameServerNames[i] + ".",
			Zone:  zoneID,
			ID:    "0",
		})
	}

	return records, nil
}

func (api *packetframeProvider) createRecord(rec *domainRecord) (*domainRecord, error) {
	endpoint := "dns/records"

	req, err := api.newRequest(http.MethodPost, endpoint, rec)
	if err != nil {
		return nil, err
	}

	_, err = api.client.Do(req)
	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (api *packetframeProvider) modifyRecord(rec *domainRecord) error {
	endpoint := "dns/records"

	req, err := api.newRequest(http.MethodPut, endpoint, rec)
	if err != nil {
		return err
	}

	_, err = api.client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (api *packetframeProvider) deleteRecord(zoneID string, recordID string) error {
	endpoint := "dns/records"
	req, err := api.newRequest(http.MethodDelete, endpoint, deleteRequest{Zone: zoneID, Record: recordID})
	if err != nil {
		return err
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return api.handleErrors(resp)
	}

	return nil
}

func (api *packetframeProvider) newRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	u := api.baseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("Authorization", "Token "+api.token)
	return req, nil
}

func (api *packetframeProvider) get(endpoint string, target interface{}) error {
	req, err := api.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return api.handleErrors(resp)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(target)
}

func (api *packetframeProvider) handleErrors(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dr := &domainResponse{}
	json.Unmarshal(body, &dr)

	return fmt.Errorf("packetframe API error: %s", dr.Message)
}
