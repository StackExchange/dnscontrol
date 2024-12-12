package hetzner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
)

const (
	baseURL = "https://dns.hetzner.com/api/v1"
)

type hetznerProvider struct {
	apiKey             string
	mu                 sync.Mutex
	cachedZones        map[string]zone
	requestRateLimiter requestRateLimiter
}

func parseHeaderAsSeconds(header http.Header, headerName string, fallback time.Duration) (time.Duration, error) {
	retryAfter, err := parseHeaderAsInt(header, headerName, int64(fallback/time.Second))
	if err != nil {
		return 0, err
	}
	delay := time.Duration(retryAfter * int64(time.Second))
	return delay, nil
}

func parseHeaderAsInt(headers http.Header, headerName string, fallback int64) (int64, error) {
	v := headers.Get(headerName)
	if v == "" {
		return fallback, nil
	}
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i, nil
	}
	return 0, fmt.Errorf("expected header %q to contain number, got %q", headerName, v)
}

func (api *hetznerProvider) bulkCreateRecords(records []record) error {
	request := bulkCreateRecordsRequest{
		Records: records,
	}
	return api.request("/records/bulk", "POST", request, nil, nil)
}

func (api *hetznerProvider) bulkUpdateRecords(records []record) error {
	request := bulkUpdateRecordsRequest{
		Records: records,
	}
	return api.request("/records/bulk", "PUT", request, nil, nil)
}

func (api *hetznerProvider) createZone(name string) error {
	request := createZoneRequest{
		Name: name,
	}
	return api.request("/zones", "POST", request, nil, nil)
}

func (api *hetznerProvider) deleteRecord(record *record) error {
	url := fmt.Sprintf("/records/%s", record.ID)
	return api.request(url, "DELETE", nil, nil, nil)
}

func (api *hetznerProvider) getAllRecords(domain string) ([]record, error) {
	z, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	page := 1
	var records []record
	for {
		response := getAllRecordsResponse{}
		url := fmt.Sprintf("/records?zone_id=%s&per_page=100&page=%d", z.ID, page)
		if err = api.request(url, "GET", nil, &response, nil); err != nil {
			return nil, fmt.Errorf("failed fetching zone records for %q: %w", domain, err)
		}
		if records == nil {
			records = make([]record, 0, response.Meta.Pagination.TotalEntries)
		}
		for _, r := range response.Records {
			if r.TTL == nil {
				r.TTL = &z.TTL
			}
			if r.Type == "SOA" {
				// SOA records are not available for editing, hide them.
				continue
			}
			records = append(records, r)
		}
		// meta.pagination may not be present. In that case LastPage is 0 and below the current page number.
		if page >= response.Meta.Pagination.LastPage {
			break
		}
		page++
	}
	return records, nil
}

func (api *hetznerProvider) resetZoneCache() {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.cachedZones = nil
}

func (api *hetznerProvider) getAllZones() (map[string]zone, error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	if api.cachedZones != nil {
		return api.cachedZones, nil
	}
	var zones map[string]zone
	page := 1
	statusOK := func(code int) bool {
		switch code {
		case http.StatusOK:
			return true
		case http.StatusNotFound:
			// Accept a 404 when requesting the first page
			return page == 1
		default:
			return false
		}
	}
	for {
		response := getAllZonesResponse{}
		url := fmt.Sprintf("/zones?per_page=100&page=%d", page)
		if err := api.request(url, "GET", nil, &response, statusOK); err != nil {
			return nil, fmt.Errorf("failed fetching zones: %w", err)
		}
		if zones == nil {
			zones = make(map[string]zone, response.Meta.Pagination.TotalEntries)
		}
		for _, z := range response.Zones {
			zones[z.Name] = z
		}
		// meta.pagination may not be present. In that case LastPage is 0 and below the current page number.
		if page >= response.Meta.Pagination.LastPage {
			break
		}
		page++
	}
	api.cachedZones = zones
	return zones, nil
}

func (api *hetznerProvider) getZone(name string) (*zone, error) {
	zones, err := api.getAllZones()
	if err != nil {
		return nil, err
	}
	z, ok := zones[name]
	if !ok {
		return nil, fmt.Errorf("%q is not a zone in this HETZNER account", name)
	}
	return &z, nil
}

func (api *hetznerProvider) request(endpoint string, method string, request interface{}, target interface{}, statusOK func(code int) bool) error {
	if statusOK == nil {
		statusOK = func(code int) bool {
			return code == http.StatusOK
		}
	}
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

		api.requestRateLimiter.delayRequest()
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		cleanupResponseBody := func() {
			err2 := resp.Body.Close()
			if err2 != nil {
				printer.Printf("failed closing response body: %q\n", err2)
			}
		}

		retry, err := api.requestRateLimiter.handleResponse(resp)
		if err != nil {
			cleanupResponseBody()
			return err
		}
		if retry {
			cleanupResponseBody()
			continue
		}

		if !statusOK(resp.StatusCode) {
			data, _ := io.ReadAll(resp.Body)
			printer.Println(string(data))
			cleanupResponseBody()
			return fmt.Errorf("bad status code from HETZNER: %d not 200", resp.StatusCode)
		}
		if target == nil {
			cleanupResponseBody()
			return nil
		}
		err = json.NewDecoder(resp.Body).Decode(target)
		cleanupResponseBody()
		return err
	}
}

type requestRateLimiter struct {
	mu          sync.Mutex
	delay       time.Duration
	lastRequest time.Time
	resetAt     time.Time
}

func (rrl *requestRateLimiter) delayRequest() {
	rrl.mu.Lock()
	// When not rate-limited, include network/server latency in delay.
	next := rrl.lastRequest.Add(rrl.delay)
	if next.After(rrl.resetAt) {
		// Do not stack delays past the reset point.
		next = rrl.resetAt
	}
	rrl.lastRequest = next
	rrl.mu.Unlock()
	time.Sleep(time.Until(next))
}

func (rrl *requestRateLimiter) handleResponse(resp *http.Response) (bool, error) {
	rrl.mu.Lock()
	defer rrl.mu.Unlock()
	if resp.StatusCode == http.StatusTooManyRequests {
		printer.Printf("Rate-Limited. Consider contacting the Hetzner Support for raising your quota. URL: %q, Headers: %q\n", resp.Request.URL, resp.Header)

		retryAfter, err := parseHeaderAsSeconds(resp.Header, "Retry-After", time.Second)
		if err != nil {
			return false, err
		}
		rrl.delay = retryAfter

		// When rate-limited, exclude network/server latency from delay.
		rrl.lastRequest = time.Now()
		return true, nil
	}

	limit, err := parseHeaderAsInt(resp.Header, "Ratelimit-Limit", 1)
	if err != nil {
		return false, err
	}

	remaining, err := parseHeaderAsInt(resp.Header, "Ratelimit-Remaining", 1)
	if err != nil {
		return false, err
	}

	reset, err := parseHeaderAsSeconds(resp.Header, "Ratelimit-Reset", 0)
	if err != nil {
		return false, err
	}

	if remaining == 0 {
		// Quota exhausted. Wait until quota resets.
		rrl.delay = reset
	} else if remaining > limit/2 {
		// Burst through half of the quota, ...
		rrl.delay = 0
	} else {
		// ... then spread requests evenly throughout the window.
		rrl.delay = reset / time.Duration(remaining+1)
	}
	rrl.resetAt = time.Now().Add(reset)
	return false, nil
}
