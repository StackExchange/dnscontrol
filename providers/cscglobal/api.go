package cscglobal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
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

// DomainsResult is the JSON returned by "/domains".  Fields we don't
// use are commented out.
type DomainsResult struct {
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

	//fmt.Printf("------------------\n")
	//fmt.Printf("BODYSTRING = %s\n", bodyString)
	//fmt.Printf("------------------\n")

	var dr DomainsResult
	json.Unmarshal(bodyString, &dr)

	if dr.Meta.Pages > 1 {
		return nil, fmt.Errorf("cscglobal getDomains: unimplemented  paganation")
	}

	var r []string
	for _, d := range dr.Domains {
		r = append(r, d.QualifiedDomainName)
	}

	//fmt.Printf("------------------\n")
	//fmt.Printf("DR = %+v\n", dr)
	//fmt.Printf("------------------\n")

	return r, nil
}

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
	ID       string `json:"id"`
	Key      string `json:"key"`
	Value    string `json:"value"`
	TTL      uint32 `json:"ttl"`
	Status   string `json:"status"`
	Priority int    `json:"priority"`
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
	Soa         []nativeRecordSOA   `json:"soa"`
}

func (client *providerClient) getZoneRecordsAll(zone string) (*zoneResponse, error) {
	var bodyString, err = client.get("/zones/" + zone)
	if err != nil {
		return nil, err
	}

	var dr zoneResponse
	json.Unmarshal(bodyString, &dr)

	return &dr, nil
}

type ZoneResourceRecordEdit = struct {
	Action       string `json:"action"`
	RecordType   string `json:"recordType"`
	CurrentKey   string `json:"currentKey"`
	CurrentValue string `json:"currentValue"`
	NewTTL       string `json:"newTtl,omitempty"`
	NewPriority  uint16 `json:"newPriority,omitempty"`
	NewWeight    uint16 `json:"newWeight,omitempty"`
	NewPort      uint16 `json:"newPort,omitempty"`
}

type ZoneEditRequest = struct {
	ZoneName string `json:"zoneName"`
	Edits    []ZoneResourceRecordEdit
}

//type ZoneEditRequestResult = struct {
//	ZoneName string `json:"zoneName"`
//	Edits    []ZoneResourceRecordEdit
//}

func (client *providerClient) SendZoneEditRequest(domainname string, edits ZoneResourceRecordEdit) (*zoneResponse, error) {

	data := ZoneEditRequest{
		ZoneName: domainname,
		Edits:    edits,
	}
	return client.post("/zones/edits", data)
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

func (client *providerClient) post(endpoint string, requestBody []byte) error {
	hclient := &http.Client{}
	req, _ := http.NewRequest("POST", apiBase+endpoint, bytes.NewReader(requestBody))

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

func (client *providerClient) get(endpoint string) ([]byte, error) {
	hclient := &http.Client{}
	req, _ := http.NewRequest("GET", apiBase+endpoint, nil)

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
