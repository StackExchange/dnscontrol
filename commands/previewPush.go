package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/pkg/normalize"
	"github.com/StackExchange/dnscontrol/pkg/notifications"
	"github.com/StackExchange/dnscontrol/pkg/printer"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
	"github.com/urfave/cli"
)

var _ = cmd(catMain, func() *cli.Command {
	var args PreviewArgs
	return &cli.Command{
		Name:  "preview",
		Usage: "read live configuration and identify changes to be made, without applying them",
		Action: func(ctx *cli.Context) error {
			return exit(Preview(args))
		},
		Flags: args.flags(),
	}
}())

// PreviewArgs contains all data/flags needed to run preview, independently of CLI
type PreviewArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
	FilterArgs
	Notify bool
}

func (args *PreviewArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	flags = append(flags, args.FilterArgs.flags()...)
	flags = append(flags, cli.BoolFlag{
		Name:        "notify",
		Destination: &args.Notify,
		Usage:       `set to true to send notifications to configured destinations`,
	})
	return flags
}

var _ = cmd(catMain, func() *cli.Command {
	var args PushArgs
	return &cli.Command{
		Name:  "push",
		Usage: "identify changes to be made, and perform them",
		Action: func(ctx *cli.Context) error {
			return exit(Push(args))
		},
		Flags: args.flags(),
	}
}())

// PushArgs contains all data/flags needed to run push, independently of CLI
type PushArgs struct {
	PreviewArgs
	Interactive bool
}

func (args *PushArgs) flags() []cli.Flag {
	flags := args.PreviewArgs.flags()
	flags = append(flags, cli.BoolFlag{
		Name:        "i",
		Destination: &args.Interactive,
		Usage:       "Interactive. Confirm or Exclude each correction before they run",
	})
	return flags
}

// Preview implements the preview subcommand.
func Preview(args PreviewArgs) error {
	return run(args, false, false, printer.ConsolePrinter{})
}

// Push implements the push subcommand.
func Push(args PushArgs) error {
	return run(args.PreviewArgs, true, args.Interactive, printer.ConsolePrinter{})
}

// run is the main routine common to preview/push
func run(args PreviewArgs, push bool, interactive bool, out printer.CLI) error {
	// TODO: make truly CLI independent. Perhaps return results on a channel as they occur
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	errs := normalize.NormalizeAndValidateConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("Exiting due to validation errors")
	}
	// TODO:
	//registrars, dnsProviders, nonDefaultProviders, notifier, err := InitializeProviders(args.CredsFile, cfg, args.Notify)
	anyErrors := false
	totalCorrections := 0
DomainLoop:
	for _, domain := range cfg.Domains {
		if !args.shouldRunDomain(domain.Name) {
			continue
		}
		out.StartDomain(domain.Name)
		nsList, err := nameservers.DetermineNameservers(domain)
		if err != nil {
			return err
		}
		domain.Nameservers = nsList
		nameservers.AddNSRecords(domain)
		for _, provider := range domain.DNSProviderInstances {
			dc, err := domain.Copy()
			if err != nil {
				return err
			}
			shouldrun := args.shouldRunProvider(provider.Name, dc)
			out.StartDNSProvider(provider.Name, !shouldrun)
			if !shouldrun {
				continue
			}
			corrections, err := provider.Driver.GetDomainCorrections(dc)
			out.EndProvider(len(corrections), err)
			if err != nil {
				anyErrors = true
				continue DomainLoop
			}
			totalCorrections += len(corrections)
			anyErrors = printOrRunCorrections(domain.Name, provider.Name, corrections, out, push, interactive, notifier) || anyErrors
		}
		run := args.shouldRunProvider(domain.RegistrarName, domain)
		out.StartRegistrar(domain.RegistrarName, !run)
		if !run {
			continue
		}
		if len(domain.Nameservers) == 0 && domain.Metadata["no_ns"] != "true" {
			out.Warnf("No nameservers declared; skipping registrar. Add {no_ns:'true'} to force.\n")
			continue
		}
		dc, err := domain.Copy()
		if err != nil {
			log.Fatal(err)
		}
		corrections, err := domain.RegistrarInstance.Driver.GetRegistrarCorrections(dc)
		out.EndProvider(len(corrections), err)
		if err != nil {
			anyErrors = true
			continue
		}
		totalCorrections += len(corrections)
		anyErrors = printOrRunCorrections(domain.Name, domain.RegistrarName, corrections, out, push, interactive, notifier) || anyErrors
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
	}
	notifier.Done()
	out.Debugf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		return fmt.Errorf("Completed with errors")
	}
	return nil
}

// InitializeProviders takes a creds file path and a DNSConfig object. Creates all providers with the proper types, and returns them.
// nonDefaultProviders is a list of providers that should not be run unless explicitly asked for by flags.
func InitializeProviders(credsFile string, cfg *models.DNSConfig, notifyFlag bool) (registrars map[string]providers.Registrar, dnsProviders map[string]providers.DNSServiceProvider, nonDefaultProviders []string, notify notifications.Notifier, err error) {
	var providerConfigs map[string]map[string]string
	var notificationCfg map[string]string
	defer func() {
		notify = notifications.Init(notificationCfg)
	}()
	providerConfigs, err = config.LoadProviderConfigs(credsFile)
	if err != nil {
		return
	}
	if notifyFlag {
		notificationCfg = providerConfigs["notifications"]
	}
	nonDefaultProviders = []string{}
	for name, vals := range providerConfigs {
		// add "_exclude_from_defaults":"true" to a provider to exclude it from being run unless
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

func printOrRunCorrections(domain string, provider string, corrections []*models.Correction, out printer.CLI, push bool, interactive bool, notifier notifications.Notifier) (anyErrors bool) {
	anyErrors = false
	if len(corrections) == 0 {
		return false
	}
	for i, correction := range corrections {
		out.PrintCorrection(i, correction)
		var err error
		if push {
			if interactive && !out.PromptToRun() {
				continue
			}
			err = correction.F()
			out.EndCorrection(err)
			if err != nil {
				anyErrors = true
			}
		}
		notifier.Notify(domain, provider, correction.Msg, err, !push)
	}
	return anyErrors
}
