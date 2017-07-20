package cmd

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/normalize"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
	"github.com/urfave/cli"
)

var previewCommand = &cli.Command{
	Name:  "preview",
	Usage: "read live configuration and identify changes to be made, without applying them",
	Action: func(ctx *cli.Context) error {
		return exit(Preview(globalPreviewArgs))
	},
	Category: catMain,
	Flags:    globalPreviewArgs.flags(),
}

// PreviewArgs contains all data/flags needed to run preview, independently of CLI
type PreviewArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
}

func (args *PreviewArgs) flags() []cli.Flag {
	flags := globalPreviewArgs.GetDNSConfigArgs.flags()
	flags = append(flags, globalPreviewArgs.GetCredentialsArgs.flags()...)
	return flags
}

var globalPreviewArgs PreviewArgs

func Preview(args PreviewArgs) error {
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	errs := normalize.NormalizeAndValidateConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("Exiting due to validation errors")
	}
	registrars, dnsProviders, nonDefaultProviders, err := InitializeProviders(args.CredsFile, cfg)
	if err != nil {
		return err
	}
	fmt.Printf("Initialized %d registrars and %d dns service providers.\n", len(registrars), len(dnsProviders))
	fmt.Println(len(nonDefaultProviders))
	return nil
}

func InitializeProviders(credsFile string, cfg *models.DNSConfig) (registrars map[string]providers.Registrar, dnsProviders map[string]providers.DNSServiceProvider, nonDefaultProviders []string, err error) {
	var providerConfigs map[string]map[string]string
	providerConfigs, err = config.LoadProviderConfigs(credsFile)
	if err != nil {
		return
	}
	nonDefaultProviders = []string{}
	for name, vals := range providerConfigs {
		// add "_exclude_from_defaults":"true" to a domain to exclude it from being run unless
		// -providers=all or -providers=name
		if vals["_exclude_from_defaults"] == "true" {
			nonDefaultProviders = append(nonDefaultProviders, name)
		}
	}
	registrars, err = providers.CreateRegistrars(cfg, providerConfigs)
	if err != nil {
		return
	}
	dnsProviders, err = providers.CreateDsps(cfg, providerConfigs)
	if err != nil {
		return
	}
	return
}
