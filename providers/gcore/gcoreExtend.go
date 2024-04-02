package gcore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	dnssdk "github.com/G-Core/gcore-dns-sdk-go"
)

type gcoreZone struct {
	DNSSECEnabled bool `json:"dnssec_enabled"`
}

type gcoreDNSSECRequest struct {
	Enabled bool `json:"enabled"`
}

type gcoreRRSets struct {
	RRSets []gcoreRRSetExtended `json:"rrsets"`
}

// Extended attributes over dnssdk.RRSet
type gcoreRRSetExtended struct {
	Name string `json:"name"`
	Type string `json:"type"`

	// Original
	TTL     int                     `json:"ttl"`
	Records []dnssdk.ResourceRecord `json:"resource_records"`
	Filters []dnssdk.RecordFilter   `json:"filters"`
}

func dnssdkDo(ctx context.Context, c *dnssdk.Client, apiKey string, method, uri string, bodyParams interface{}, dest interface{}) error {
	// Adapted from https://github.com/G-Core/gcore-dns-sdk-go/blob/main/client.go#L289
	// No way to reflect a private method in Golang

	var bs []byte
	if bodyParams != nil {
		var err error
		bs, err = json.Marshal(bodyParams)
		if err != nil {
			return fmt.Errorf("encode bodyParams: %w", err)
		}
	}

	endpoint, err := c.BaseURL.Parse(path.Join(c.BaseURL.Path, uri))
	if err != nil {
		return fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if c.Debug {
		log.Printf("[DEBUG] dns api request: %s %s %s \n", method, uri, bs)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint.String(), strings.NewReader(string(bs)))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("APIKey %s", apiKey))
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= http.StatusMultipleChoices {
		all, _ := io.ReadAll(resp.Body)
		e := dnssdk.APIError{
			StatusCode: resp.StatusCode,
		}
		err := json.Unmarshal(all, &e)
		if err != nil {
			e.Message = string(all)
		}
		return e
	}

	if dest == nil {
		return nil
	}

	// nolint: wrapcheck
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *gcoreProvider) dnssdkRRSets(domain string) (gcoreRRSets, error) {
	// Turns out G-Core has a hidden parameter "all=true"
	// https://github.com/octodns/octodns-gcore/blob/main/octodns_gcore/__init__.py#L105
	// But this isn't exposed with their API, need to manually call it

	var result gcoreRRSets
	url := fmt.Sprintf("/v2/zones/%s/rrsets?all=true", domain)

	err := dnssdkDo(c.ctx, c.provider, c.apiKey, http.MethodGet, url, nil, &result)
	if err != nil {
		return gcoreRRSets{}, err
	}

	return result, nil
}

func (c *gcoreProvider) dnssdkGetDNSSEC(domain string) (bool, error) {
	var result gcoreZone
	url := fmt.Sprintf("/v2/zones/%s", domain)

	err := dnssdkDo(c.ctx, c.provider, c.apiKey, http.MethodGet, url, nil, &result)
	if err != nil {
		return false, err
	}

	return result.DNSSECEnabled, nil
}

func (c *gcoreProvider) dnssdkSetDNSSEC(domain string, enabled bool) error {
	var request gcoreDNSSECRequest
	request.Enabled = enabled

	url := fmt.Sprintf("/v2/zones/%s/dnssec", domain)

	err := dnssdkDo(c.ctx, c.provider, c.apiKey, http.MethodPatch, url, request, nil)
	if err != nil {
		return err
	}

	return nil
}
