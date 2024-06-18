package cloudflare_test

import (
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-json"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func ExampleAPI_AccessAuditLogs() {
	api, err := cloudflare.New("deadbeef", "test@example.org")
	if err != nil {
		log.Fatal(err)
	}

	filterOpts := cloudflare.AccessAuditLogFilterOptions{}
	results, _ := api.AccessAuditLogs(context.Background(), "someaccountid", filterOpts)

	for _, record := range results {
		b, _ := json.Marshal(record)
		fmt.Println(string(b))
	}
}
