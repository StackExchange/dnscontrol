package cloudflare_test

import (
	context "context"
	"fmt"
	"log"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func ExampleAPI_ListLoadBalancers() {
	// Construct a new API object.
	api, err := cloudflare.New("deadbeef", "test@example.com")
	if err != nil {
		log.Fatal(err)
	}

	// List LBs configured in zone.
	lbList, err := api.ListLoadBalancers(context.Background(), cloudflare.ZoneIdentifier("d56084adb405e0b7e32c52321bf07be6"), cloudflare.ListLoadBalancerParams{})
	if err != nil {
		log.Fatal(err)
	}

	for _, lb := range lbList {
		fmt.Println(lb)
	}
}

func ExampleAPI_GetLoadBalancerPoolHealth() {
	// Construct a new API object.
	api, err := cloudflare.New("deadbeef", "test@example.com")
	if err != nil {
		log.Fatal(err)
	}

	// Fetch pool health details.
	healthInfo, err := api.GetLoadBalancerPoolHealth(context.Background(), cloudflare.AccountIdentifier("01a7362d577a6c3019a474fd6f485823"), "example-pool-id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(healthInfo)
}
