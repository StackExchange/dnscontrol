package infomaniak

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const baseURL = "https://api.infomaniak.com/2"

type dnssecRecord struct {
	IsEnabled bool `json:"is_enabled"`
}

type dnsZoneResponse struct {
	Result string  `json:"result"`
	Data   dnsZone `json:"data"`
}

type dnsRecordResponse struct {
	Result string      `json:"result"`
	Data   []dnsRecord `json:"data"`
}

type boolResponse struct {
	Result string `json:"result"`
	Data   bool   `json:"data"`
}
type dnsZone struct {
	ID          int64        `json:"id,omitempty"`
	FQDN        string       `json:"fqdn,omitempty"`
	DNSSEC      dnssecRecord `json:"dnssec,omitempty"`
	Nameservers []string     `json:"nameservers,omitempty"`
}

type dnsRecord struct {
	ID        string `json:"hostname,omitempty"`
	Source    string `json:"source,omitempty"`
	Type      string `json:"type,omitempty"`
	TTL       int64  `json:"ttl,omitempty"`
	Target    string `json:"target,omitempty"`
	UpdatedAt int64  `json:"updated_at,omitempty"`
}

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
	response := &dnsZoneResponse{}

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return &response.Data, nil
}

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
	response := &dnsRecordResponse{}

	err = json.NewDecoder(res.Body).Decode(response)
	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

func (p *infomaniakProvider) deleteDNSRecord(zone string, record string) error {
	reqURL := fmt.Sprintf("%s/zones/%s/records/%s", baseURL, zone, record)

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

	return nil
}

// func (p *infomaniakProvider) createDNSRecord(zoneID string, rec *dnsRecordCreate) (*dnsRecord, error) {
// 	reqURL := fmt.Sprintf("%s/dns_zones/%s/dns_records", baseURL, zoneID)

// 	data, err := json.Marshal(rec)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(data))
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.Header.Add("Authorization", "Bearer "+n.apiToken)
// 	req.Header.Add("Content-Type", "application/json")

// 	res, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()

// 	record := &dnsRecord{}

// 	err = json.NewDecoder(res.Body).Decode(record)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return record, nil
// }
