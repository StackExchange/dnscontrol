package cloudflare_test

import (
	"context"
	"fmt"
	"log"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func ExampleAPI_RegistrarDomain() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	domain, err := api.RegistrarDomain(context.Background(), "01a7362d577a6c3019a474fd6f485823", "cloudflare.com")
	if err != nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", domain)
}

func ExampleAPI_RegistrarDomains() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	domains, err := api.RegistrarDomains(context.Background(), "01a7362d577a6c3019a474fd6f485823")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", domains)
}

func ExampleAPI_TransferRegistrarDomain() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	domain, err := api.TransferRegistrarDomain(context.Background(), "01a7362d577a6c3019a474fd6f485823", "cloudflare.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", domain)
}

func ExampleAPI_CancelRegistrarDomainTransfer() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	domains, err := api.CancelRegistrarDomainTransfer(context.Background(), "01a7362d577a6c3019a474fd6f485823", "cloudflare.com")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", domains)
}

func ExampleAPI_UpdateRegistrarDomain() {
	api, err := cloudflare.New(apiKey, user)
	if err != nil {
		log.Fatal(err)
	}

	domain, err := api.UpdateRegistrarDomain(context.Background(), "01a7362d577a6c3019a474fd6f485823", "cloudflare.com", cloudflare.RegistrarDomainConfiguration{
		NameServers: []string{"ns1.cloudflare.com", "ns2.cloudflare.com"},
		Locked:      false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", domain)
}
