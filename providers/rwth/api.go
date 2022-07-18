package rwth

// The documentation is hosted at https://noc-portal.rz.rwth-aachen.de/dns-admin/en/api_tokens and
// https://blog.rwth-aachen.de/itc/2022/07/13/api-im-dns-admin/

import (
	"encoding/json"
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/prettyzone"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/miekg/dns"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	baseURL = "https://noc-portal.rz.rwth-aachen.de/dns-admin/api/v1"
)

type rwthProvider struct {
	apiToken string
	zones    map[string]zone
}

// Custom dns.NewRR with RWTH default TTL
func NewRR(s string) (dns.RR, error) {
	if len(s) > 0 && s[len(s)-1] != '\n' { // We need a closing newline
		return ReadRR(strings.NewReader(s + "\n"))
	}
	return ReadRR(strings.NewReader(s))
}

func ReadRR(r io.Reader) (dns.RR, error) {
	zp := dns.NewZoneParser(r, ".", "")
	zp.SetDefaultTTL(172800)
	zp.SetIncludeAllowed(true)
	rr, _ := zp.Next()
	return rr, zp.Err()
}

func checkIsLockedSystemApiRecord(record RecordReply) error {
	if record.Type == "soa_record" {
		// The upload of a BIND zone file can change the SOA record.
		// Implementing this edge case this is too complex for now.
		return fmt.Errorf("SOA records are locked in RWTH zones. They are hence not available for updating")
	}
	return nil
}

func checkIsLockedSystemRecord(record *models.RecordConfig) error {
	if record.Type == "SOA" {
		// The upload of a BIND zone file can change the SOA record.
		// Implementing this edge case this is too complex for now.
		return fmt.Errorf("SOA records are locked in RWTH zones. They are hence not available for updating")
	}
	return nil
}

func (api *rwthProvider) createRecord(domain string, record *models.RecordConfig) error {
	if err := checkIsLockedSystemRecord(record); err != nil {
		return err
	}

	req := url.Values{}
	req.Set("record_content", api.printRecConfig(*record))
	return api.request("/create_record", "POST", req, nil, nil)
}

func (api *rwthProvider) destroyRecord(record RecordReply) error {
	if err := checkIsLockedSystemApiRecord(record); err != nil {
		return err
	}
	req := url.Values{}
	req.Set("record_id", strconv.Itoa(record.ID))
	return api.request("/destroy_record", "DELETE", req, nil, nil)
}

func (api *rwthProvider) updateRecord(id int, record models.RecordConfig) error {
	if err := checkIsLockedSystemRecord(&record); err != nil {
		return err
	}
	req := url.Values{}
	req.Set("record_id", strconv.Itoa(id))
	req.Set("record_content", api.printRecConfig(record))
	return api.request("/update_record", "POST", req, nil, nil)
}

func (api *rwthProvider) getAllRecords(domain string) ([]models.RecordConfig, error) {
	zone, err := api.getZone(domain)
	if err != nil {
		return nil, err
	}
	records := make([]models.RecordConfig, 0)
	response := &[]RecordReply{}
	request := url.Values{}
	request.Set("zone_id", strconv.Itoa(zone.ID))
	if err := api.request("/list_records", "GET", request, response, nil); err != nil {
		return nil, fmt.Errorf("failed fetching zone records for %q: %w", domain, err)
	}
	for _, apiRecord := range *response {
		if checkIsLockedSystemApiRecord(apiRecord) != nil {
			continue
		}
		dnsRec, err := NewRR(apiRecord.Content) // Parse content as DNS record
		if err != nil {
			return nil, err
		}

		recConfig, err := models.RRtoRC(dnsRec, domain) // and make it a RC
		if err != nil {
			return nil, err
		}
		recConfig.Original = apiRecord // but keep our ApiRecord as the original

		records = append(records, recConfig)
	}
	return records, nil
}

func (api *rwthProvider) getAllZones() error {
	if api.zones != nil {
		return nil
	}
	zones := map[string]zone{}
	response := &[]zone{}
	if err := api.request("/list_zones", "GET", url.Values{}, response, nil); err != nil {
		return fmt.Errorf("failed fetching zones: %w", err)
	}
	for _, zone := range *response {
		zones[zone.ZoneName] = zone
	}
	api.zones = zones
	return nil
}

// Print the generateZoneFileHelper
func (apo *rwthProvider) printRecConfig(rr models.RecordConfig) string {
	// Similar to prettyzone
	// Fake types are commented out.
	prefix := ""
	_, ok := dns.StringToType[rr.Type]
	if !ok {
		prefix = ";"
	}

	// ttl
	ttl := ""
	if rr.TTL != 172800 && rr.TTL != 0 {
		ttl = fmt.Sprint(rr.TTL)
	}

	// type
	typeStr := rr.Type

	// the remaining line
	target := rr.GetTargetCombined()

	// comment
	comment := ";"

	return fmt.Sprintf("%s%s%s\n",
		prefix, prettyzone.FormatLine([]int{10, 5, 2, 5, 0}, []string{rr.NameFQDN, ttl, "IN", typeStr, target}), comment)
}

func (api *rwthProvider) getZone(name string) (*zone, error) {
	if err := api.getAllZones(); err != nil {
		return nil, err
	}
	zone, ok := api.zones[name]
	if !ok {
		return nil, fmt.Errorf("%q is not a zone in this RWTH account", name)
	}
	return &zone, nil
}

// Send a request
func (api *rwthProvider) request(endpoint string, method string, request url.Values, target interface{}, statusOK func(code int) bool) error {
	if statusOK == nil {
		statusOK = func(code int) bool {
			return code == http.StatusOK
		}
	}
	requestBody := strings.NewReader(request.Encode())
	req, err := http.NewRequest(method, baseURL+endpoint, requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("PRIVATE-TOKEN", api.apiToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	cleanupResponseBody := func() {
		err := resp.Body.Close()
		if err != nil {
			printer.Printf("failed closing response body: %q\n", err)
		}
	}

	defer cleanupResponseBody()
	if !statusOK(resp.StatusCode) {
		data, _ := ioutil.ReadAll(resp.Body)
		printer.Printf(string(data))
		return fmt.Errorf("bad status code from RWTH: %d not 200", resp.StatusCode)
	}
	if target == nil {
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder.Decode(target)
}

// Deploy the zone
func (api *rwthProvider) deployZone(domain string) error {
	zone, err := api.getZone(domain)
	if err != nil {
		return err
	}
	req := url.Values{}
	req.Set("zone_id", strconv.Itoa(zone.ID))
	return api.request("/deploy_zone", "POST", req, nil, nil)
}
