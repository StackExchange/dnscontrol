package commands

import (
	"fmt"
	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
	"os"
)

var _ = cmd(catUtils, func() *cli.Command {
	var args GetUnmanagedZonesArgs
	return &cli.Command{
		Name:  "list-unmanaged-zones",
		Usage: "gets a zone from a provider (stand-alone)",
		Action: func(ctx *cli.Context) error {
			if ctx.NArg() < 1 {
				return cli.Exit("You need to supply a credential name as first argument", 1)
			}
			args.CredName = ctx.Args().Get(0)

			return exit(GetUnmanagedZones(args))
		},
		Flags:     args.flags(),
		UsageText: "dnscontrol list-unmanaged-zones [command options] credkey",
		Description: `List unmanaged zones.  This is a stand-alone utility.

ARGUMENTS:
   credkey:  The name used in creds.json (first parameter to NewDnsProvider() in dnsconfig.js)

EXAMPLES:
   dnscontrol list-unmanaged-zones mycred`,
	}
}())

type GetUnmanagedZonesArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs        // Args related to creds.json
	CredName           string // key in creds.json
}

func (args *GetUnmanagedZonesArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	return flags
}

func GetUnmanagedZones(args GetUnmanagedZonesArgs) error {
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

	lister, ok := provider.(providers.ZoneLister)
	if !ok {
		return fmt.Errorf("provider type of %s cannot list zones to use the 'list-unmanaged-zones' feature", args.CredName)
	}
	deployedZones, err := lister.ListZones()
	if err != nil {
		return fmt.Errorf("failed GetZone LZ: %w", err)
	}

	// first open output stream and print initial header (if applicable)
	w := os.Stdout
	defer w.Close()

	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		fmt.Fprintln(w, "Error getting dnsconfig")
	}

	managedZones := make([]string, 0, len(cfg.Domains))
	for _, zone := range cfg.Domains {
		managedZones = append(managedZones, zone.Name)
	}

	for _, deployedZone := range deployedZones {
		if !slices.Contains(managedZones, deployedZone) {
			fmt.Fprintln(w, deployedZone)
		}
	}

	return nil
}
