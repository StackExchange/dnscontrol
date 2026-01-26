package unifi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	cloudAPIBase = "https://api.ui.com"
)

// unifiClient handles HTTP communication with the UniFi API.
type unifiClient struct {
	host       string // Local: "https://10.19.80.1", Cloud: empty
	consoleID  string // Cloud access: console ID (e.g., "28704E24...:1008810555")
	apiKey     string
	site       string
	skipTLS    bool
	debug      bool
	httpClient *http.Client
}

// newClient creates a new UniFi API client.
func newClient(host, consoleID, apiKey, site string, skipTLS, debug bool) *unifiClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipTLS,
		},
	}

	return &unifiClient{
		host:      strings.TrimRight(host, "/"),
		consoleID: consoleID,
		apiKey:    apiKey,
		site:      site,
		skipTLS:   skipTLS,
		debug:     debug,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
	}
}

// isCloudAccess returns true if using cloud API access.
func (c *unifiClient) isCloudAccess() bool {
	return c.consoleID != ""
}

// buildURL constructs the full URL for an API endpoint.
// Local:  https://{host}/proxy/network/v2/api/site/{site}/static-dns
// Cloud:  https://api.ui.com/v1/connector/consoles/{consoleID}/proxy/network/v2/api/site/{site}/static-dns
func (c *unifiClient) buildURL(path string) string {
	if c.isCloudAccess() {
		return fmt.Sprintf("%s/v1/connector/consoles/%s/proxy/network%s",
			cloudAPIBase, c.consoleID, path)
	}
	return fmt.Sprintf("%s/proxy/network%s", c.host, path)
}

// do executes an HTTP request to the UniFi API.
func (c *unifiClient) do(method, path string, body any) ([]byte, error) {
	url := c.buildURL(path)

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		rdr = bytes.NewReader(b)
		if c.debug {
			fmt.Printf("[UNIFI] [DEBUG] %s %s\nPayload: %s\n", method, url, string(b))
		}
	} else if c.debug {
		fmt.Printf("[UNIFI] [DEBUG] %s %s\n", method, url)
	}

	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if c.debug {
		fmt.Printf("[UNIFI] [DEBUG] Response (%d): %s\n", resp.StatusCode, string(respBytes))
	}

	if resp.StatusCode >= 300 {
		return respBytes, fmt.Errorf("[UNIFI] %s %s -> %d: %s",
			method, path, resp.StatusCode, strings.TrimSpace(string(respBytes)))
	}

	return respBytes, nil
}

// getRecordsLegacy fetches all static DNS records using the OLD API.
func (c *unifiClient) getRecordsLegacy() ([]legacyDNSRecord, error) {
	path := fmt.Sprintf("/v2/api/site/%s/static-dns", c.site)

	respBytes, err := c.do("GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch records: %w", err)
	}

	var records []legacyDNSRecord
	if err := json.Unmarshal(respBytes, &records); err != nil {
		return nil, fmt.Errorf("failed to parse records: %w", err)
	}

	return records, nil
}

// createRecordLegacy creates a new DNS record using the OLD API.
func (c *unifiClient) createRecordLegacy(r map[string]any) (*legacyDNSRecord, error) {
	path := fmt.Sprintf("/v2/api/site/%s/static-dns", c.site)

	respBytes, err := c.do("POST", path, r)
	if err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	var created legacyDNSRecord
	if err := json.Unmarshal(respBytes, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created record: %w", err)
	}

	return &created, nil
}

// deleteRecordLegacy deletes a DNS record using the OLD API.
func (c *unifiClient) deleteRecordLegacy(id string) error {
	path := fmt.Sprintf("/v2/api/site/%s/static-dns/%s", c.site, id)

	_, err := c.do("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete record %s: %w", id, err)
	}

	return nil
}

// updateRecordLegacy updates a record by deleting and recreating it (OLD API has no PUT).
func (c *unifiClient) updateRecordLegacy(id string, r map[string]any) (*legacyDNSRecord, error) {
	// Delete old record
	if err := c.deleteRecordLegacy(id); err != nil {
		return nil, fmt.Errorf("update failed during delete: %w", err)
	}

	// Create new record
	created, err := c.createRecordLegacy(r)
	if err != nil {
		return nil, fmt.Errorf("update failed during create: %w", err)
	}

	return created, nil
}
