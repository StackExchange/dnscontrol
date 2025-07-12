package joker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/diff2"
	"github.com/StackExchange/dnscontrol/v4/providers"
)

/*

Joker DMAPI provider:

Info required in `creds.json`:
   - username
   - password
   OR
   - api-key

*/

var features = providers.DocumentationNotes{
	// The default for unlisted capabilities is 'Cannot'.
	// See providers/capabilities.go for the entire list of capabilities.
	providers.CanGetZones:            providers.Can(),
	providers.CanConcur:              providers.Cannot("Joker API has session-based authentication"),
	providers.CanUseAlias:            providers.Cannot(),
	providers.CanUseCAA:              providers.Can(),
	providers.CanUseDNSKEY:           providers.Cannot(),
	providers.CanUseDS:               providers.Cannot(),
	providers.CanUseDSForChildren:    providers.Cannot(),
	providers.CanUseHTTPS:            providers.Cannot(),
	providers.CanUseLOC:              providers.Cannot(),
	providers.CanUseNAPTR:            providers.Can(),
	providers.CanUsePTR:              providers.Cannot(),
	providers.CanUseSOA:              providers.Cannot(),
	providers.CanUseSRV:              providers.Can(),
	providers.CanUseSSHFP:            providers.Cannot(),
	providers.CanUseSVCB:             providers.Cannot(),
	providers.CanUseTLSA:             providers.Cannot(),
	providers.DocCreateDomains:       providers.Can(),
	providers.DocDualHost:            providers.Cannot(),
	providers.DocOfficiallySupported: providers.Cannot(),
}

func init() {
	const providerName = "JOKER"
	const providerMaintainer = "@alextrull"
	fns := providers.DspFuncs{
		Initializer:   newJoker,
		RecordAuditor: AuditRecords,
	}
	providers.RegisterDomainServiceProviderType(providerName, fns, features)
	providers.RegisterMaintainer(providerName, providerMaintainer)
}

// jokerProvider is the handle for API calls.
type jokerProvider struct {
	apiURL     string
	username   string
	password   string
	apiKey     string
	authSID    string
	httpClient *http.Client
}

