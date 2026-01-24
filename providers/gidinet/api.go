package gidinet

import (
	"bytes"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	// DNS API endpoint
	dnsAPIURL     = "https://api.quickservicebox.com/API/Beta/DNSAPI.asmx"
	dnsSoapAction = "https://api.quickservicebox.com/DNS/DNSAPI/"

	// Core API endpoint (for domain listing)
	coreAPIURL     = "https://api.quickservicebox.com/API/Beta/CoreAPI.asmx"
	coreSoapAction = "http://api.quickservicebox.com/API/Beta/CoreAPI/"
)

// gidinetProvider holds the API credentials and HTTP client
type gidinetProvider struct {
	username    string
	passwordB64 string
	client      *http.Client
}

// newClient creates a new Gidinet API client
func newClient(username, password string) *gidinetProvider {
	return &gidinetProvider{
		username:    username,
		passwordB64: base64.StdEncoding.EncodeToString([]byte(password)),
		client:      &http.Client{},
	}
}

// buildSOAPRequest creates a SOAP envelope with the given body content
func buildSOAPRequest(body interface{}) ([]byte, error) {
	// Build XML manually to handle namespaces properly
	bodyXML, err := xml.MarshalIndent(body, "    ", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SOAP body: %w", err)
	}

	soapEnvelope := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<soap:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/">
  <soap:Body>
%s
  </soap:Body>
</soap:Envelope>`, string(bodyXML))

	return []byte(soapEnvelope), nil
}

// doSOAPRequest sends a SOAP request to the DNS API and returns the response body
func (c *gidinetProvider) doSOAPRequest(action string, requestBody interface{}) ([]byte, error) {
	return c.doSOAPRequestToURL(dnsAPIURL, dnsSoapAction+action, requestBody)
}

// doCoreAPIRequest sends a SOAP request to the Core API and returns the response body
func (c *gidinetProvider) doCoreAPIRequest(action string, requestBody interface{}) ([]byte, error) {
	return c.doSOAPRequestToURL(coreAPIURL, coreSoapAction+action, requestBody)
}

// doSOAPRequestToURL sends a SOAP request to the specified URL and returns the response body
func (c *gidinetProvider) doSOAPRequestToURL(url, soapAction string, requestBody interface{}) ([]byte, error) {
	soapData, err := buildSOAPRequest(requestBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(soapData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", soapAction)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// parseSOAPResponse extracts the response from a SOAP envelope
func parseSOAPResponse(data []byte, response interface{}) error {
	// Find the Body content and unmarshal it
	// We need to handle the SOAP envelope wrapper
	type envelope struct {
		Body struct {
			InnerXML []byte `xml:",innerxml"`
		} `xml:"Body"`
	}

	var env envelope
	if err := xml.Unmarshal(data, &env); err != nil {
		return fmt.Errorf("failed to parse SOAP envelope: %w", err)
	}

	if err := xml.Unmarshal(env.Body.InnerXML, response); err != nil {
		return fmt.Errorf("failed to parse SOAP response body: %w", err)
	}

	return nil
}

// recordGetList fetches all DNS records for a domain
func (c *gidinetProvider) recordGetList(domainName string) ([]*DNSRecordListItem, error) {
	request := &RecordGetListRequest{
		AccountUsername:    c.username,
		AccountPasswordB64: c.passwordB64,
		DomainName:         domainName,
	}

	responseData, err := c.doSOAPRequest("recordGetList", request)
	if err != nil {
		return nil, err
	}

	var response RecordGetListResponse
	if err := parseSOAPResponse(responseData, &response); err != nil {
		return nil, err
	}

	if response.ResultCode != ResultCodeSuccess {
		return nil, fmt.Errorf("gidinet API error (code %d): %s", response.ResultCode, response.ResultText)
	}

	return response.ResultItems, nil
}

// recordAdd creates a new DNS record
func (c *gidinetProvider) recordAdd(record *DNSRecord) error {
	request := &RecordAddRequest{
		AccountUsername:    c.username,
		AccountPasswordB64: c.passwordB64,
		Record:             record,
	}

	responseData, err := c.doSOAPRequest("recordAdd", request)
	if err != nil {
		return err
	}

	var response RecordAddResponse
	if err := parseSOAPResponse(responseData, &response); err != nil {
		return err
	}

	if response.ResultCode != ResultCodeSuccess {
		return fmt.Errorf("gidinet API recordAdd error (code %d): %s", response.ResultCode, response.ResultText)
	}

	return nil
}

// recordUpdate modifies an existing DNS record
func (c *gidinetProvider) recordUpdate(oldRecord, newRecord *DNSRecord) error {
	request := &RecordUpdateRequest{
		AccountUsername:    c.username,
		AccountPasswordB64: c.passwordB64,
		OldRecord:          oldRecord,
		NewRecord:          newRecord,
	}

	responseData, err := c.doSOAPRequest("recordUpdate", request)
	if err != nil {
		return err
	}

	var response RecordUpdateResponse
	if err := parseSOAPResponse(responseData, &response); err != nil {
		return err
	}

	if response.ResultCode != ResultCodeSuccess {
		return fmt.Errorf("gidinet API recordUpdate error (code %d): %s", response.ResultCode, response.ResultText)
	}

	return nil
}

// recordDelete removes a DNS record
func (c *gidinetProvider) recordDelete(record *DNSRecord) error {
	request := &RecordDeleteRequest{
		AccountUsername:    c.username,
		AccountPasswordB64: c.passwordB64,
		Record:             record,
	}

	responseData, err := c.doSOAPRequest("recordDelete", request)
	if err != nil {
		return err
	}

	var response RecordDeleteResponse
	if err := parseSOAPResponse(responseData, &response); err != nil {
		return err
	}

	if response.ResultCode != ResultCodeSuccess {
		return fmt.Errorf("gidinet API recordDelete error (code %d): %s", response.ResultCode, response.ResultText)
	}

	return nil
}

// domainGetList fetches all domains from the account
func (c *gidinetProvider) domainGetList() ([]*DomainListItem, error) {
	var allDomains []*DomainListItem
	pageNumber := 1
	pageSize := 200 // Maximum allowed

	for {
		request := &DomainGetListRequest{
			AccountUsername:     c.username,
			AccountPasswordB64:  c.passwordB64,
			OrderFieldID:        0, // Order by name
			OrderMode:           0, // Ascending
			PageSize:            pageSize,
			PageNumber:          pageNumber,
			GroupFilter:         0,  // All domains
			DomainFilter:        "", // No filter
			NameserversFilter:   "", // No filter
			RegistrantContactID: 0,  // No filter
			TechContactID:       0,  // No filter
		}

		responseData, err := c.doCoreAPIRequest("domainGetList", request)
		if err != nil {
			return nil, err
		}

		var response DomainGetListResponse
		if err := parseSOAPResponse(responseData, &response); err != nil {
			return nil, err
		}

		if response.ResultCode != ResultCodeSuccess {
			return nil, fmt.Errorf("gidinet API domainGetList error (code %d): %s", response.ResultCode, response.ResultText)
		}

		// Only include active domains with DNS service
		for _, domain := range response.ResultItems {
			if domain.StatusCode == DomainStatusActive {
				allDomains = append(allDomains, domain)
			}
		}

		// Check if we need to fetch more pages
		if pageNumber >= response.TotalPages {
			break
		}
		pageNumber++
	}

	return allDomains, nil
}

// fixTTL snaps a TTL value to the nearest allowed value
func fixTTL(ttl uint32) uint32 {
	// If TTL is larger than the largest allowed value, return the largest
	if ttl > AllowedTTLValues[len(AllowedTTLValues)-1] {
		return AllowedTTLValues[len(AllowedTTLValues)-1]
	}

	// Find the smallest allowed value that is >= ttl
	for _, v := range AllowedTTLValues {
		if v >= ttl {
			return v
		}
	}

	return AllowedTTLValues[0]
}

// toFQDN converts a hostname to FQDN if needed
func toFQDN(hostname, domain string) string {
	if hostname == "@" || hostname == "" {
		return domain
	}
	if strings.HasSuffix(hostname, "."+domain) {
		return hostname
	}
	if strings.HasSuffix(hostname, ".") {
		return strings.TrimSuffix(hostname, ".")
	}
	return hostname + "." + domain
}

// fromFQDN converts a FQDN to a relative hostname
func fromFQDN(fqdn, domain string) string {
	if fqdn == domain || fqdn == domain+"." {
		return "@"
	}
	suffix := "." + domain
	if strings.HasSuffix(fqdn, suffix) {
		return strings.TrimSuffix(fqdn, suffix)
	}
	return fqdn
}

// --- Registrar API methods ---

// getNameserversForDomain fetches the current nameservers for a specific domain
func (c *gidinetProvider) getNameserversForDomain(domainName string) ([]string, error) {
	// Fetch all domains and find the one we're looking for
	allDomains, err := c.domainGetList()
	if err != nil {
		return nil, err
	}

	for _, domain := range allDomains {
		// API returns DomainName and DomainExtension separately (e.g., "example" + "com")
		fullName := domain.DomainName + "." + domain.DomainExtension
		if strings.EqualFold(fullName, domainName) {
			if domain.Nameservers == "" {
				return nil, nil
			}
			// API returns nameservers separated by semicolons
			nsList := strings.Split(domain.Nameservers, ";")
			for i, ns := range nsList {
				nsList[i] = strings.TrimSpace(ns)
			}
			return nsList, nil
		}
	}

	return nil, fmt.Errorf("domain %s not found in account", domainName)
}

// setNameservers updates the nameservers for a domain at the registrar level
func (c *gidinetProvider) setNameservers(domainName string, nameservers []string) error {
	// Join nameservers with comma, no spaces
	nsString := strings.Join(nameservers, ",")

	request := &DomainNameServersChangeRequest{
		AccountUsername:      c.username,
		AccountPasswordB64:   c.passwordB64,
		Domain:               domainName,
		Nameservers:          nsString,
		AdditionalParameters: nil,
	}

	responseData, err := c.doCoreAPIRequest("domainNameServersChange", request)
	if err != nil {
		return err
	}

	var response DomainNameServersChangeResponse
	if err := parseSOAPResponse(responseData, &response); err != nil {
		return err
	}

	if response.ResultCode != ResultCodeSuccess {
		return fmt.Errorf("gidinet API domainNameServersChange error (code %d): %s", response.ResultCode, response.ResultText)
	}

	// Check operation result
	for _, item := range response.ResultItems {
		if item.ExitCode == 2 { // Failed
			return fmt.Errorf("gidinet nameserver change failed: %s", item.AdditionalDetails)
		}
	}

	return nil
}
