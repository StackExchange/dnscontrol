package autodns

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/StackExchange/dnscontrol/v3/models"
)

// ZoneListFilter describes a JSON list filter.
type ZoneListFilter struct {
	Key      string            `json:"key"`
	Value    string            `json:"value"`
	Operator string            `json:"operator"`
	Link     string            `json:"link,omitempty"`
	Filter   []*ZoneListFilter `json:"filters,omitempty"`
}

// ZoneListRequest describes a JSON zone list request.
type ZoneListRequest struct {
	Filter []*ZoneListFilter `json:"filters"`
}

func (api *autoDnsProvider) request(method string, requestPath string, data interface{}) ([]byte, error) {
	client := &http.Client{}

	requestURL := api.baseURL
	requestURL.Path = api.baseURL.Path + requestPath

	request := &http.Request{
		URL:    &requestURL,
		Header: api.defaultHeaders,
		Method: method,
	}

	if data != nil {
		body, _ := json.Marshal(data)
		buffer := bytes.NewBuffer(body)
		request.Body = io.NopCloser(buffer)
	}

	response, error := client.Do(request)
	if error != nil {
		return nil, error
	}
	defer response.Body.Close()

	responseText, _ := io.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return nil, errors.New("Request to " + requestURL.Path + " failed: " + string(responseText))
	}

	return responseText, nil
}

func (api *autoDnsProvider) findZoneSystemNameServer(domain string) (*models.Nameserver, error) {
	request := &ZoneListRequest{}

	request.Filter = append(request.Filter, &ZoneListFilter{
		Key:      "name",
		Value:    domain,
		Operator: "EQUAL",
	})

	responseData, err := api.request("POST", "zone/_search", request)
	if err != nil {
		return nil, err
	}

	var responseObject JSONResponseDataZone
	_ = json.Unmarshal(responseData, &responseObject)
	if len(responseObject.Data) != 1 {
		return nil, errors.New("Domain " + domain + " could not be found in AutoDNS")
	}

	systemNameServer := &models.Nameserver{Name: responseObject.Data[0].SystemNameServer}

	return systemNameServer, nil
}

func (api *autoDnsProvider) getZone(domain string) (*Zone, error) {
	systemNameServer, err := api.findZoneSystemNameServer(domain)
	if err != nil {
		return nil, err
	}

	// if resolving of a systemNameServer succeeds the system contains this zone
	var responseData, _ = api.request("GET", "zone/"+domain+"/"+systemNameServer.Name, nil)
	var responseObject JSONResponseDataZone
	// make sure that the response is valid, the zone is in AutoDNS but we're not sure the returned data meets our expectation
	unmErr := json.Unmarshal(responseData, &responseObject)
	if unmErr != nil {
		return nil, unmErr
	}

	return responseObject.Data[0], nil
}

func (api *autoDnsProvider) updateZone(domain string, resourceRecords []*ResourceRecord, nameServers []*models.Nameserver, zoneTTL uint32) error {
	systemNameServer, err := api.findZoneSystemNameServer(domain)

	if err != nil {
		return err
	}

	zone, _ := api.getZone(domain)

	zone.Origin = domain
	zone.SystemNameServer = systemNameServer.Name

	zone.IncludeWwwForMain = false

	zone.Soa.TTL = zoneTTL

	// empty out NameServers and ResourceRecords, add what it should be
	zone.NameServers = []*models.Nameserver{}
	zone.ResourceRecords = []*ResourceRecord{}

	zone.ResourceRecords = append(zone.ResourceRecords, resourceRecords...)

	// naive approach, the first nameserver passed should be the systemNameServer, the will be named alphabetically
	sort.Slice(nameServers, func(i, j int) bool {
		return nameServers[i].Name < nameServers[j].Name
	})

	zone.NameServers = append(zone.NameServers, nameServers...)

	var _, putErr = api.request("PUT", "zone/"+domain+"/"+systemNameServer.Name, zone)

	if putErr != nil {
		return putErr
	}

	return nil
}
