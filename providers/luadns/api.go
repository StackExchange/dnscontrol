package luadns

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// Api layer for LuaDNS

const (
	apiURL = "https://api.luadns.com/v1"
)

type luadnsProvider struct {
	domainIndex      map[string]uint32
	nameserversNames []string
	creds            struct {
		email  string
		apikey string
	}
}

type errorResponse struct {
	Status    string `json:"status"`
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
}

type userInfoResponse struct {
	Email       string   `json:"email"`
	Name        string   `json:"name"`
	TTL         uint32   `json:"ttl"`
	NameServers []string `json:"name_servers"`
}

type zoneRecord struct {
	ID   uint32 `json:"id"`
	Name string `json:"name"`
}

type zoneResponse []zoneRecord

type domainRecord struct {
	ID      uint32 `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     uint32 `json:"ttl"`
}

type recordResponse []domainRecord

type requestParams map[string]string
type jsonRequestParams map[string]any

func (l *luadnsProvider) fetchAvailableNameservers() error {
	l.nameserversNames = nil
	var bodyString, err = l.get("/users/me", "GET", requestParams{})
	if err != nil {
		return fmt.Errorf("failed fetching available nameservers list from LuaDNS: %s", err)
	}
	var ui userInfoResponse
	json.Unmarshal(bodyString, &ui)
	l.nameserversNames = ui.NameServers
	return nil
}

func (l *luadnsProvider) fetchDomainList() error {
	l.domainIndex = map[string]uint32{}
	var bodyString, err = l.get("/zones", "GET", requestParams{})
	if err != nil {
		return fmt.Errorf("failed fetching domain list from LuaDNS: %s", err)
	}
	var dr zoneResponse
	json.Unmarshal(bodyString, &dr)
	for _, domain := range dr {
		l.domainIndex[domain.Name] = domain.ID
	}
	return nil
}

func (l *luadnsProvider) getDomainID(name string) (uint32, error) {
	if l.domainIndex == nil {
		if err := l.fetchDomainList(); err != nil {
			return 0, err
		}
	}
	id, ok := l.domainIndex[name]
	if !ok {
		return 0, fmt.Errorf("'%s' not a zone in luadns account", name)
	}
	return id, nil
}

func (l *luadnsProvider) createDomain(domain string) error {
	params := jsonRequestParams{
		"name": domain,
	}
	if _, err := l.get("/zones", "POST", params); err != nil {
		return fmt.Errorf("failed create domain (LuaDNS): %s", err)
	}
	return nil
}

func (l *luadnsProvider) createRecord(domainID uint32, rec jsonRequestParams) error {
	if _, err := l.get(fmt.Sprintf("/zones/%d/records", domainID), "POST", rec); err != nil {
		return fmt.Errorf("failed create record (LuaDNS): %s", err)
	}
	return nil
}

func (l *luadnsProvider) deleteRecord(domainID uint32, recordID uint32) error {
	if _, err := l.get(fmt.Sprintf("/zones/%d/records/%d", domainID, recordID), "DELETE", requestParams{}); err != nil {
		return fmt.Errorf("failed delete record (LuaDNS): %s", err)
	}
	return nil
}

func (l *luadnsProvider) modifyRecord(domainID uint32, recordID uint32, rec jsonRequestParams) error {
	if _, err := l.get(fmt.Sprintf("/zones/%d/records/%d", domainID, recordID), "PUT", rec); err != nil {
		return fmt.Errorf("failed update (LuaDNS): %s", err)
	}
	return nil
}

func (l *luadnsProvider) getRecords(domainID uint32) ([]domainRecord, error) {
	var bodyString, err = l.get(fmt.Sprintf("/zones/%d/records", domainID), "GET", requestParams{})
	if err != nil {
		return nil, fmt.Errorf("failed fetching record list from LuaDNS: %s", err)
	}
	var dr recordResponse
	json.Unmarshal(bodyString, &dr)
	var records []domainRecord
	for _, rec := range dr {
		if rec.Type == "SOA" {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

func (l *luadnsProvider) get(endpoint string, method string, params any) ([]byte, error) {
	client := &http.Client{}
	var req, err = l.makeRequest(endpoint, method, params)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(l.creds.email, l.creds.apikey)
	// LuaDNS has a rate limit of 1200 request per 5 minute.
	// So we do a very primitive rate limiting here - delay every request for 250ms - so max. 4 requests/second.
	time.Sleep(250 * time.Millisecond)
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}

	bodyString, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return bodyString, nil
	}

	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err == nil {
		return bodyString, fmt.Errorf("LuaDNS API error: %s URL:%s%s", errResp.Message, req.Host, req.URL.RequestURI())
	} else {
		return bodyString, fmt.Errorf("LuaDNS API Error: %s URL:%s%s", string(bodyString), req.Host, req.URL.RequestURI())
	}
}

func (l *luadnsProvider) makeRequest(endpoint string, method string, params any) (*http.Request, error) {
	switch v := params.(type) {
	case requestParams:
		req, _ := http.NewRequest(method, apiURL+endpoint, nil)
		q := req.URL.Query()
		for pName, pValue := range v {
			q.Add(pName, pValue)
		}
		req.URL.RawQuery = q.Encode()
		return req, nil
	case jsonRequestParams:
		requestJSON, err := json.Marshal(params)
		if err != nil {
			return nil, err
		}
		req, _ := http.NewRequest(method, apiURL+endpoint, bytes.NewBuffer(requestJSON))
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	default:
		return nil, fmt.Errorf("Invalid request type")
	}
}

func nativeToRecord(domain string, r *domainRecord) *models.RecordConfig {
	rc := &models.RecordConfig{
		Type:     r.Type,
		TTL:      r.TTL,
		Original: r,
	}
	rc.SetLabelFromFQDN(r.Name, domain)
	switch rtype := rc.Type; rtype {
	case "TXT":
		rc.SetTargetTXT(r.Content)
	default:
		rc.PopulateFromString(rtype, r.Content, domain)
	}
	return rc
}

func recordsToNative(rc *models.RecordConfig) jsonRequestParams {
	r := jsonRequestParams{
		"name": fmt.Sprintf("%s.", rc.GetLabelFQDN()),
		"type": rc.Type,
		"ttl":  rc.TTL,
	}
	switch rtype := rc.Type; rtype {
	case "TXT":
		r["content"] = rc.GetTargetTXTJoined()
	default:
		r["content"] = rc.GetTargetCombined()
	}
	return r
}

func checkNS(dc *models.DomainConfig) {
	newList := make([]*models.RecordConfig, 0, len(dc.Records))
	for _, rec := range dc.Records {
		// LuaDNS does not support changing the TTL of the default nameservers, so forcefully change the TTL to 86400.
		if rec.Type == "NS" && strings.HasSuffix(rec.GetTargetField(), ".luadns.net.") && rec.TTL != 86400 {
			rec.TTL = 86400
		}
		newList = append(newList, rec)
	}
	dc.Records = newList
}
