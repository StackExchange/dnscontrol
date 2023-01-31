package commands

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args GetUnmanagedArgs
	return &cli.Command{
		Name:  "list-unmanaged",
		Usage: "gets a zone from a provider (stand-alone)",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return cli.Exit("You need to supply a credential name as first argument", 1)
			}
			args.CredName = ctx.Args().Get(0)

			return exit(GetUnmanaged(args))
		},
		Flags:     args.flags(),
		UsageText: "dnscontrol list-unmanaged [command options] credkey",
		Description: `List unmanaged domains and zones.  This is a stand-alone utility.

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)

EXAMPLES:
   dnscontrol list-unmanaged mycred`,
	}
}())

type GetUnmanagedArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs        // Args related to creds.json
	CredName           string // key in creds.json
}

func (args *GetUnmanagedArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

func GetUnmanaged(args GetUnmanagedArgs) error {
	var providerConfigs map[string]map[string]string
	var err error

	// Read it in:
	providerConfigs, err = credsfile.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return fmt.Errorf("failed GetZone LoadProviderConfigs(%q): %w", args.CredsFile, err)
	}
	provider, err := providers.CreateDNSProvider("-", providerConfigs[args.CredName], nil)
	if err != nil {
		return fmt.Errorf("failed GetZone CDP: %w", err)
	}

	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return fmt.Errorf("Error getting dnsconfig")
	}

	managedDomains := make([]string, 0, len(cfg.Domains))
	for _, zone := range cfg.Domains {
		managedDomains = append(managedDomains, zone.Name)
	}

	domainLister, ok := provider.(providers.DomainLister)
	if ok {
		deployedDomains, err := domainLister.ListDomains()
		if err != nil {
			return fmt.Errorf("failed ListDomains: %w\n", err)
		}

		fmt.Printf("Unmanaged domains:")
		for _, deployedDomain := range deployedDomains {
			if !slices.Contains(managedDomains, deployedDomain) {
				fmt.Printf("%s\n", deployedDomain)
			}
		}
	} else {
		fmt.Printf("provider type of %s cannot list domains\n", args.CredName)
	}

	zoneLister, ok := provider.(providers.ZoneLister)
	if ok {
		deployedZones, err := zoneLister.ListZones()
		if err != nil {
			return fmt.Errorf("failed ListZones: %w\n", err)
		}

		fmt.Printf("\nUnmanaged zones:\n")
		for _, deployedZone := range deployedZones {
			if !slices.Contains(managedDomains, deployedZone) {
				fmt.Printf("%s\n", deployedZone)
			}
		}
	} else {
		fmt.Printf("provider type of %s cannot list zones\n", args.CredName)
	}

	return nil
}
