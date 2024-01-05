package realtimeregister

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type realtimeregisterAPI struct {
	apikey      string
	endpoint    string
	Zones       map[string]*Zone //cache
	ServiceType string
}

type Zones struct {
	Entities []Zone `json:"entities"`
}

type Domain struct {
	Nameservers []string `json:"ns"`
}

type Zone struct {
	Name    string   `json:"name,omitempty"`
	Service string   `json:"service,omitempty"`
	ID      int      `json:"id,omitempty"`
	Records []Record `json:"records"`
	Dnssec  bool     `json:"dnssec"`
}

type Record struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Content  string `json:"content"`
	Priority int    `json:"prio,omitempty"`
	TTL      int    `json:"ttl"`
}

const (
	endpoint        = "https://api.yoursrs.com/v2"
	endpointSandbox = "https://api.yoursrs-ote.com/v2"
)

func (api *realtimeregisterAPI) request(method string, url string, body io.Reader) ([]byte, error) {
	client := &http.Client{}
	req, _ := http.NewRequest(
		method,
		url,
		body,
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "ApiKey "+api.apikey)

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	bodyString, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("realtime Register API error on request to %s: %d, %s", url, resp.StatusCode,
			string(bodyString))
	}

	return bodyString, nil
}

func (api *realtimeregisterAPI) getZone(domain string) (*Zone, error) {
	zones, err := api.getDomainZones(domain)
	if err != nil {
		return nil, err
	}

	if len(zones.Entities) == 0 {
		return nil, fmt.Errorf("zone %s does not exist", domain)
	}

	api.Zones[domain] = &zones.Entities[0]

	return &zones.Entities[0], nil
}

func (api *realtimeregisterAPI) getDomainZones(domain string) (*Zones, error) {

	url := fmt.Sprintf(api.endpoint+"/dns/zones?name=%s&service=%s", domain, api.ServiceType)

	return api.getZones(url)
}

func (api *realtimeregisterAPI) getAllZones() ([]string, error) {
	url := fmt.Sprintf(api.endpoint+"/dns/zones?service=%s&export=true&fields=id,name", api.ServiceType)

	zones, err := api.getZones(url)
	if err != nil {
		return nil, err
	}

	zoneNames := make([]string, len(zones.Entities))

	for i, zone := range zones.Entities {
		zoneNames[i] = zone.Name
	}

	return zoneNames, nil
}

func (api *realtimeregisterAPI) getZones(url string) (*Zones, error) {
	bodyBytes, err := api.request(
		"GET",
		url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	respData := &Zones{}
	err = json.Unmarshal(bodyBytes, &respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

func (api *realtimeregisterAPI) createZone(domain string) error {
	zone := &Zone{
		Records: []Record{},
		Name:    domain,
		Service: api.ServiceType,
	}

	err := api.createOrUpdateZone(zone, api.endpoint+"/dns/zones")
	if err != nil {
		return err
	}

	return nil
}

func (api *realtimeregisterAPI) zoneExists(domain string) (bool, error) {
	if api.Zones[domain] != nil {
		return true, nil
	}
	zones, err := api.getDomainZones(domain)
	if err != nil {
		return false, err
	}
	return len(zones.Entities) > 0, nil
}

func (api *realtimeregisterAPI) getDomainNameservers(domainName string) ([]string, error) {
	respData, err := api.request(
		"GET",
		fmt.Sprintf(api.endpoint+"/domains/%s", domainName),
		nil,
	)
	if err != nil {
		return nil, err
	}
	domain := &Domain{}
	err = json.Unmarshal(respData, &domain)
	if err != nil {
		return nil, err
	}
	return domain.Nameservers, nil
}

func (api *realtimeregisterAPI) updateZone(domain string, body *Zone) error {
	return api.createOrUpdateZone(
		body,
		fmt.Sprintf(api.endpoint+"/dns/zones/%d/update", api.Zones[domain].ID),
	)
}

func (api *realtimeregisterAPI) updateNameservers(domainName string, nameservers []string) error {
	domain := &Domain{
		Nameservers: nameservers,
	}

	bodyBytes, err := json.Marshal(domain)
	if err != nil {
		return err
	}

	_, err = api.request(
		"POST",
		fmt.Sprintf(api.endpoint+"/domains/%s/update", domainName),
		bytes.NewReader(bodyBytes),
	)

	if err != nil {
		return err
	}
	return nil
}

func (api *realtimeregisterAPI) createOrUpdateZone(body *Zone, url string) error {
	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return err
	}

	//Ugly hack for MX records with null target
	requestBody := strings.Replace(string(bodyBytes), "\"prio\":-1", "\"prio\":0", -1)

	_, err = api.request("POST", url, strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	return nil
}
