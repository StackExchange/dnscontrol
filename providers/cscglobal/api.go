package cscglobal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/v3/pkg/printer"

	"github.com/mattn/go-isatty"
)

const apiBase = "https://apis.cscglobal.com/dbs/api/v2"

// Api layer for CSC Global

type requestParams map[string]string

type errorResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Value       string `json:"value,omitempty"`
}

type nsModRequest struct {
	Domain        string   `json:"qualifiedDomainName"`
	NameServers   []string `json:"nameServers"`
	DNSType       string   `json:"dnsType,omitempty"`
	Notifications struct {
		Enabled bool     `json:"enabled,omitempty"`
		Emails  []string `json:"additionalNotificationEmails,omitempty"`
	} `json:"notifications"`
	ShowPrice    bool     `json:"showPrice,omitempty"`
	CustomFields []string `json:"customFields,omitempty"`
}

type nsModRequestResult struct {
	Result struct {
		Domain string `json:"qualifiedDomainName"`
		Status struct {
			Code                  string `json:"code"`
			Message               string `json:"message"`
			AdditionalInformation string `json:"additionalInformation"`
			UUID                  string `json:"uuid"`
		} `json:"status"`
	} `json:"result"`
}

type domainRecord struct {
	Nameserver []string `json:"nameservers"`
}

// Get zone

type nativeRecordA = struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	TTL    uint32 `json:"ttl"`
	Status string `json:"status"`
}
type nativeRecordCNAME = struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	TTL    uint32 `json:"ttl"`
	Status string `json:"status"`
}
type nativeRecordAAAA = struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	TTL    uint32 `json:"ttl"`
	Status string `json:"status"`
}
type nativeRecordTXT = struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	TTL    uint32 `json:"ttl"`
	Status string `json:"status"`
}
type nativeRecordMX = struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl"`
	Status   string `json:"status"`
	Priority uint16 `json:"priority"`
}
type nativeRecordNS = struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
}
type nativeRecordSRV = struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl"`
	Status   string `json:"status"`
	Priority uint16 `json:"priority"`
	Weight   uint16 `json:"weight"`
	Port     uint16 `json:"port"`
}
type nativeRecordCAA = struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Value  string `json:"value"`
	TTL    uint32 `json:"ttl"`
	Status string `json:"status"`
	Tag    string `json:"tag"`
	Flag   uint8  `json:"flag"`
}
type nativeRecordSOA = struct {
	Serial     int    `json:"serial"`
	Refresh    int    `json:"refresh"`
	Retry      int    `json:"retry"`
	Expire     int    `json:"expire"`
	TTL        uint32 `json:"ttlMin"`
	TTLNeg     int    `json:"ttlNeg"`
	TTLZone    int    `json:"ttlZone"`
	TechEmail  string `json:"techEmail"`
	MasterHost string `json:"masterHost"`
}

type zoneResponse struct {
	ZoneName    string              `json:"zoneName"`
	HostingType string              `json:"hostingType"`
	A           []nativeRecordA     `json:"a"`
	Cname       []nativeRecordCNAME `json:"cname"`
	Aaaa        []nativeRecordAAAA  `json:"aaaa"`
	Txt         []nativeRecordTXT   `json:"txt"`
	Mx          []nativeRecordMX    `json:"mx"`
	Ns          []nativeRecordNS    `json:"ns"`
	Srv         []nativeRecordSRV   `json:"srv"`
	Caa         []nativeRecordCAA   `json:"caa"`
	Soa         nativeRecordSOA     `json:"soa"`
}

// Zone edits

