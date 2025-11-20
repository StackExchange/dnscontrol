package vercel

import (
	"context"
	"fmt"

	vercelClient "github.com/vercel/terraform-provider-vercel/client"
)

// domainRecord is a helper struct to unmarshal the JSON response.
// It embeds vercelClient.DNSRecord to reuse the upstream type,
// but adds fields to handle API inconsistencies (type vs recordType, mxPriority).
type domainRecord struct {
	vercelClient.DNSRecord
	Type       string `json:"type"`
	MXPriority int64  `json:"mxPriority"`
}

// pagination represents the pagination object in Vercel API responses.
type pagination struct {
	Count int64  `json:"count"`
	Next  *int64 `json:"next"`
	Prev  *int64 `json:"prev"`
}

// listResponse represents the response from the Vercel List DNS Records API.
type listResponse struct {
	Records    []domainRecord `json:"records"`
	Pagination pagination     `json:"pagination"`
}

// listDNSRecords retrieves all DNS records for a domain, handling pagination.
// It replaces the client.ListDNSRecords method which does not support pagination.
// The official Vercel client's ListDNSRecords is a test helper that limits results to 100
// and does not implement the pagination logic required for production use with large zones.
func (c *vercelProvider) listDNSRecords(domain string) ([]domainRecord, error) {
	var allRecords []domainRecord
	var nextTimestamp int64

	// Vercel API limit is max 100
	limit := 100

	for {
		url := fmt.Sprintf("https://api.vercel.com/v4/domains/%s/records?limit=%d", domain, limit)
		if c.teamID != "" {
			url += fmt.Sprintf("&teamId=%s", c.teamID)
		}
		if nextTimestamp != 0 {
			url += fmt.Sprintf("&until=%d", nextTimestamp)
		}

		var result listResponse
		err := c.doRequest(clientRequest{
			ctx:    context.Background(),
			method: "GET",
			url:    url,
		}, &result)

		if err != nil {
			return nil, fmt.Errorf("failed to list DNS records: %w", err)
		}

		for _, r := range result.Records {
			// The official SDK expects 'recordType' but the API returns 'type'.
			// We explicitly map it here to fix the discrepancy.
			r.RecordType = r.Type
			// Ensure Domain field is set (it might not be in the record object itself)
			if r.Domain == "" {
				r.Domain = domain
			}

			allRecords = append(allRecords, r)
		}

		if result.Pagination.Next == nil {
			break
		}
		nextTimestamp = *result.Pagination.Next
	}

	return allRecords, nil
}
