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
	Type string `json:"type"`
	// Normally MXPriority would be uint16 type, but since vercelClient.DNSRecord uses int64, we'd better be consistent here
	// Later in GetZoneRecords we do a `uint16OrZero` to ensure the type is correct
	MXPriority int64 `json:"mxPriority"`
}

// pagination represents the pagination object in Vercel API responses.
// From the Vercel API docs and the actual JSON response, the cursor appears to be some kind of timestamps
// But since pagination is not implemented in the Vercel official Go SDK, and every number used by Vercel's
// official SDK (TTL, Priority) is int64, I just use int64 as well here, just to be safe.
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

// Vercel API limit is max 100
const vercelApiPaginationLimit = 100

// listDNSRecords retrieves all DNS records for a domain, handling pagination.
// It replaces the client.ListDNSRecords method which does not support pagination.
// The official Vercel client's ListDNSRecords is a test helper that limits results to 100
// and does not implement the pagination logic required for production use with large zones.
func (c *vercelProvider) listDNSRecords(domain string) ([]domainRecord, error) {
	var allRecords []domainRecord
	var nextTimestamp int64

	for {
		url := fmt.Sprintf("https://api.vercel.com/v4/domains/%s/records?limit=%d", domain, vercelApiPaginationLimit)
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
			if r.TeamID == "" {
				r.TeamID = c.teamID
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