type zoneResourceRecordEdit = struct {
	Action       string `json:"action"`
	RecordType   string `json:"recordType"`
	CurrentKey   string `json:"currentKey,omitempty"`
	CurrentValue string `json:"currentValue,omitempty"`
	NewKey       string `json:"newKey,omitempty"`
	NewValue     string `json:"newValue,omitempty"`
	NewTTL       uint32 `json:"newTtl,omitempty"`
	// MX and SRV:
	NewPriority uint16 `json:"newPriority,omitempty"`
	// SRV:
	NewWeight uint16 `json:"newWeight,omitempty"`
	NewPort   uint16 `json:"newPort,omitempty"`
	// CAA:
	// These are pointers so that we can display the zero-value on demand.  If
	// they were not pointers, the zero-value ("" and 0) would result in no JSON
	// output for those fields.  Sometimes we want to generate fields with
	// zero-values, such as `"newTag":""`.  Thus we make these pointers. The
	// zero-value is now "nil".  If we want the field to appear in the JSON, we
	// set the pointer to a value. It is no longer nil, and will be output even
	// if the value at the pointer is zero-value.
	// See: https://emretanriverdi.medium.com/json-serialization-in-go-a27aeeb968de
	CurrentTag *string `json:"currentTag,omitempty"`
	NewTag     *string `json:"newTag,omitempty"`  // "" needs to be sent explicitly.
	NewFlag    *uint8  `json:"newFlag,omitempty"` // 0 needs to be sent explictly.
}

type zoneEditRequest = struct {
	ZoneName string                    `json:"zoneName"`
	Edits    *[]zoneResourceRecordEdit `json:"edits"`
}

