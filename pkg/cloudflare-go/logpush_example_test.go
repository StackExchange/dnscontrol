package cloudflare_test

import (
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-json"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func ExampleAPI_CreateLogpushJob() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	job, err := api.CreateLogpushJob(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.CreateLogpushJobParams{
		Enabled:         false,
		Name:            "example.com",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339",
		DestinationConf: "s3://mybucket/logs?region=us-west-2",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", job)
}

func ExampleAPI_UpdateLogpushJob() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	err = api.UpdateLogpushJob(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.UpdateLogpushJobParams{
		ID:              1,
		Enabled:         true,
		Name:            "updated.com",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp",
		DestinationConf: "gs://mybucket/logs",
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleAPI_ListLogpushJobs() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	jobs, err := api.ListLogpushJobs(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListLogpushJobsParams{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", jobs)
	for _, r := range jobs {
		fmt.Printf("%+v\n", r)
	}
}

func ExampleAPI_GetLogpushJob() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	job, err := api.GetLogpushJob(context.Background(), cloudflare.ZoneIdentifier(zoneID), 1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", job)
}

func ExampleAPI_DeleteLogpushJob() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	err = api.DeleteLogpushJob(context.Background(), cloudflare.ZoneIdentifier(zoneID), 1)
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleAPI_GetLogpushOwnershipChallenge() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	ownershipChallenge, err := api.GetLogpushOwnershipChallenge(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.GetLogpushOwnershipChallengeParams{DestinationConf: "destination_conf"})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", ownershipChallenge)
}

func ExampleAPI_ValidateLogpushOwnershipChallenge() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	isValid, err := api.ValidateLogpushOwnershipChallenge(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ValidateLogpushOwnershipChallengeParams{
		DestinationConf:    "destination_conf",
		OwnershipChallenge: "ownership_challenge",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", isValid)
}

func ExampleAPI_CheckLogpushDestinationExists() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatal(err)
	}

	exists, err := api.CheckLogpushDestinationExists(context.Background(), cloudflare.ZoneIdentifier(zoneID), "destination_conf")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", exists)
}

func ExampleLogpushJob_MarshalJSON() {
	job := cloudflare.LogpushJob{
		Name:            "example.com static assets",
		LogpullOptions:  "fields=RayID,ClientIP,EdgeStartTimestamp&timestamps=rfc3339&CVE-2021-44228=true",
		Dataset:         "http_requests",
		DestinationConf: "s3://<BUCKET_PATH>?region=us-west-2/",
		Filter: &cloudflare.LogpushJobFilters{
			Where: cloudflare.LogpushJobFilter{
				And: []cloudflare.LogpushJobFilter{
					{Key: "ClientRequestPath", Operator: cloudflare.Contains, Value: "/static\\"},
					{Key: "ClientRequestHost", Operator: cloudflare.Equal, Value: "example.com"},
				},
			},
		},
	}

	jobstring, err := json.Marshal(job)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", jobstring)
	// Output: {"filter":"{\"where\":{\"and\":[{\"key\":\"ClientRequestPath\",\"operator\":\"contains\",\"value\":\"/static\\\\\"},{\"key\":\"ClientRequestHost\",\"operator\":\"eq\",\"value\":\"example.com\"}]}}","dataset":"http_requests","enabled":false,"name":"example.com static assets","logpull_options":"fields=RayID,ClientIP,EdgeStartTimestamp\u0026timestamps=rfc3339\u0026CVE-2021-44228=true","destination_conf":"s3://\u003cBUCKET_PATH\u003e?region=us-west-2/"}
}
