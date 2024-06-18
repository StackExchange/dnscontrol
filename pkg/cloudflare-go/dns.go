package cloudflare

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"golang.org/x/net/idna"
)

// ErrMissingBINDContents is for when the BIND file contents is required but not set.
var ErrMissingBINDContents = errors.New("required BIND config contents missing")

// DNSRecord represents a DNS record in a zone.
type DNSRecord struct {
	CreatedOn  time.Time   `json:"created_on,omitempty"`
	ModifiedOn time.Time   `json:"modified_on,omitempty"`
	Type       string      `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	Content    string      `json:"content,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	Data       interface{} `json:"data,omitempty"` // data returned by: SRV, LOC
	ID         string      `json:"id,omitempty"`
	ZoneID     string      `json:"zone_id,omitempty"`
	ZoneName   string      `json:"zone_name,omitempty"`
	Priority   *uint16     `json:"priority,omitempty"`
	TTL        int         `json:"ttl,omitempty"`
	Proxied    *bool       `json:"proxied,omitempty"`
	Proxiable  bool        `json:"proxiable,omitempty"`
	Comment    string      `json:"comment,omitempty"` // the server will omit the comment field when the comment is empty
	Tags       []string    `json:"tags,omitempty"`
}

// DNSRecordResponse represents the response from the DNS endpoint.
type DNSRecordResponse struct {
	Result DNSRecord `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

type ListDirection string

const (
	ListDirectionAsc  ListDirection = "asc"
	ListDirectionDesc ListDirection = "desc"
)

type ListDNSRecordsParams struct {
	Type      string        `url:"type,omitempty"`
	Name      string        `url:"name,omitempty"`
	Content   string        `url:"content,omitempty"`
	Proxied   *bool         `url:"proxied,omitempty"`
	Comment   string        `url:"comment,omitempty"` // currently, the server does not support searching for records with an empty comment
	Tags      []string      `url:"tag,omitempty"`     // potentially multiple `tag=`
	TagMatch  string        `url:"tag-match,omitempty"`
	Order     string        `url:"order,omitempty"`
	Direction ListDirection `url:"direction,omitempty"`
	Match     string        `url:"match,omitempty"`
	Priority  *uint16       `url:"-"`

	ResultInfo
}

type UpdateDNSRecordParams struct {
	Type     string      `json:"type,omitempty"`
	Name     string      `json:"name,omitempty"`
	Content  string      `json:"content,omitempty"`
	Data     interface{} `json:"data,omitempty"` // data for: SRV, LOC
	ID       string      `json:"-"`
	Priority *uint16     `json:"priority,omitempty"`
	TTL      int         `json:"ttl,omitempty"`
	Proxied  *bool       `json:"proxied,omitempty"`
	Comment  *string     `json:"comment,omitempty"` // nil will keep the current comment, while StringPtr("") will empty it
	Tags     []string    `json:"tags"`
}

// DNSListResponse represents the response from the list DNS records endpoint.
type DNSListResponse struct {
	Result []DNSRecord `json:"result"`
	Response
	ResultInfo `json:"result_info"`
}

// listDNSRecordsDefaultPageSize represents the default per_page size of the API.
var listDNSRecordsDefaultPageSize int = 100

// nontransitionalLookup implements the nontransitional processing as specified in
// Unicode Technical Standard 46 with almost all checkings off to maximize user freedom.
var nontransitionalLookup = idna.New(
	idna.MapForLookup(),
	idna.StrictDomainName(false),
	idna.ValidateLabels(false),
)

// toUTS46ASCII tries to convert IDNs (international domain names)
// from Unicode form to Punycode, using non-transitional process specified
// in UTS 46.
//
// Note: conversion errors are silently discarded and partial conversion
// results are used.
func toUTS46ASCII(name string) string {
	name, _ = nontransitionalLookup.ToASCII(name)
	return name
}

// proxiedRecordsRe is the regular expression for determining if a DNS record
// is proxied or not.
var proxiedRecordsRe = regexp.MustCompile(`(?m)^.*\.\s+1\s+IN\s+CNAME.*$`)

// proxiedRecordImportTemplate is the multipart template for importing *only*
// proxied records. See `nonProxiedRecordImportTemplate` for importing records
// that are not proxied.
var proxiedRecordImportTemplate = `--------------------------BOUNDARY
Content-Disposition: form-data; name="file"; filename="bind.txt"

%s
--------------------------BOUNDARY
Content-Disposition: form-data; name="proxied"

true
--------------------------BOUNDARY--`

// nonProxiedRecordImportTemplate is the multipart template for importing DNS
// records that are not proxed. For importing proxied records, use
// `proxiedRecordImportTemplate`.
var nonProxiedRecordImportTemplate = `--------------------------BOUNDARY
Content-Disposition: form-data; name="file"; filename="bind.txt"

%s
--------------------------BOUNDARY--`

// sanitiseBINDFileInput accepts the BIND file as a string and removes parts
// that are not required for importing or would break the import (like SOA
// records).
func sanitiseBINDFileInput(s string) string {
	// Remove SOA records.
	soaRe := regexp.MustCompile(`(?m)[\r\n]+^.*IN\s+SOA.*$`)
	s = soaRe.ReplaceAllString(s, "")

	// Remove all comments.
	commentRe := regexp.MustCompile(`(?m)[\r\n]+^.*;;.*$`)
	s = commentRe.ReplaceAllString(s, "")

	// Swap all the tabs to spaces.
	r := strings.NewReplacer(
		"\t", " ",
		"\n\n", "\n",
	)
	s = r.Replace(s)
	s = strings.TrimSpace(s)

	return s
}

// extractProxiedRecords accepts a BIND file (as a string) and returns only the
// proxied DNS records.
func extractProxiedRecords(s string) string {
	proxiedOnlyRecords := proxiedRecordsRe.FindAllString(s, -1)
	return strings.Join(proxiedOnlyRecords, "\n")
}

// removeProxiedRecords accepts a BIND file (as a string) and returns the file
// contents without any proxied records included.
func removeProxiedRecords(s string) string {
	return proxiedRecordsRe.ReplaceAllString(s, "")
}

type ExportDNSRecordsParams struct{}
type ImportDNSRecordsParams struct {
	BINDContents string
}

type CreateDNSRecordParams struct {
	CreatedOn  time.Time   `json:"created_on,omitempty" url:"created_on,omitempty"`
	ModifiedOn time.Time   `json:"modified_on,omitempty" url:"modified_on,omitempty"`
	Type       string      `json:"type,omitempty" url:"type,omitempty"`
	Name       string      `json:"name,omitempty" url:"name,omitempty"`
	Content    string      `json:"content,omitempty" url:"content,omitempty"`
	Meta       interface{} `json:"meta,omitempty"`
	Data       interface{} `json:"data,omitempty"` // data returned by: SRV, LOC
	ID         string      `json:"id,omitempty"`
	ZoneID     string      `json:"zone_id,omitempty"`
	ZoneName   string      `json:"zone_name,omitempty"`
	Priority   *uint16     `json:"priority,omitempty"`
	TTL        int         `json:"ttl,omitempty"`
	Proxied    *bool       `json:"proxied,omitempty" url:"proxied,omitempty"`
	Proxiable  bool        `json:"proxiable,omitempty"`
	Comment    string      `json:"comment,omitempty" url:"comment,omitempty"` // to the server, there's no difference between "no comment" and "empty comment"
	Tags       []string    `json:"tags,omitempty"`
}

// CreateDNSRecord creates a DNS record for the zone identifier.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-create-dns-record
func (api *API) CreateDNSRecord(ctx context.Context, rc *ResourceContainer, params CreateDNSRecordParams) (DNSRecord, error) {
	if rc.Identifier == "" {
		return DNSRecord{}, ErrMissingZoneID
	}
	params.Name = toUTS46ASCII(params.Name)

	uri := fmt.Sprintf("/zones/%s/dns_records", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodPost, uri, params)
	if err != nil {
		return DNSRecord{}, err
	}

	var recordResp *DNSRecordResponse
	err = json.Unmarshal(res, &recordResp)
	if err != nil {
		return DNSRecord{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return recordResp.Result, nil
}

// ListDNSRecords returns a slice of DNS records for the given zone identifier.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-list-dns-records
func (api *API) ListDNSRecords(ctx context.Context, rc *ResourceContainer, params ListDNSRecordsParams) ([]DNSRecord, *ResultInfo, error) {
	if rc.Identifier == "" {
		return nil, nil, ErrMissingZoneID
	}

	params.Name = toUTS46ASCII(params.Name)

	autoPaginate := true
	if params.PerPage >= 1 || params.Page >= 1 {
		autoPaginate = false
	}

	if params.PerPage < 1 {
		params.PerPage = listDNSRecordsDefaultPageSize
	}

	if params.Page < 1 {
		params.Page = 1
	}

	var records []DNSRecord
	var lastResultInfo ResultInfo

	for {
		uri := buildURI(fmt.Sprintf("/zones/%s/dns_records", rc.Identifier), params)
		res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
		if err != nil {
			return []DNSRecord{}, &ResultInfo{}, err
		}
		var listResponse DNSListResponse
		err = json.Unmarshal(res, &listResponse)
		if err != nil {
			return []DNSRecord{}, &ResultInfo{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
		}
		records = append(records, listResponse.Result...)
		lastResultInfo = listResponse.ResultInfo
		params.ResultInfo = listResponse.ResultInfo.Next()
		if params.ResultInfo.Done() || !autoPaginate {
			break
		}
	}
	return records, &lastResultInfo, nil
}

// ErrMissingDNSRecordID is for when DNS record ID is needed but not given.
var ErrMissingDNSRecordID = errors.New("required DNS record ID missing")

// GetDNSRecord returns a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-dns-record-details
func (api *API) GetDNSRecord(ctx context.Context, rc *ResourceContainer, recordID string) (DNSRecord, error) {
	if rc.Identifier == "" {
		return DNSRecord{}, ErrMissingZoneID
	}
	if recordID == "" {
		return DNSRecord{}, ErrMissingDNSRecordID
	}

	uri := fmt.Sprintf("/zones/%s/dns_records/%s", rc.Identifier, recordID)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return DNSRecord{}, err
	}
	var r DNSRecordResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return DNSRecord{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return r.Result, nil
}

// UpdateDNSRecord updates a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-update-dns-record
func (api *API) UpdateDNSRecord(ctx context.Context, rc *ResourceContainer, params UpdateDNSRecordParams) (DNSRecord, error) {
	if rc.Identifier == "" {
		return DNSRecord{}, ErrMissingZoneID
	}

	if params.ID == "" {
		return DNSRecord{}, ErrMissingDNSRecordID
	}

	params.Name = toUTS46ASCII(params.Name)

	uri := fmt.Sprintf("/zones/%s/dns_records/%s", rc.Identifier, params.ID)
	res, err := api.makeRequestContext(ctx, http.MethodPatch, uri, params)
	if err != nil {
		return DNSRecord{}, err
	}

	var recordResp *DNSRecordResponse
	err = json.Unmarshal(res, &recordResp)
	if err != nil {
		return DNSRecord{}, fmt.Errorf("%s: %w", errUnmarshalError, err)
	}

	return recordResp.Result, nil
}

// DeleteDNSRecord deletes a single DNS record for the given zone & record
// identifiers.
//
// API reference: https://api.cloudflare.com/#dns-records-for-a-zone-delete-dns-record
func (api *API) DeleteDNSRecord(ctx context.Context, rc *ResourceContainer, recordID string) error {
	if rc.Identifier == "" {
		return ErrMissingZoneID
	}
	if recordID == "" {
		return ErrMissingDNSRecordID
	}

	uri := fmt.Sprintf("/zones/%s/dns_records/%s", rc.Identifier, recordID)
	res, err := api.makeRequestContext(ctx, http.MethodDelete, uri, nil)
	if err != nil {
		return err
	}
	var r DNSRecordResponse
	err = json.Unmarshal(res, &r)
	if err != nil {
		return fmt.Errorf("%s: %w", errUnmarshalError, err)
	}
	return nil
}

// ExportDNSRecords returns all DNS records for a zone in the BIND format.
//
// API reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-export-dns-records
func (api *API) ExportDNSRecords(ctx context.Context, rc *ResourceContainer, params ExportDNSRecordsParams) (string, error) {
	if rc.Level != ZoneRouteLevel {
		return "", ErrRequiredZoneLevelResourceContainer
	}

	if rc.Identifier == "" {
		return "", ErrMissingZoneID
	}

	uri := fmt.Sprintf("/zones/%s/dns_records/export", rc.Identifier)
	res, err := api.makeRequestContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

// ImportDNSRecords takes the contents of a BIND configuration file and imports
// all records at once.
//
// The current state of the API doesn't allow the proxying field to be
// automatically set on records where the TTL is 1. Instead you need to
// explicitly tell the endpoint which records are proxied in the form data. To
// achieve a simpler abstraction, we do the legwork in the method of making the
// two separate API calls (one for proxied and one for non-proxied) instead of
// making the end user know about this detail.
//
// API reference: https://developers.cloudflare.com/api/operations/dns-records-for-a-zone-import-dns-records
func (api *API) ImportDNSRecords(ctx context.Context, rc *ResourceContainer, params ImportDNSRecordsParams) error {
	if rc.Level != ZoneRouteLevel {
		return ErrRequiredZoneLevelResourceContainer
	}

	if rc.Identifier == "" {
		return ErrMissingZoneID
	}

	if params.BINDContents == "" {
		return ErrMissingBINDContents
	}

	sanitisedBindData := sanitiseBINDFileInput(params.BINDContents)
	nonProxiedRecords := removeProxiedRecords(sanitisedBindData)
	proxiedOnlyRecords := extractProxiedRecords(sanitisedBindData)

	nonProxiedRecordPayload := []byte(fmt.Sprintf(nonProxiedRecordImportTemplate, nonProxiedRecords))
	nonProxiedReqBody := bytes.NewReader(nonProxiedRecordPayload)

	uri := fmt.Sprintf("/zones/%s/dns_records/import", rc.Identifier)
	multipartUploadHeaders := http.Header{
		"Content-Type": {"multipart/form-data; boundary=------------------------BOUNDARY"},
	}

	_, err := api.makeRequestContextWithHeaders(ctx, http.MethodPost, uri, nonProxiedReqBody, multipartUploadHeaders)
	if err != nil {
		return err
	}

	proxiedRecordPayload := []byte(fmt.Sprintf(proxiedRecordImportTemplate, proxiedOnlyRecords))
	proxiedReqBody := bytes.NewReader(proxiedRecordPayload)

	_, err = api.makeRequestContextWithHeaders(ctx, http.MethodPost, uri, proxiedReqBody, multipartUploadHeaders)
	if err != nil {
		return err
	}

	return nil
}
