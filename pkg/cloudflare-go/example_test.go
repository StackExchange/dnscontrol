package cloudflare_test

import (
	"context"
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

const (
	user   = "cloudflare@example.org"
	domain = "example.com"
	apiKey = "deadbeef"
)

func Example() {
	api, err := cloudflare.New("deadbeef", "cloudflare@example.org")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fetch the zone ID for zone example.org
	zoneID, err := api.ZoneIDByName("example.org")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Fetch all DNS records for example.org
	records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, r := range records {
		fmt.Printf("%s: %s\n", r.Name, r.Content)
	}
}
