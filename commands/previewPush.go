package commands

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/pkg/normalize"
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
	Delay int
}

func (args *PreviewArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	flags = append(flags, args.FilterArgs.flags()...)
	flags = append(flags, cli.IntFlag{
		Name:        "d",
		Destination: &args.Delay,
		Usage:       "delay between domains to avoid rate limits (in ms)",
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

func Preview(args PreviewArgs) error {
	return run(args, false, false, printer.ConsolePrinter{})
}

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
	registrars, dnsProviders, nonDefaultProviders, err := InitializeProviders(args.CredsFile, cfg)
	if err != nil {
		return err
	}
	out.Debugf("Initialized %d registrars and %d dns service providers.\n", len(registrars), len(dnsProviders))
	anyErrors := false
	totalCorrections := 0
DomainLoop:
	for _, domain := range cfg.Domains {
		if !args.shouldRunDomain(domain.Name) {
			continue
		}
		out.StartDomain(domain.Name)
		nsList, err := nameservers.DetermineNameservers(domain, 0, dnsProviders)
		if err != nil {
			return err
		}
		domain.Nameservers = nsList
		nameservers.AddNSRecords(domain)
		for prov := range domain.DNSProviders {
			dc, err := domain.Copy()
			if err != nil {
				return err
			}
			shouldrun := args.shouldRunProvider(prov, dc, nonDefaultProviders)
			out.StartDNSProvider(prov, !shouldrun)
			if !shouldrun {
				continue
			}
			// TODO: make provider discovery like this a validate-time operation
			dsp, ok := dnsProviders[prov]
			if !ok {
				log.Fatalf("DSP %s not declared.", prov)
			}
			corrections, err := dsp.GetDomainCorrections(dc)
			out.EndProvider(len(corrections), err)
			if err != nil {
				anyErrors = true
				continue DomainLoop
			}
			totalCorrections += len(corrections)
			anyErrors = printOrRunCorrections(corrections, out, push, interactive) || anyErrors
		}
		run := args.shouldRunProvider(domain.Registrar, domain, nonDefaultProviders)
		out.StartRegistrar(domain.Registrar, !run)
		if !run {
			continue
		}
		reg, ok := registrars[domain.Registrar]
		if !ok {
			log.Fatalf("Registrar %s not declared.", reg)
		}
		if len(domain.Nameservers) == 0 && domain.Metadata["no_ns"] != "true" {
			out.Warnf("No nameservers declared; skipping registrar. Add {no_ns:'true'} to force.\n")
			continue
		}
		dc, err := domain.Copy()
		if err != nil {
			log.Fatal(err)
		}
		corrections, err := reg.GetRegistrarCorrections(dc)
		out.EndProvider(len(corrections), err)
		if err != nil {
			anyErrors = true
			continue
		}
		totalCorrections += len(corrections)
		anyErrors = printOrRunCorrections(corrections, out, push, interactive) || anyErrors
		if args.Delay != 0 {
			time.Sleep(time.Duration(args.Delay) * time.Millisecond)
		}
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
	}
	out.Debugf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		return fmt.Errorf("Completed with errors")
	}
	return nil
}

// InitializeProviders takes a creds file path and a DNSConfig object. Creates all providers with the proper types, and returns them.
// nonDefaultProviders is a list of providers that should not be run unless explicitly asked for by flags.
func InitializeProviders(credsFile string, cfg *models.DNSConfig) (registrars map[string]providers.Registrar, dnsProviders map[string]providers.DNSServiceProvider, nonDefaultProviders []string, err error) {
	var providerConfigs map[string]map[string]string
	providerConfigs, err = config.LoadProviderConfigs(credsFile)
	if err != nil {
		return
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

func printOrRunCorrections(corrections []*models.Correction, out printer.CLI, push bool, interactive bool) (anyErrors bool) {
	anyErrors = false
	if len(corrections) == 0 {
		return false
	}
	for i, correction := range corrections {
		out.PrintCorrection(i, correction)
		if push {
			if interactive && !out.PromptToRun() {
				continue
			}
			err := correction.F()
			out.EndCorrection(err)
			if err != nil {
				anyErrors = true
			}
		}
	}
	return anyErrors
}
