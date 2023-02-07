package commands

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/urfave/cli/v2"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args CreateDomainsArgs
	return &cli.Command{
		Name:  "create-domains",
		Usage: "DEPRECATED: Ensures that all domains in your configuration are activated at their Domain Service Provider (This does not purchase the domain or otherwise interact with Registrars.)",
		Action: func(ctx *cli.Context) error {
			return exit(CreateDomains(args))
		},
		Flags: args.flags(),
		Before: func(context *cli.Context) error {
			fmt.Println("DEPRECATED: This command is deprecated. The domain is automatically created at the Domain Service Provider during the push command.")
			fmt.Println("DEPRECATED: To prevent disable auto-creating, use --no-populate with the push command.")
			return nil
		},
	}
}())

// CreateDomainsArgs args required for the create-domain subcommand.
type CreateDomainsArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
}

func (args *CreateDomainsArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

// CreateDomains contains all data/flags needed to run create-domains, independently of CLI.
func CreateDomains(args CreateDomainsArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	providerConfigs, err := credsfile.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return err
	}
	_, err = InitializeProviders(cfg, providerConfigs, false)
	if err != nil {
		return err
	}
	for _, domain := range cfg.Domains {
		fmt.Println("*** ", domain.Name)
		for _, provider := range domain.DNSProviderInstances {
			if creator, ok := provider.Driver.(providers.ZoneCreator); ok {
				fmt.Println("  -", provider.Name)
				err := creator.EnsureZoneExists(domain.Name)
				if err != nil {
					fmt.Printf("Error creating domain: %s\n", err)
				}
			}
		}
	}
	return nil
}
