package fortigate

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//
// Structure
//

// apiClient wraps all HTTP traffic to endpoints of the form:
//
//	https://<host>/api/v2/cmdb/<path>?vdom=<vdom>&datasource=1
type apiClient struct {
	base  string       // e.g. "https://fw.example.com/api/v2/cmdb/"
	vdom  string       // target VDOM
	key   string       // API token (Bearer)
	debug bool         // Debug Mode
	http  *http.Client // configured HTTP client
}

// fgDNSRecord represents a single entry inside the FortiGate dns-entry array.
// It is used for both JSON decoding (GET) and encoding (PUT/POST).
type fgDNSRecord struct {
	ID            int    `json:"id,omitempty"`             // FortiGate uses 1-based IDs
	Status        string `json:"status"`                   // "enable" / "disable"
	Type          string `json:"type"`                     // A, AAAA, CNAME, NS, PTR …
	TTL           uint32 `json:"ttl"`                      // 0 = inherit zone TTL
	Preference    uint16 `json:"preference,omitempty"`     // MX/SRV (not used yet)
	IP            string `json:"ip,omitempty"`             // A / PTR
	IPv6          string `json:"ipv6,omitempty"`           // AAAA (FortiGate keeps "" for unused)
	Hostname      string `json:"hostname,omitempty"`       // record name / label
	CanonicalName string `json:"canonical-name,omitempty"` // CNAME/NS/PTR target
}

//
// Constructor
//

// newClient builds a new apiClient.
//
// Parameters:
//
//	host     – base URL with protocol, without trailing slash
//	vdom     – VDOM (tenant) to operate on
//	key      – REST API token (System ▸ Administrators ▸ REST API Admin)
//	insecure – true = skip TLS certificate verification (self‑signed, etc.)
func newClient(host, vdom, key string, insecure bool, debug bool) *apiClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}
	return &apiClient{
		base:  strings.TrimRight(host, "/") + "/api/v2/cmdb/",
		vdom:  vdom,
		key:   key,
		debug: debug,
		http: &http.Client{
			Transport: tr,
			Timeout:   20 * time.Second,
		},
	}
}

//
// Central request helper
//

// do executes a request.
//
// Arguments:
//
//	method – HTTP verb (GET, POST, PUT, DELETE …)
//	path   – part after /cmdb/, e.g. "system/dns-database"
//	qs     – optional query parameters; vdom/datasource added automatically
//	body   – request body (struct, map, etc.) or nil
//	out    – pointer to struct for JSON decode or nil
//
// A non‑2xx HTTP status is returned as error.
// If out ≠ nil, the JSON response body is decoded into it.
func (c *apiClient) do(method, path string, qs url.Values, body any, out any) error {
	//
	// Build query string
	//
	if qs == nil {
		qs = url.Values{}
	}
	qs.Set("vdom", c.vdom)    // mandatory
	qs.Set("datasource", "1") // same as used by the web UI

	u := c.base + strings.TrimLeft(path, "/") + "?" + qs.Encode()

	//
	// Serialize body (if any)
	//
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)

		if c.debug {
			fmt.Printf("[FORTIGATE] [DEBUG] %s %s\nPayload:\n%s\n", method, u, string(b))
		}

	} else if c.debug {
		fmt.Printf("[FORTIGATE] [DEBUG] %s %s\n", method, u)
	}

	//
	// Build request
	//
	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.key)

	//
	// Execute request
	//

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	//
	// Read response body (once)
	//
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("[FORTIGATE] Failed to read response: %w", err)
	}

	//
	// Handle non‑success status codes
	//
	if resp.StatusCode >= 300 {
		return fmt.Errorf("[FORTIGATE] %s %s → %s: %s", method, path, resp.Status, strings.TrimSpace(string(respBytes)))
	}

	//
	// Optionally decode JSON response
	//
	if out != nil {
		if err := json.Unmarshal(respBytes, out); err != nil {
			return fmt.Errorf("[FORTIGATE] Failed to decode json: %w", err)
		}
		if c.debug {
			fmt.Printf("[FORTIGATE] [DEBUG] Response:\n%s\n", prettyJSON(respBytes))
		}
	}
	return nil
}

//
// Helper
//

// isNotFound returns true if the error represents a 404 Not Found response.
func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "404") && strings.Contains(strings.ToLower(msg), "not found")
}

func prettyJSON(b []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	if err != nil {
		return string(b) // Fallback: raw JSON
	}
	return out.String()
}
