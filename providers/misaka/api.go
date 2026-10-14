package misaka

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/DNSControl/dnscontrol/v5/pkg/printer"
	"github.com/DNSControl/dnscontrol/v5/pkg/version"
)

const (
	// defaultAPIURL is the base URL of the v3 API.
	defaultAPIURL = "https://api.misaka.io/v3"

	// pageSize is the largest page the list endpoints accept.
	pageSize = 100

	// maxErrorBody bounds how much of a body that is not the API's own error
	// document (a Cloudflare block page, for example) goes into an error.
	maxErrorBody = 512

	// The API occasionally drops a connection in the middle of a request: the
	// client sees "unexpected EOF" and no HTTP status. It happened in three
	// of four full integration runs, on a GET, a PUT and a DELETE. A
	// request that cannot create anything twice is sent once more after a
	// short pause; a POST is not, because the first one may have been applied.
	maxAttempts = 2
	retryPause  = time.Second
)

type apiClient struct {
	baseURL    string
	apiToken   string
	userAgent  string
	httpClient *http.Client

	// sleep is time.Sleep, replaced in tests so that a retry does not cost
	// the wall-clock time of its pause.
	sleep func(time.Duration)
}

// listResponse is the envelope of every list endpoint. There are no next or
// previous links; Count is the total number of items.
type listResponse[T any] struct {
	Results []T `json:"results"`
	Count   int `json:"count"`
}

// apiZone is a zone as returned by GET /dns/zones.
type apiZone struct {
	Name string `json:"name"`
}

// apiRecordSet is a record set as returned by GET /dns/zones/{domain}/recordsets.
// Name is relative to the zone, "@" at the apex. Each value is one
// presentation-format string, stored exactly as it was written.
type apiRecordSet struct {
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	TTL       uint32            `json:"ttl"`
	Enabled   bool              `json:"enabled"`
	Protected bool              `json:"protected"`
	Alias     *string           `json:"alias"`
	Records   []apiRecord       `json:"records"`
	Filters   []json.RawMessage `json:"filters"`
}

type apiRecord struct {
	Value string `json:"value"`
}

// recordSetRequest is the body of a record set create or replace. A PUT is a
// partial update on the API side, so every field is always sent: that is what
// makes it a replacement of the whole set.
type recordSetRequest struct {
	TTL     uint32      `json:"ttl"`
	Records []apiRecord `json:"records"`
	Filters []any       `json:"filters"`
}

// apiError is a response with an HTTP status of 400 or above.
type apiError struct {
	Method     string
	Path       string
	StatusCode int
	Detail     string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("MISAKA: %s %s: HTTP %d: %s", e.Method, e.Path, e.StatusCode, e.Detail)
}

