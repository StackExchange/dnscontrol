package infomaniak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const baseURL = "https://api.infomaniak.com/2"

type dnssecRecord struct {
	IsEnabled bool `json:"is_enabled"`
}

type errorRecord struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type dnsZoneResponse struct {
	Result string      `json:"result"`
	Data   dnsZone     `json:"data,omitempty"`
	Error  errorRecord `json:"error,omitempty"`
}

type dnsRecordsResponse struct {
	Result string      `json:"result"`
	Data   []dnsRecord `json:"data,omitempty"`
	Error  errorRecord `json:"error,omitempty"`
}

type dnsRecordResponse struct {
	Result string      `json:"result"`
	Data   dnsRecord   `json:"data,omitempty"`
	Error  errorRecord `json:"error,omitempty"`
}

type boolResponse struct {
	Result string      `json:"result"`
	Data   bool        `json:"data,omitempty"`
	Error  errorRecord `json:"error,omitempty"`
}
type dnsZone struct {
	ID          int64        `json:"id,omitempty"`
	FQDN        string       `json:"fqdn,omitempty"`
	DNSSEC      dnssecRecord `json:"dnssec,omitempty"`
	Nameservers []string     `json:"nameservers,omitempty"`
}

type dnsRecord struct {
	ID        int64  `json:"id,omitempty"`
	Source    string `json:"source,omitempty"`
	Type      string `json:"type,omitempty"`
	TTL       int64  `json:"ttl,omitempty"`
	Target    string `json:"target,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

type dnsRecordCreate struct {
	Source string `json:"source,omitempty"`
	Type   string `json:"type,omitempty"`
	TTL    int64  `json:"ttl,omitempty"`
	Target string `json:"target,omitempty"`
}

// Get zone information
// See https://developer.infomaniak.com/docs/api/get/2/zones/%7Bzone%7D
func (p *infomaniakProvider) getDNSZone(zone string) (*dnsZone, error) {
	reqURL := fmt.Sprintf("%s/zones/%s", baseURL, zone)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+p.apiToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := &dnsZoneResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

// Retrieve all dns record for a given zone
// See https://developer.infomaniak.com/docs/api/get/2/zones/%7Bzone%7D/records
func (p *infomaniakProvider) getDNSRecords(zone string) ([]dnsRecord, error) {
	reqURL := fmt.Sprintf("%s/zones/%s/records", baseURL, zone)

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+p.apiToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := &dnsRecordsResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// Delete a dns record
// See https://developer.infomaniak.com/docs/api/delete/2/zones/%7Bzone%7D/records/%7Brecord%7D
func (p *infomaniakProvider) deleteDNSRecord(zone string, recordID string) error {
	reqURL := fmt.Sprintf("%s/zones/%s/records/%s", baseURL, zone, recordID)

	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+p.apiToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	response := &boolResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return err
	}

	if response.Result == "error" {
		return fmt.Errorf("failed to delete record %s in zone %s: %s", recordID, zone, response.Error.Description)
	}

	return nil
}

// Create a dns record in a given zone
// See https://developer.infomaniak.com/docs/api/post/2/zones/%7Bzone%7D/records
func (p *infomaniakProvider) createDNSRecord(zone string, rec *dnsRecordCreate) (*dnsRecord, error) {
	reqURL := fmt.Sprintf("%s/zones/%s/records", baseURL, zone)

	data, err := json.Marshal(rec)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+p.apiToken)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	response := &dnsRecordResponse{}
	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	if response.Result == "error" {
		return nil, fmt.Errorf("failed to create %s record in zone %s: %s", rec.Type, zone, response.Error.Description)
	}

	return &response.Data, nil
}
