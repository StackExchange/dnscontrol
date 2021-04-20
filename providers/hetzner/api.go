package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL = "https://dns.hetzner.com/api/v1"
)

type hetznerProvider struct {
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

func getHomogenousDelay(headers http.Header, quotaName string) (time.Duration, error) {
	quota, err := parseHeaderAsInt(headers, "X-Ratelimit-Limit-"+strings.Title(quotaName))
	if err != nil {
		return 0, err
	}

	var unit time.Duration
	switch quotaName {
	case "hour":
		unit = time.Hour
	case "minute":
		unit = time.Minute
	case "second":
		unit = time.Second
	}

	delay := time.Duration(int64(unit) / quota)
	return delay, nil
}

func getRetryAfterDelay(header http.Header) (time.Duration, error) {
	retryAfter, err := parseHeaderAsInt(header, "Retry-After")
	if err != nil {
		return 0, err
	}
	delay := time.Duration(retryAfter * int64(time.Second))
	return delay, nil
}

func parseHeaderAsInt(headers http.Header, headerName string) (int64, error) {
	value, ok := headers[headerName]
	if !ok {
		return 0, fmt.Errorf("header %q is missing", headerName)
	}
	return strconv.ParseInt(value[0], 10, 0)
}

func (api *hetznerProvider) bulkCreateRecords(records []record) error {
	for _, record := range records {
		if err := checkIsLockedSystemRecord(record); err != nil {
			return err
		}
	}

	request := bulkCreateRecordsRequest{
		Records: records,
	}
	return api.request("/records/bulk", "POST", request, nil)
}

func (api *hetznerProvider) bulkUpdateRecords(records []record) error {
	for _, record := range records {
		if err := checkIsLockedSystemRecord(record); err != nil {
			return err
		}
	}

	request := bulkUpdateRecordsRequest{
		Records: records,
	}
	return api.request("/records/bulk", "PUT", request, nil)
}

func (api *hetznerProvider) createRecord(record record) error {
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

func (api *hetznerProvider) createZone(name string) error {
	request := createZoneRequest{
		Name: name,
	}
	return api.request("/zones", "POST", request, nil)
}

func (api *hetznerProvider) deleteRecord(record record) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	url := fmt.Sprintf("/records/%s", record.ID)
	return api.request(url, "DELETE", nil, nil)
}

func (api *hetznerProvider) getAllRecords(domain string) ([]record, error) {
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

func (api *hetznerProvider) getAllZones() error {
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

func (api *hetznerProvider) getZone(name string) (*zone, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	zone, ok := api.zones[name]
	if !ok {
		return nil, fmt.Errorf("%q is not a zone in this HETZNER account", name)
	}
	return &zone, nil
}

func (api *hetznerProvider) request(endpoint string, method string, request interface{}, target interface{}) error {
	for {
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

		api.requestRateLimiter.handleResponse(*resp)
		// retry the request when rate-limited
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

func (api *hetznerProvider) startRateLimited() {
	// _Now_ is the best reference we can get for the last request.
	// Head-On-Head invocations of DNSControl benefit from fewer initial
	//  rate-limited requests.
	api.requestRateLimiter.lastRequest = time.Now()
	// use the default delay until we have had a chance to parse limits.
	api.requestRateLimiter.setDefaultDelay()
}

func (api *hetznerProvider) updateRecord(record record) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	url := fmt.Sprintf("/records/%s", record.ID)
	return api.request(url, "PUT", record, nil)
}

type requestRateLimiter struct {
	delay                     time.Duration
	lastRequest               time.Time
	optimizeForRateLimitQuota string
}

func (requestRateLimiter *requestRateLimiter) afterRequest() {
	requestRateLimiter.lastRequest = time.Now()
}

func (requestRateLimiter *requestRateLimiter) beforeRequest() {
	if requestRateLimiter.delay == 0 {
		return
	}
	time.Sleep(time.Until(requestRateLimiter.lastRequest.Add(requestRateLimiter.delay)))
}

func (requestRateLimiter *requestRateLimiter) setDefaultDelay() {
	// default to a rate-limit of 1 req/s -- the next response should update it.
	requestRateLimiter.delay = time.Second
}

func (requestRateLimiter *requestRateLimiter) setOptimizeForRateLimitQuota(quota string) error {
	quotaNormalized := strings.ToLower(quota)
	switch quotaNormalized {
	case "hour", "minute", "second":
		requestRateLimiter.optimizeForRateLimitQuota = quotaNormalized
	case "":
		requestRateLimiter.optimizeForRateLimitQuota = "second"
	default:
		return fmt.Errorf("%q is not a valid quota, expected 'Hour', 'Minute', 'Second' or unset", quota)
	}
	return nil
}

func (requestRateLimiter *requestRateLimiter) handleRateLimitedRequest() {
	message := "Rate-Limited, consider bumping the setting 'optimize_for_rate_limit_quota': %q -> %q"
	switch requestRateLimiter.optimizeForRateLimitQuota {
	case "hour":
		message = "Rate-Limited, you are already using the slowest request rate. Consider contacting the Hetzner Support for raising your quota."
	case "minute":
		message = fmt.Sprintf(message, "Minute", "Hour")
	case "second":
		message = fmt.Sprintf(message, "Second", "Minute")
	}
	fmt.Println(message)
}

func (requestRateLimiter *requestRateLimiter) handleResponse(resp http.Response) {
	homogenousDelay, err := getHomogenousDelay(resp.Header, requestRateLimiter.optimizeForRateLimitQuota)
	if err != nil {
		requestRateLimiter.setDefaultDelay()
		return
	}

	delay := homogenousDelay
	if resp.StatusCode == 429 {
		retryAfterDelay, err := getRetryAfterDelay(resp.Header)
		if err == nil {
			delay = retryAfterDelay
		}
	}
	requestRateLimiter.delay = delay
}
