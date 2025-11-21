package vercel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/StackExchange/dnscontrol/v4/models"
	vercelClient "github.com/vercel/terraform-provider-vercel/client"
)

// HTTPS record structure for Vercel API
type httpsRecord struct {
	Priority uint16 `json:"priority"`
	Target   string `json:"target"`
	Params   string `json:"params,omitempty"`
}

type CreateOrUpdateHTTPSDNSRecordRequest struct {
	Name string `json:"name"`
	// Normally TTL would be uint32 type, but since vercelClient.DNSRecord uses int64, we'd better be consistent here
	TTL     int64        `json:"ttl,omitempty"`
	Type    string       `json:"type"`
	Value   string       `json:"value,omitempty"`
	Comment string       `json:"comment"`
	HTTPS   *httpsRecord `json:"https,omitempty"`
}

// createHTTPSRecord creates an HTTPS DNS record using Vercel's custom JSON structure.
// HTTPS records require a special "https" field instead of the standard fields.
func (c *vercelProvider) createHTTPSRecord(ctx context.Context, domain string, rc *models.RecordConfig) error {
	url := fmt.Sprintf("https://api.vercel.com/v4/domains/%s/records", domain)
	if c.teamID != "" {
		url += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	// Parse HTTPS record from RecordConfig
	httpsRec := &httpsRecord{
		Priority: rc.SvcPriority,
		Target:   rc.GetTargetField(),
		Params:   rc.SvcParams,
	}

	payload := CreateOrUpdateHTTPSDNSRecordRequest{
		Name:    rc.Name,
		TTL:     int64(rc.TTL),
		Type:    "HTTPS",
		Comment: "",
		HTTPS:   httpsRec,
	}

	var response struct {
		RecordID string `json:"uid"`
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return c.doRequest(clientRequest{
		ctx:    ctx,
		method: "POST",
		url:    url,
		body:   string(payloadJSON),
	}, &response)
}

// updateHTTPSRecord updates an HTTPS DNS record using Vercel's custom JSON structure.
func (c *vercelProvider) updateHTTPSRecord(ctx context.Context, recordID string, rc *models.RecordConfig) error {
	url := fmt.Sprintf("https://api.vercel.com/v4/domains/records/%s", recordID)
	if c.teamID != "" {
		url += fmt.Sprintf("?teamId=%s", c.teamID)
	}

	// Parse HTTPS record from RecordConfig
	httpsRec := &httpsRecord{
		Priority: rc.SvcPriority,
		Target:   rc.GetTargetField(),
		Params:   rc.SvcParams,
	}

	payload := CreateOrUpdateHTTPSDNSRecordRequest{
		Name:    rc.Name,
		TTL:     int64(rc.TTL),
		Type:    "HTTPS",
		Comment: "",
		HTTPS:   httpsRec,
	}

	var result vercelClient.DNSRecord
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.doRequest(clientRequest{
		ctx:    ctx,
		method: "PATCH",
		url:    url,
		body:   string(payloadJSON),
	}, &result)
}
