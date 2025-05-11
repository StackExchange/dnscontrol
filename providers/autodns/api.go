package autodns

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
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

var (
	// Default retry configuration
	defaultRetryWait = 3 * time.Second
	defaultRetryMax  = 4
)

func (api *autoDNSProvider) request(method string, requestPath string, data interface{}) ([]byte, error) {
	var retryCounter = 0

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

	for {
		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		responseText, _ := io.ReadAll(response.Body)

		if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusTooManyRequests {
			return nil, errors.New("Request to " + requestURL.Path + " failed: " + string(responseText))
		}

		if response.StatusCode == http.StatusOK {
			return responseText, nil
		}

		if retryCounter == defaultRetryMax { // the condition stops matching
			break // break out of the loop
		}

		retryCounter++

		sleepDuration := time.Duration(math.Pow(2, float64(retryCounter)) * float64(defaultRetryWait))

		time.Sleep(sleepDuration)
	}

	return nil, errors.New("Failed to fetch" + requestURL.Path + " after 4 retries")
}

func (api *autoDNSProvider) findZoneSystemNameServer(domain string) (*models.Nameserver, error) {
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
	if err := json.Unmarshal(responseData, &responseObject); err != nil {
		return nil, err
	}

	if len(responseObject.Data) != 1 {
		return nil, fmt.Errorf("Zone "+domain+" could not be found in AutoDNS: %w", os.ErrNotExist)
	}

	systemNameServer := &models.Nameserver{Name: responseObject.Data[0].SystemNameServer}

	return systemNameServer, nil
}

func (api *autoDNSProvider) createZone(domain string, zone *Zone) (*Zone, error) {
	responseData, err := api.request("POST", "zone", zone)
	if err != nil {
		return nil, err
	}

	var responseObject JSONResponseDataZone
	if err := json.Unmarshal(responseData, &responseObject); err != nil {
		return nil, err
	}

	if len(responseObject.Data) != 1 {
		return nil, errors.New("Zone " + domain + " not returned")
	}

	return responseObject.Data[0], nil
}

func (api *autoDNSProvider) getZone(domain string) (*Zone, error) {
	systemNameServer, err := api.findZoneSystemNameServer(domain)
	if err != nil {
		return nil, err
	}

	// if resolving of a systemNameServer succeeds the system contains this zone
	responseData, err2 := api.request("GET", "zone/"+domain+"/"+systemNameServer.Name, nil)

	if err2 != nil {
		return nil, err2
	}

	var responseObject JSONResponseDataZone
	// make sure that the response is valid, the zone is in AutoDNS but we're not sure the returned data meets our expectation
	if err := json.Unmarshal(responseData, &responseObject); err != nil {
		return nil, err
	}

	return responseObject.Data[0], nil
}

func (api *autoDNSProvider) getZones() ([]string, error) {
	responseData, err := api.request("POST", "zone/_search", nil)
	if err != nil {
		return nil, err
	}

	var responseObject JSONResponseDataZone
	if err := json.Unmarshal(responseData, &responseObject); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(responseObject.Data))
	for _, zone := range responseObject.Data {
		names = append(names, zone.Origin)
	}

	return names, nil
}

func (api *autoDNSProvider) updateZone(domain string, resourceRecords []*ResourceRecord, nameServers []*models.Nameserver, zoneTTL uint32) error {
	systemNameServer, err := api.findZoneSystemNameServer(domain)
	if err != nil {
		return err
	}

	zone, err2 := api.getZone(domain)

	if err2 != nil {
		return err2
	}

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

	_, putErr := api.request("PUT", "zone/"+domain+"/"+systemNameServer.Name, zone)

	if putErr != nil {
		return putErr
	}

	return nil
}
