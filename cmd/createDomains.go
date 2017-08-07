package cmd

import (
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/providers"
	"github.com/urfave/cli"
)

var createDomainsCommand = &cli.Command{
	Name:  "create-domains",
	Usage: "ensures that all domains in your configuration are present in all providers.",
	Action: func(ctx *cli.Context) error {
		return exit(CreateDomains(globalCreateDomainsArgs))
	},
	Category: catMain,
	Flags:    globalPreviewArgs.flags(),
}

type CreateDomainsArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
}

func (args *CreateDomainsArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

var globalCreateDomainsArgs CreateDomainsArgs

func CreateDomains(args CreateDomainsArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	registrars, dnsProviders, _, err := InitializeProviders(args.CredsFile, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Initialized %d registrars and %d dns service providers.\n", len(registrars), len(dnsProviders))
	for _, domain := range cfg.Domains {
		fmt.Println("*** ", domain.Name)
		for prov := range domain.DNSProviders {
			dsp, ok := dnsProviders[prov]
			if !ok {
				log.Fatalf("DSP %s not declared.", prov)
			}
			if creator, ok := dsp.(providers.DomainCreator); ok {
				fmt.Println("  -", prov)
				// TODO: maybe return bool if it did anything.
				err := creator.EnsureDomainExists(domain.Name)
				if err != nil {
					fmt.Printf("Error creating domain: %s\n", err)
				}
			}
		}
	}
	return nil
}