// isNotFound reports whether err is the API saying the zone or record set does not exist.
func isNotFound(err error) bool {
	var apiErr *apiError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

func newAPIClient(baseURL, token string) *apiClient {
	return &apiClient{
		baseURL:  baseURL,
		apiToken: token,
		// Cloudflare in front of the API rejects the default agents of
		// scripting libraries outright.
		userAgent: "dnscontrol/" + version.Version(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		sleep: time.Sleep,
	}
}

func (c *apiClient) listZones() ([]apiZone, error) {
	return listAll[apiZone](c, "/dns/zones")
}

func (c *apiClient) createZone(domain string) error {
	return c.doRequest(http.MethodPost, "/dns/zones", nil, map[string]string{"name": domain}, nil)
}

// getZoneSettings is used to find out whether a zone exists; the settings
// themselves are not needed.
func (c *apiClient) getZoneSettings(domain string) error {
	return c.doRequest(http.MethodGet, zonePath(domain)+"/settings", nil, nil, nil)
}

func (c *apiClient) listRecordSets(domain string) ([]apiRecordSet, error) {
	return listAll[apiRecordSet](c, zonePath(domain)+"/recordsets")
}

func (c *apiClient) getRecordSet(domain, name, rtype string) (*apiRecordSet, error) {
	var rs apiRecordSet
	if err := c.doRequest(http.MethodGet, recordSetPath(domain, name, rtype), nil, nil, &rs); err != nil {
		return nil, err
	}
	return &rs, nil
}

func (c *apiClient) createRecordSet(domain, name, rtype string, req recordSetRequest) error {
	return c.doRequest(http.MethodPost, recordSetPath(domain, name, rtype), nil, req, nil)
}

func (c *apiClient) updateRecordSet(domain, name, rtype string, req recordSetRequest) error {
	return c.doRequest(http.MethodPut, recordSetPath(domain, name, rtype), nil, req, nil)
}

func (c *apiClient) deleteRecordSet(domain, name, rtype string) error {
	return c.doRequest(http.MethodDelete, recordSetPath(domain, name, rtype), nil, nil, nil)
}

func zonePath(domain string) string {
	return "/dns/zones/" + url.PathEscape(domain)
}

func recordSetPath(domain, name, rtype string) string {
	return zonePath(domain) + "/recordsets/" + url.PathEscape(name) + "/" + rtype
}

// listAll pages through a list endpoint until it has the number of items the
// API says there are. A short page is not the end of the list; a page that
// adds nothing before then is an error rather than a silently incomplete zone.
func listAll[T any](c *apiClient, path string) ([]T, error) {
	var all []T
	for {
		query := url.Values{}
		query.Set("limit", strconv.Itoa(pageSize))
		query.Set("offset", strconv.Itoa(len(all)))

		var page listResponse[T]
		if err := c.doRequest(http.MethodGet, path, query, nil, &page); err != nil {
			return nil, err
		}

		all = append(all, page.Results...)
		if len(all) >= page.Count {
			return all, nil
		}
		if len(page.Results) == 0 {
			return nil, fmt.Errorf("MISAKA: GET %s: API returned %d of %d items", path, len(all), page.Count)
		}
	}
}

// doRequest performs one API call and decodes the response into result.
func (c *apiClient) doRequest(method, path string, query url.Values, body, result any) error {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var status int
	var respBody []byte
	for attempt := 1; ; attempt++ {
		var err error
		status, respBody, err = c.send(method, u, payload)
		if err == nil {
			// The first DELETE may have been applied before its connection
			// was dropped.
			if attempt > 1 && method == http.MethodDelete && status == http.StatusNotFound {
				return nil
			}
			break
		}
		if method == http.MethodPost || attempt == maxAttempts {
			return err
		}
		printer.Warnf("MISAKA: %v; retrying %s %s\n", err, method, path)
		c.sleep(retryPause)
	}

	if status >= 400 {
		return &apiError{
			Method:     method,
			Path:       path,
			StatusCode: status,
			Detail:     errorDetail(respBody),
		}
	}
	if result == nil {
		return nil
	}
	if err := json.Unmarshal(respBody, result); err != nil {
		return fmt.Errorf("MISAKA: %s %s: decoding response: %w", method, path, err)
	}
	return nil
}

// send sends one request and reads the whole response.
func (c *apiClient) send(method, u string, payload []byte) (int, []byte, error) {
	var bodyReader io.Reader
	if payload != nil {
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, u, bodyReader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Token "+c.apiToken)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, respBody, nil
}

// errorDetail extracts the message of an error response. The API answers
// {"detail": "..."} on 401, 403 and 404, and {"detail": [{"detail": "..."}, ...]}
// on a 422 validation error. Anything else, such as a Cloudflare block page,
// is reported as the beginning of the raw body.
func errorDetail(body []byte) string {
	var envelope struct {
		Detail json.RawMessage `json:"detail"`
	}
	if json.Unmarshal(body, &envelope) == nil && len(envelope.Detail) > 0 {
		var single string
		if json.Unmarshal(envelope.Detail, &single) == nil {
			return single
		}

		var list []struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(envelope.Detail, &list) == nil && len(list) > 0 {
			msgs := make([]string, 0, len(list))
			for _, d := range list {
				msgs = append(msgs, d.Detail)
			}
			return strings.Join(msgs, "; ")
		}
	}

	raw := strings.TrimSpace(string(body))
	if len(raw) > maxErrorBody {
		raw = raw[:maxErrorBody] + "..."
	}
	return raw
}
