package commands

import (
	"fmt"

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
	if err = InitializeProviders(args.CredsFile, cfg); err != nil {
		return err
	}
	for _, domain := range cfg.Domains {
		fmt.Println("*** ", domain.Name)
		for _, provider := range domain.DNSProviders {
			if creator, ok := provider.(models.DomainCreator); ok {
				fmt.Println("  -", provider.Name())
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
