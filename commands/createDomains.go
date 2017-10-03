package commands

import (
	"fmt"
	"log"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/urfave/cli"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args CreateDomainsArgs
	return &cli.Command{
		Name:  "create-domains",
		Usage: "ensures that all domains in your configuration are present in all providers.",
		Action: func(ctx *cli.Context) error {
			return exit(CreateDomains(args))
		},
		Flags: args.flags(),
	}
}())

type CreateDomainsArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
}

func (args *CreateDomainsArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

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
		for prov := range domain.DNSProviderNames {
			dsp, ok := dnsProviders[prov]
			if !ok {
				log.Fatalf("DSP %s not declared.", prov)
			}
			if creator, ok := dsp.(models.DomainCreator); ok {
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