// newJoker creates a new Joker DMAPI provider.
func newJoker(m map[string]string, metadata json.RawMessage) (providers.DNSServiceProvider, error) {
	api := &jokerProvider{
		apiURL:     "https://dmapi.joker.com/request/",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	// Check for authentication methods
	api.username = m["username"]
	api.password = m["password"]
	api.apiKey = m["api-key"]

	if api.apiKey == "" && (api.username == "" || api.password == "") {
		return nil, errors.New("missing Joker credentials: either 'api-key' or both 'username' and 'password' required")
	}

	// Authenticate to get session ID
	if err := api.authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return api, nil
}

// GetNameservers returns the nameservers for a domain.
func (api *jokerProvider) GetNameservers(domain string) ([]*models.Nameserver, error) {
	// For DNS-only providers like Joker, we can return an empty list
	// since nameserver management is typically handled separately
	return []*models.Nameserver{}, nil
}

// EnsureZoneExists creates a zone if it doesn't exist.
func (api *jokerProvider) EnsureZoneExists(domain string) error {
	// Check if zone already exists by trying to get it
	_, _, err := api.makeRequest("dns-zone-get", url.Values{"domain": {domain}})
	if err == nil {
		// Zone exists
		return nil
	}

	// Zone doesn't exist, but Joker automatically creates zones for domains you manage
	// We'll create an empty zone by putting an empty zone file
	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", "# Empty zone file created by dnscontrol")

	_, _, err = api.makeRequest("dns-zone-put", params)
	return err
}

// authenticate logs in to Joker DMAPI and stores the session ID.
func (api *jokerProvider) authenticate() error {
	data := url.Values{}

	if api.apiKey != "" {
		data.Set("api-key", api.apiKey)
	} else {
		data.Set("username", api.username)
		data.Set("password", api.password)
	}

	resp, err := api.httpClient.PostForm(api.apiURL+"login", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse the response headers and body
	respStr := string(body)
	headers, _ := api.parseResponse(respStr)

	if headers["Status-Code"] != "" && headers["Status-Code"] != "0" {
		return fmt.Errorf("login failed: %s", headers["Status-Text"])
	}

	authSID := headers["Auth-Sid"]
	if authSID == "" {
		return fmt.Errorf("no Auth-Sid received from login. Response: %s", respStr)
	}

	api.authSID = authSID
	return nil
}

// parseResponse parses the Joker DMAPI response format.
func (api *jokerProvider) parseResponse(response string) (map[string]string, string) {
	headers := make(map[string]string)
	lines := strings.Split(response, "\n")

	var bodyStart int
	for i, line := range lines {
		if line == "" {
			bodyStart = i + 1
			break
		}
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	body := ""
	if bodyStart < len(lines) {
		body = strings.Join(lines[bodyStart:], "\n")
	}

	return headers, body
}

// makeRequest makes an authenticated request to Joker DMAPI.
func (api *jokerProvider) makeRequest(endpoint string, params url.Values) (map[string]string, string, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("auth-sid", api.authSID)

	resp, err := api.httpClient.PostForm(api.apiURL+endpoint, params)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	headers, responseBody := api.parseResponse(string(body))

	if headers["Status-Code"] != "" && headers["Status-Code"] != "0" {
		return nil, "", fmt.Errorf("API error: %s", headers["Status-Text"])
	}

	return headers, responseBody, nil
}

// GetZoneRecords gets the records of a zone and returns them in RecordConfig format.
func (api *jokerProvider) GetZoneRecords(domain string, meta map[string]string) (models.Records, error) {
	params := url.Values{}
	params.Set("domain", domain)

	_, body, err := api.makeRequest("dns-zone-get", params)
	if err != nil {
		return nil, err
	}

	// Debug: print the raw zone data
	fmt.Printf("DEBUG: Raw zone data for %s:\n%s\n", domain, body)

	records, err := api.parseZoneRecords(domain, body)
	if err != nil {
		return nil, err
	}

	return records, nil
}

// parseZoneLine parses a zone file line while preserving quoted strings.
func parseZoneLine(line string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range line {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			current.WriteRune(r)
			continue
		}

		if r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
			continue
		}

		if !inQuotes && (r == ' ' || r == '\t') {
			// Skip multiple consecutive spaces
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	// Add the final part if any
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// parseZoneRecords parses Joker zone format into RecordConfig format.
func (api *jokerProvider) parseZoneRecords(domain, zoneData string) (models.Records, error) {
	var records models.Records

	lines := strings.Split(strings.TrimSpace(zoneData), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "$") {
			continue
		}

		// Parse the line while preserving quoted strings
		parts := parseZoneLine(line)
		if len(parts) < 4 {
			continue
		}

		label := parts[0]
		recordType := parts[1]
		priority := parts[2]
		target := parts[3]

		// Default TTL if not specified in zone record
		var ttl uint32 = 300
		if len(parts) >= 5 {
			if ttlParsed, err := strconv.ParseUint(parts[4], 10, 32); err == nil {
				ttl = uint32(ttlParsed)
			}
		}

		// Convert @ to empty string for root domain
		if label == "@" {
			label = ""
		}

		rc := &models.RecordConfig{
			TTL: ttl,
		}

		// Set the label and domain correctly
		rc.SetLabel(label, domain)

		// Handle different record types
		switch recordType {
		case "A", "AAAA":
			rc.Type = recordType
			if err := rc.SetTarget(target); err != nil {
				continue
			}
		case "CNAME":
			rc.Type = recordType
			// Ensure CNAME targets are fully qualified
			if !strings.HasSuffix(target, ".") {
				target = target + "."
			}
			if err := rc.SetTarget(target); err != nil {
				continue
			}
		case "TXT":
			rc.Type = recordType
			// Remove quotes from TXT records
			if strings.HasPrefix(target, "\"") && strings.HasSuffix(target, "\"") {
				target = strings.Trim(target, "\"")
			}
			if err := rc.SetTarget(target); err != nil {
				continue
			}
		case "MX":
			rc.Type = recordType
			if prio, err := strconv.ParseUint(priority, 10, 16); err == nil {
				rc.MxPreference = uint16(prio)
			}
			// Ensure MX targets are fully qualified
			if !strings.HasSuffix(target, ".") {
				target = target + "."
			}
			if err := rc.SetTarget(target); err != nil {
				continue
			}
		case "SRV":
			rc.Type = recordType
			// SRV format: priority/weight target:port
			if strings.Contains(priority, "/") {
				priorityParts := strings.Split(priority, "/")
				if len(priorityParts) == 2 {
					if prio, err := strconv.ParseUint(priorityParts[0], 10, 16); err == nil {
						rc.SrvPriority = uint16(prio)
					}
					if weight, err := strconv.ParseUint(priorityParts[1], 10, 16); err == nil {
						rc.SrvWeight = uint16(weight)
					}
				}
			}
			if strings.Contains(target, ":") {
				targetParts := strings.Split(target, ":")
				if len(targetParts) == 2 {
					if port, err := strconv.ParseUint(targetParts[1], 10, 16); err == nil {
						rc.SrvPort = uint16(port)
					}
					srvTarget := targetParts[0]
					// Ensure SRV targets are fully qualified
					if !strings.HasSuffix(srvTarget, ".") {
						srvTarget = srvTarget + "."
					}
					if err := rc.SetTarget(srvTarget); err != nil {
						continue
					}
				}
			}
		case "CAA":
			rc.Type = recordType
			// CAA format: flags tag "value"
			if len(parts) >= 7 {
				flags := parts[2]
				tag := parts[6]
				value := strings.Join(parts[7:], " ")
				value = strings.Trim(value, "\"")

				if flagsInt, err := strconv.ParseUint(flags, 10, 8); err == nil {
					rc.CaaFlag = uint8(flagsInt)
				}
				rc.CaaTag = tag
				if err := rc.SetTarget(value); err != nil {
					continue
				}
			}
		case "NAPTR":
			rc.Type = recordType
			// NAPTR format: order/preference replacement flags service regex
			if len(parts) >= 8 {
				if strings.Contains(priority, "/") {
					priorityParts := strings.Split(priority, "/")
					if len(priorityParts) == 2 {
						if order, err := strconv.ParseUint(priorityParts[0], 10, 16); err == nil {
							rc.NaptrOrder = uint16(order)
						}
						if pref, err := strconv.ParseUint(priorityParts[1], 10, 16); err == nil {
							rc.NaptrPreference = uint16(pref)
						}
					}
				}
				// Ensure NAPTR targets are fully qualified if they're not empty or "."
				if target != "" && target != "." && !strings.HasSuffix(target, ".") {
					target = target + "."
				}
				if err := rc.SetTarget(target); err != nil {
					continue
				}
				if len(parts) > 7 {
					rc.NaptrFlags = strings.Trim(parts[6], "\"")
				}
				if len(parts) > 8 {
					rc.NaptrService = strings.Trim(parts[7], "\"")
				}
				if len(parts) > 9 {
					rc.NaptrRegexp = strings.Trim(parts[8], "\"")
				}
			}
		default:
			// Skip unsupported record types
			continue
		}

		records = append(records, rc)
	}

	return records, nil
}

// GetZoneRecordsCorrections returns a list of corrections that will turn existing records into dc.Records.
func (api *jokerProvider) GetZoneRecordsCorrections(dc *models.DomainConfig, existingRecords models.Records) ([]*models.Correction, int, error) {
	var corrections []*models.Correction

	changes, actualChangeCount, err := diff2.ByRecord(existingRecords, dc, nil)
	if err != nil {
		return nil, 0, err
	}

	for _, change := range changes {
		switch change.Type {
		case diff2.REPORT:
			corrections = append(corrections, &models.Correction{Msg: change.MsgsJoined})
		case diff2.CREATE, diff2.CHANGE, diff2.DELETE:
			corrections = append(corrections, &models.Correction{
				Msg: change.MsgsJoined,
				F: func() error {
					return api.updateZone(dc.Name, dc.Records)
				},
			})
			// Only add one correction for zone update since we replace the entire zone
			break
		}
	}

	return corrections, actualChangeCount, nil
}

// updateZone replaces the entire zone with new records.
func (api *jokerProvider) updateZone(domain string, records models.Records) error {
	zoneData := api.recordsToZoneFormat(domain, records)

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", zoneData)

	_, _, err := api.makeRequest("dns-zone-put", params)
	return err
}

// recordsToZoneFormat converts RecordConfig records to Joker zone format.
func (api *jokerProvider) recordsToZoneFormat(domain string, records models.Records) string {
	var lines []string

	for _, rc := range records {
		label := rc.Name
		if label == "" {
			label = "@"
		}

		switch rc.Type {
		case "A", "AAAA":
			line := fmt.Sprintf("%s %s 0 %s %d", label, rc.Type, rc.GetTargetField(), rc.TTL)
			lines = append(lines, line)
		case "CNAME":
			line := fmt.Sprintf("%s %s 0 %s %d", label, rc.Type, rc.GetTargetField(), rc.TTL)
			lines = append(lines, line)
		case "MX":
			line := fmt.Sprintf("%s %s %d %s %d", label, rc.Type, rc.MxPreference, rc.GetTargetField(), rc.TTL)
			lines = append(lines, line)
		case "TXT":
			line := fmt.Sprintf("%s %s 0 \"%s\" %d", label, rc.Type, rc.GetTargetField(), rc.TTL)
			lines = append(lines, line)
		case "SRV":
			target := fmt.Sprintf("%s:%d", rc.GetTargetField(), rc.SrvPort)
			priority := fmt.Sprintf("%d/%d", rc.SrvPriority, rc.SrvWeight)
			line := fmt.Sprintf("%s %s %s %s %d", label, rc.Type, priority, target, rc.TTL)
			lines = append(lines, line)
		case "CAA":
			line := fmt.Sprintf("%s %s %d %s \"%s\" %d", label, rc.Type, rc.CaaFlag, rc.CaaTag, rc.GetTargetField(), rc.TTL)
			lines = append(lines, line)
		case "NAPTR":
			priority := fmt.Sprintf("%d/%d", rc.NaptrOrder, rc.NaptrPreference)
			line := fmt.Sprintf("%s %s %s %s 0 0 \"%s\" \"%s\" \"%s\" %d",
				label, rc.Type, priority, rc.GetTargetField(),
				rc.NaptrFlags, rc.NaptrService, rc.NaptrRegexp, rc.TTL)
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

// ListZones returns a list of zones managed by this provider.
func (api *jokerProvider) ListZones() ([]string, error) {
	_, body, err := api.makeRequest("dns-zone-list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	var zones []string
	lines := strings.Split(strings.TrimSpace(body), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			zones = append(zones, line)
		}
	}

	return zones, nil
}
