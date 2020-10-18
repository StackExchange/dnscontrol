package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	baseURL = "https://dns.hetzner.com/api/v1"
)

type api struct {
	apiKey             string
	zones              map[string]zone
	requestRateLimiter requestRateLimiter
}

func checkIsLockedSystemRecord(record record) error {
	if record.Type == "SOA" {
		// The upload of a BIND zone file can change the SOA record.
		// Implementing this edge case this is too complex for now.
		return fmt.Errorf("SOA records are locked in HETZNER zones. They are hence not available for updating")
	}
	return nil
}

func (api *api) createRecord(record record) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	request := createRecordRequest{
		Name:   record.Name,
		TTL:    *record.TTL,
		Type:   record.Type,
		Value:  record.Value,
		ZoneID: record.ZoneID,
	}
	return api.request("/records", "POST", request, nil)
}

func (api *api) createZone(name string) error {
	request := createZoneRequest{
		Name: name,
	}
	return api.request("/zones", "POST", request, nil)
}

func (api *api) deleteRecord(record record) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	url := fmt.Sprintf("/records/%s", record.ID)
	return api.request(url, "DELETE", nil, nil)
}

func (api *api) getAllRecords(domain string) ([]record, error) {
	zone, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	page := 1
	records := make([]record, 0)
	for {
		response := &getAllRecordsResponse{}
		url := fmt.Sprintf("/records?zone_id=%s&per_page=100&page=%d", zone.ID, page)
		if err := api.request(url, "GET", nil, response); err != nil {
			return nil, fmt.Errorf("failed fetching zone records for %q: %w", domain, err)
		}
		for _, record := range response.Records {
			if record.TTL == nil {
				record.TTL = &zone.TTL
			}

			if checkIsLockedSystemRecord(record) != nil {
				// Some records are not available for updating, hide them.
				continue
			}

			records = append(records, record)
		}
		// meta.pagination may not be present. In that case LastPage is 0 and below the current page number.
		if page >= response.Meta.Pagination.LastPage {
			break
		}
		page++
	}
	return records, nil
}

func (api *api) getAllZones() error {
	if api.zones != nil {
		return nil
	}
	zones := map[string]zone{}
	page := 1
	for {
		response := &getAllZonesResponse{}
		url := fmt.Sprintf("/zones?per_page=100&page=%d", page)
		if err := api.request(url, "GET", nil, response); err != nil {
			return fmt.Errorf("failed fetching zones: %w", err)
		}
		for _, zone := range response.Zones {
			zones[zone.Name] = zone
		}
		// meta.pagination may not be present. In that case LastPage is 0 and below the current page number.
		if page >= response.Meta.Pagination.LastPage {
			break
		}
		page++
	}
	api.zones = zones
	return nil
}

func (api *api) getZone(name string) (*zone, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	zone, ok := api.zones[name]
	if !ok {
		return nil, fmt.Errorf("%q is not a zone in this HETZNER account", name)
	}
	return &zone, nil
}

func (api *api) request(endpoint string, method string, request interface{}, target interface{}) error {
	var requestBody io.Reader
	if request != nil {
		requestBodySerialised, err := json.Marshal(request)
		if err != nil {
			return err
		}
		requestBody = bytes.NewBuffer(requestBodySerialised)
	}
	req, err := http.NewRequest(method, baseURL+endpoint, requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Auth-API-Token", api.apiKey)

	for {
		api.requestRateLimiter.beforeRequest()
		resp, err := http.DefaultClient.Do(req)
		api.requestRateLimiter.afterRequest()
		if err != nil {
			return err
		}
		cleanupResponseBody := func() {
			err := resp.Body.Close()
			if err != nil {
				fmt.Println(fmt.Sprintf("failed closing response body: %q", err))
			}
		}

		if resp.StatusCode == 429 {
			api.requestRateLimiter.handleRateLimitedRequest()
			cleanupResponseBody()
			continue
		}

		defer cleanupResponseBody()
		if resp.StatusCode != 200 {
			data, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(data))
			return fmt.Errorf("bad status code from HETZNER: %d not 200", resp.StatusCode)
		}
		if target == nil {
			return nil
		}
		decoder := json.NewDecoder(resp.Body)
		return decoder.Decode(target)
	}
}

func (api *api) startRateLimited() {
	// Simulate a request that is getting a 429 response.
	api.requestRateLimiter.afterRequest()
	api.requestRateLimiter.bumpDelay()
}

func (api *api) updateRecord(record record) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	url := fmt.Sprintf("/records/%s", record.ID)
	return api.request(url, "PUT", record, nil)
}

type requestRateLimiter struct {
	delay       time.Duration
	lastRequest time.Time
}

func (rateLimiter *requestRateLimiter) afterRequest() {
	rateLimiter.lastRequest = time.Now()
}

func (rateLimiter *requestRateLimiter) beforeRequest() {
	if rateLimiter.delay == 0 {
		return
	}
	time.Sleep(time.Until(rateLimiter.lastRequest.Add(rateLimiter.delay)))
}

func (rateLimiter *requestRateLimiter) bumpDelay() string {
	var backoffType string
	if rateLimiter.delay == 0 {
		// At the time this provider was implemented (2020-10-18),
		//  one request per second could go though when rate-limited.
		rateLimiter.delay = time.Second
		backoffType = "constant"
	} else {
		// The initial assumption of 1 req/s may no hold true forever.
		// Future proof this provider, use exponential back-off.
		rateLimiter.delay = rateLimiter.delay * 2
		backoffType = "exponential"
	}
	return backoffType
}

func (rateLimiter *requestRateLimiter) handleRateLimitedRequest() {
	backoffType := rateLimiter.bumpDelay()
	fmt.Println(fmt.Sprintf("WARNING: request rate-limited, %s back-off is now at %s.", backoffType, rateLimiter.delay))
}
