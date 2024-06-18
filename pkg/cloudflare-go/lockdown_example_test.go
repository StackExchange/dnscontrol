package cloudflare_test

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func ExampleAPI_ListZoneLockdowns_all() {
	api, err := cloudflare.New("deadbeef", "test@example.org")
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName("example.com")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch all Zone Lockdown rules for a zone, by page.
	rules, _, err := api.ListZoneLockdowns(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.LockdownListParams{})
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range rules {
		fmt.Printf("%s: %s\n", strings.Join(r.URLs, ", "), r.Configurations)
	}
}

func ExampleAPI_CreateZoneLockdown() {
	api, err := cloudflare.New("deadbeef", "test@example.org")
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName("example.org")
	if err != nil {
		log.Fatal(err)
	}

	newZoneLockdown := cloudflare.ZoneLockdownCreateParams{
		Description: "Test Zone Lockdown Rule",
		URLs: []string{
			"*.example.org/test",
		},
		Configurations: []cloudflare.ZoneLockdownConfig{
			{
				Target: "ip",
				Value:  "198.51.100.1",
			},
		},
		Paused: false,
	}

	response, err := api.CreateZoneLockdown(context.Background(), cloudflare.ZoneIdentifier(zoneID), newZoneLockdown)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Response: ", response)
}