type zoneEditRequestResultZoneEditRequestResult struct {
	Content struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"content"`
	Links struct {
		Self   string `json:"self"`
		Status string `json:"status"`
	} `json:"links"`
}

type zoneEditStatusResultZoneEditStatusResult struct {
	Content struct {
		Status           string `json:"status"`
		ErrorDescription string `json:"errorDescription"`
	} `json:"content"`
	Links struct {
		Cancel string `json:"cancel"`
	} `json:"links"`
}

func (client *providerClient) getNameservers(domain string) ([]string, error) {
	var bodyString, err = client.get("/domains/" + domain)
	if err != nil {
		return nil, err
	}

	var dr domainRecord
	json.Unmarshal(bodyString, &dr)
	ns := []string{}
	ns = append(ns, dr.Nameserver...)
	sort.Strings(ns)
	return ns, nil
}

func (client *providerClient) updateNameservers(ns []string, domain string) error {
	req := nsModRequest{
		Domain:      domain,
		NameServers: ns,
		DNSType:     "OTHER_DNS",
		ShowPrice:   false,
	}
	if client.notifyEmails != nil {
		req.Notifications.Enabled = true
		req.Notifications.Emails = client.notifyEmails
	}
	req.CustomFields = []string{}

	requestBody, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}

	bodyString, err := client.put("/domains/nsmodification", requestBody)
	if err != nil {
		return fmt.Errorf("CSC Global: Error update NS : %w", err)
	}

	var res nsModRequestResult
	json.Unmarshal(bodyString, &res)
	if res.Result.Status.Code != "SUBMITTED" {
		return fmt.Errorf("CSC Global: Error update NS Code: %s Message: %s AdditionalInfo: %s", res.Result.Status.Code, res.Result.Status.Message, res.Result.Status.AdditionalInformation)
	}

	return nil
}

// domainsResult is the JSON returned by "/domains".  Fields we don't
// use are commented out.
type domainsResult struct {
	Meta struct {
		NumResults int `json:"numResults"`
		Pages      int `json:"pages"`
	} `json:"meta"`
	Domains []struct {
		QualifiedDomainName string `json:"qualifiedDomainName"`
		//		Domain                   string        `json:"domain"`
		//		Idn                      string        `json:"idn"`
		//		Extension                string        `json:"extension"`
		//		NewGtld                  bool          `json:"newGtld"`
		//		ManagedStatus            string        `json:"managedStatus"`
		//		RegistrationDate         string        `json:"registrationDate"`
		//		RegistryExpiryDate       string        `json:"registryExpiryDate"`
		//		PaidThroughDate          string        `json:"paidThroughDate"`
		//		CountryCode              string        `json:"countryCode"`
		//		ServerDeleteProhibited   bool          `json:"serverDeleteProhibited"`
		//		ServerTransferProhibited bool          `json:"serverTransferProhibited"`
		//		ServerUpdateProhibited   bool          `json:"serverUpdateProhibited"`
		//		DNSType                  string        `json:"dnsType"`
		//		WhoisPrivacy             bool          `json:"whoisPrivacy"`
		//		LocalAgent               bool          `json:"localAgent"`
		//		DnssecActivated          string        `json:"dnssecActivated"`
		//		CriticalDomain           bool          `json:"criticalDomain"`
		//		BusinessUnit             string        `json:"businessUnit"`
		//		BrandName                string        `json:"brandName"`
		//		IdnReferenceName         string        `json:"idnReferenceName"`
		//		CustomFields             []interface{} `json:"customFields"`
		//		Account                  struct {
		//			AccountNumber string `json:"accountNumber"`
		//			AccountName   string `json:"accountName"`
		//		} `json:"account"`
		//		Urlf struct {
		//			RedirectType  string `json:"redirectType"`
		//			URLForwarding bool   `json:"urlForwarding"`
		//		} `json:"urlf"`
		//		NameServers   []string `json:"nameServers"`
		//		WhoisContacts []struct {
		//			ContactType   string `json:"contactType"`
		//			FirstName     string `json:"firstName"`
		//			LastName      string `json:"lastName"`
		//			Organization  string `json:"organization"`
		//			Street1       string `json:"street1"`
		//			Street2       string `json:"street2"`
		//			City          string `json:"city"`
		//			StateProvince string `json:"stateProvince"`
		//			Country       string `json:"country"`
		//			PostalCode    string `json:"postalCode"`
		//			Email         string `json:"email"`
		//			Phone         string `json:"phone"`
		//			PhoneExtn     string `json:"phoneExtn"`
		//			Fax           string `json:"fax"`
		//		} `json:"whoisContacts"`
		//		LastModifiedDate        string `json:"lastModifiedDate"`
		//		LastModifiedReason      string `json:"lastModifiedReason"`
		//		LastModifiedDescription string `json:"lastModifiedDescription"`
	} `json:"domains"`
	//	Links struct {
	//		Self string `json:"self"`
	//	} `json:"links"`
}

func (client *providerClient) getDomains() ([]string, error) {
	var bodyString, err = client.get("/domains")
	if err != nil {
		return nil, err
	}

	//printer.Printf("------------------\n")
	//printer.Printf("DEBUG: GETDOMAINS bodystring  = %s\n", bodyString)
	//printer.Printf("------------------\n")

	var dr domainsResult
	json.Unmarshal(bodyString, &dr)

	if dr.Meta.Pages > 1 {
		return nil, fmt.Errorf("cscglobal getDomains: unimplemented  paganation")
	}

	var r []string
	for _, d := range dr.Domains {
		r = append(r, d.QualifiedDomainName)
	}

	//printer.Printf("------------------\n")
	//printer.Printf("DEBUG: GETDOMAINS dr = %+v\n", dr)
	//printer.Printf("------------------\n")

	return r, nil
}

func (client *providerClient) getZoneRecordsAll(zone string) (*zoneResponse, error) {
	var bodyString, err = client.get("/zones/" + zone)
	if err != nil {
		return nil, err
	}

	if cscDebug {
		printer.Printf("------------------\n")
		printer.Printf("DEBUG: ZONE RESPONSE = %s\n", bodyString)
		printer.Printf("------------------\n")
	}

	var dr zoneResponse
	json.Unmarshal(bodyString, &dr)

	return &dr, nil
}

func (client *providerClient) sendZoneEditRequest(domainname string, edits []zoneResourceRecordEdit) error {

	req := zoneEditRequest{
		ZoneName: domainname,
		Edits:    &edits,
	}

	requestBody, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if cscDebug {
		printer.Printf("DEBUG: edit request = %s\n", requestBody)
	}
	responseBody, err := client.post("/zones/edits", requestBody)
	if err != nil {
		return err
	}

	// What did we get back?
	var errResp zoneEditRequestResultZoneEditRequestResult
	err = json.Unmarshal(responseBody, &errResp)
	if err != nil {
		return fmt.Errorf("CSC Global API error: %s DATA: %q", err, errResp)
	}
	if errResp.Content.Status != "SUCCESS" {
		return fmt.Errorf("CSC Global API error: %s DATA: %q", errResp.Content.Status, errResp.Content.Message)
	}

	// NB(tlim): The request was successfully submitted. The "statusURL" is what we query if we want to wait until the request was processed and see if it was a success.
	// Right now we don't want to wait. Instead, the next time we do a mutation we wait then.
	// If we ever change our mind and want to wait for success/failure, it would look like:
	//statusURL := errResp.Links.Status
	//return client.waitRequestURL(statusURL)

	return nil
}

func (client *providerClient) waitRequest(reqID string) error {
	return client.waitRequestURL(apiBase + "/zones/edits/status/" + reqID)
}

func (client *providerClient) waitRequestURL(statusURL string) error {
	t1 := time.Now()
	for {
		statusBody, err := client.geturl(statusURL)
		if err != nil {
			return fmt.Errorf("CSC Global API error: %s DATA: %q", err, statusBody)
		}
		var statusResp zoneEditStatusResultZoneEditStatusResult
		err = json.Unmarshal(statusBody, &statusResp)
		if err != nil {
			return fmt.Errorf("CSC Global API error: %s DATA: %q", err, statusBody)
		}
		status, msg := statusResp.Content.Status, statusResp.Content.ErrorDescription

		if isatty.IsTerminal(os.Stdout.Fd()) {
			dur := time.Since(t1).Round(time.Second)
			if msg == "" {
				printer.Printf("WAITING: % 6s STATUS=%s           \r", dur, status)
			} else {
				printer.Printf("WAITING: % 6s STATUS=%s MSG=%q    \r", dur, status, msg)
			}
		}
		if status == "FAILED" {
			parts := strings.Split(statusResp.Links.Cancel, "/")
			client.cancelRequest(parts[len(parts)-1])
			return fmt.Errorf("update failed: %s %s", msg, statusURL)
		}
		if status == "COMPLETED" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil

	// Response looks like:
	//{
	//	"content": {
	//	  "status": "SUCCESS",
	//	  "message": "The publish request was successfully enqueued."
	//	},
	//	"links": {
	//	  "self": "https://apis.cscglobal.com/dbs/api/v2/zones/edits/9e139e34-a2a1-462e-88ab-3645833a55d4",
	//	  "status": "https://apis.cscglobal.com/dbs/api/v2/zones/edits/status/9e139e34-a2a1-462e-88ab-3645833a55d4"
	//	}
	//  }
}

// Cancel pending/stuck edits

type pagedZoneEditResponsePagedZoneEditResponse struct {
	Meta struct {
		NumResults int `json:"numResults"`
		Pages      int `json:"pages"`
	} `json:"meta"`
	ZoneEdits []struct {
		ZoneName string `json:"zoneName"`
		ID       string `json:"id"`
		Status   string `json:"status"`
	} `json:"zoneEdits"`
}

func (client *providerClient) clearRequests(domain string) error {
	var bodyString, err = client.get("/zones/edits?filter=zoneName==" + domain)
	if err != nil {
		return err
	}

	var dr pagedZoneEditResponsePagedZoneEditResponse
	json.Unmarshal(bodyString, &dr)

	// TODO(tlim): Properly handle paganation.
	if dr.Meta.Pages != 1 {
		return fmt.Errorf("cancelPendingEdits failed: Pages=%d", dr.Meta.Pages)
	}

	for i, ze := range dr.ZoneEdits {
		if cscDebug {
			if ze.Status != "COMPLETED" && ze.Status != "CANCELED" {
				printer.Printf("REQUEST %d: %s %s\n", i, ze.ID, ze.Status)
			}
		}
		switch ze.Status {
		case "PROPAGATING":
			printer.Printf("INFO: Waiting for id=%s status=%s\n", ze.ID, ze.Status)
			client.waitRequest(ze.ID)
		case "FAILED":
			printer.Printf("INFO: Deleting request status=%s id=%s\n", ze.Status, ze.ID)
			client.cancelRequest(ze.ID)
		case "COMPLETED", "CANCELED":
			continue
		default:
			return fmt.Errorf("cscglobal ClearRequests: unimplemented status: %q", ze.Status)
		}

	}

	return nil
}

func (client *providerClient) cancelRequest(reqID string) error {
	_, err := client.delete("/zones/edits/" + reqID)
	return err
}

func (client *providerClient) put(endpoint string, requestBody []byte) ([]byte, error) {
	hclient := &http.Client{}
	req, _ := http.NewRequest("PUT", apiBase+endpoint, bytes.NewReader(requestBody))

	// Add headers
	req.Header.Add("apikey", client.key)
	req.Header.Add("Authorization", "Bearer "+client.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return bodyString, nil
	}

	// Got a error response from API, see if it's json format
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		// Some error messages are plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}

func (client *providerClient) delete(endpoint string) ([]byte, error) {
	hclient := &http.Client{}
	printer.Printf("DEBUG: delete endpoint: %q\n", apiBase+endpoint)
	req, _ := http.NewRequest("DELETE", apiBase+endpoint, nil)

	// Add headers
	req.Header.Add("apikey", client.key)
	req.Header.Add("Authorization", "Bearer "+client.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		printer.Printf("DEBUG: Delete successful (200)\n")
		return bodyString, nil
	}
	printer.Printf("DEBUG: Delete failed (%d)\n", resp.StatusCode)

	// Got a error response from API, see if it's json format
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		// Some error messages are plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}

func (client *providerClient) post(endpoint string, requestBody []byte) ([]byte, error) {
	hclient := &http.Client{}
	req, _ := http.NewRequest("POST", apiBase+endpoint, bytes.NewBuffer(requestBody))

	// Add headers
	req.Header.Add("apikey", client.key)
	req.Header.Add("Authorization", "Bearer "+client.token)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	//printer.Printf("------------------\n")
	//printer.Printf("DEBUG: resp.StatusCode == %d\n", resp.StatusCode)
	//printer.Printf("POST RESPONSE = %s\n", bodyString)
	//printer.Printf("------------------\n")
	if resp.StatusCode == 201 {
		return bodyString, nil
	}

	// Got a error response from API, see if it's json format
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		// Some error messages are plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}

func (client *providerClient) get(endpoint string) ([]byte, error) {
	return client.geturl(apiBase + endpoint)
}

func (client *providerClient) geturl(url string) ([]byte, error) {
	hclient := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)

	// Add headers
	req.Header.Add("apikey", client.key)
	req.Header.Add("Authorization", "Bearer "+client.token)
	req.Header.Add("Accept", "application/json")

	resp, err := hclient.Do(req)
	if err != nil {
		return nil, err
	}

	bodyString, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		return bodyString, nil
	}

	if resp.StatusCode == 400 {
		// 400, error message is in the body as plain text
		return nil, fmt.Errorf("CSC Global API error: %s URL: %s%s",
			bodyString,
			req.Host, req.URL.RequestURI())
	}

	// Got a json error response from API
	var errResp errorResponse
	err = json.Unmarshal(bodyString, &errResp)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("CSC Global API error code: %s description: %s URL: %s%s",
		errResp.Code, errResp.Description,
		req.Host, req.URL.RequestURI())
}
