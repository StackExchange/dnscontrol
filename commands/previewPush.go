package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v3/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v3/pkg/notifications"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/providers"
	"github.com/StackExchange/dnscontrol/v3/providers/config"
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
	Notify      bool
	WarnChanges bool
}

func (args *PreviewArgs) flags() []cli.Flag {
	flags := args.GetDNSConfigArgs.flags()
	flags = append(flags, args.GetCredentialsArgs.flags()...)
	flags = append(flags, args.FilterArgs.flags()...)
	flags = append(flags, &cli.BoolFlag{
		Name:        "notify",
		Destination: &args.Notify,
		Usage:       `set to true to send notifications to configured destinations`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "expect-no-changes",
		Destination: &args.WarnChanges,
		Usage:       `set to true for non-zero return code if there are changes`,
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
	flags = append(flags, &cli.BoolFlag{
		Name:        "i",
		Destination: &args.Interactive,
		Usage:       "Interactive. Confirm or Exclude each correction before they run",
	})
	return flags
}

// Preview implements the preview subcommand.
func Preview(args PreviewArgs) error {
	return run(args, false, false, printer.DefaultPrinter)
}

// Push implements the push subcommand.
func Push(args PushArgs) error {
	return run(args.PreviewArgs, true, args.Interactive, printer.DefaultPrinter)
}

// run is the main routine common to preview/push
func run(args PreviewArgs, push bool, interactive bool, out printer.CLI) error {
	// TODO: make truly CLI independent. Perhaps return results on a channel as they occur
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}
	errs := normalize.ValidateAndNormalizeConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("exiting due to validation errors")
	}
	// TODO:
	notifier, err := InitializeProviders(args.CredsFile, cfg, args.Notify)
	if err != nil {
		return err
	}
	anyErrors := false
	totalCorrections := 0
DomainLoop:
	for _, domain := range cfg.Domains {
		if !args.shouldRunDomain(domain.UniqueName) {
			continue
		}
		out.StartDomain(domain.UniqueName)
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

			/// This is where we should audit?

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
	out.Printf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		return fmt.Errorf("completed with errors")
	}
	if totalCorrections != 0 && args.WarnChanges {
		return fmt.Errorf("there are pending changes")
	}
	return nil
}

// InitializeProviders takes a creds file path and a DNSConfig object. Creates all providers with the proper types, and returns them.
// nonDefaultProviders is a list of providers that should not be run unless explicitly asked for by flags.
func InitializeProviders(credsFile string, cfg *models.DNSConfig, notifyFlag bool) (notify notifications.Notifier, err error) {
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
	isNonDefault := map[string]bool{}
	for name, vals := range providerConfigs {
		// add "_exclude_from_defaults":"true" to a provider to exclude it from being run unless
		// -providers=all or -providers=name
		if vals["_exclude_from_defaults"] == "true" {
			isNonDefault[name] = true
		}
	}

	// Find all "-" provider names and replace with actual provider.
	populateProviderTypes(cfg, providerConfigs)

	registrars := map[string]providers.Registrar{}
	dnsProviders := map[string]providers.DNSServiceProvider{}
	for _, d := range cfg.Domains {
		if registrars[d.RegistrarName] == nil {
			rCfg := cfg.RegistrarsByName[d.RegistrarName]
			r, err := providers.CreateRegistrar(rCfg.Type, providerConfigs[d.RegistrarName])
			if err != nil {
				return nil, err
			}
			registrars[d.RegistrarName] = r
		}
		d.RegistrarInstance.Driver = registrars[d.RegistrarName]
		d.RegistrarInstance.IsDefault = !isNonDefault[d.RegistrarName]
		for _, pInst := range d.DNSProviderInstances {
			if dnsProviders[pInst.Name] == nil {
				dCfg := cfg.DNSProvidersByName[pInst.Name]
				prov, err := providers.CreateDNSProvider(dCfg.Type, providerConfigs[dCfg.Name], dCfg.Metadata)
				if err != nil {
					return nil, err
				}
				dnsProviders[pInst.Name] = prov
			}
			pInst.Driver = dnsProviders[pInst.Name]
			pInst.IsDefault = !isNonDefault[pInst.Name]
		}
	}
	return
}

// populateProviderTypes scans a DNSConfig for blank provider types and fills them in based on providerConfigs.
func populateProviderTypes(cfg *models.DNSConfig, providerConfigs map[string]map[string]string) ([]string, error) {
	var msgs []string

	for i := range cfg.Registrars {
		pName := cfg.Registrars[i].Name
		pType := cfg.Registrars[i].Type
		nt, warnMsg, err := refineProviderType(pType, providerConfigs[pName])
		cfg.Registrars[i].Type = nt
		if warnMsg != "" {
			msgs = append(msgs, warnMsg)
		}
		if err != nil {
			return msgs, err
		}
	}

	for i := range cfg.DNSProviders {
		pName := cfg.DNSProviders[i].Name
		pType := cfg.DNSProviders[i].Type
		nt, warnMsg, err := refineProviderType(pType, providerConfigs[pName])
		cfg.Registrars[i].Type = nt
		if warnMsg != "" {
			msgs = append(msgs, warnMsg)
		}
		if err != nil {
			return msgs, err
		}
	}

	return msgs, nil
}

func refineProviderType(t string, credFields map[string]string) (replacementType string, warnMsg string, err error) {

	// t="" and t="-" are processed the same. Standardize on "" to reduce the number of cases to check.
	if t == "-" {
		t = ""
	}

	// Handle the preferred case.
	ct := t
	if credFields == nil {
		ct = t
		return fmt.Sprintf(`WARNING: creds.json is missing an entry called %q. It should look like: %q: { "PROVIDER": "NONE" },`, n_Easier, n_Easier), ""
		// In 3.x this is permitted.
		// In 4.0 this will be an error or maybe we'll default to ct = "NONE".
	} else {
		ct = providerConfigs[n_Easier]["PROVIDER"]
	}
	if ct == "-" {
		return "", fmt.Sprintf("Provider %q has invalid PROVIDER field: %q", n_Easier, ct), true
	}

	// name     type    credsType
	// usergan  -       GANDI        lookup worked. Nothing to say.
	// usergan  GANDI   ""           "Working but.... Please fix as follows..."
	// usergan  GANDI   GANDI        "working but unneeded: clean up as follows..."
	// usergan  GANDI   NAMEDOT      "error mismatched: please fix as follows..."

	if t == "-" {
		// "-" means "look in creds.json for the value". Some day this will be the norm.
		if ct == "" {
			// creds.json is missing the PROVIDER field.
			return "", fmt.Sprintf("creds.json entry %q is missing the PROVIDER field. See https://FILL IN#creds", n_Easier), true
			// In 4.0, this will be a hard error.
		}
		return ct, "", false
	}

	if ct == "" {
		return "", fmt.Sprintf("Provider %q has no PROVIDER field. Please update. See https://FILLIN#creds", n_Easier), true
	}

	if t != ct {
		// creds.json lists a PROVIDER but it doesn't match what's in dnsconfig.js!
		return t, fmt.Sprintf("creds.json entry %q has PROVIDER set to %q but dnsconfig.js specifies %q, which is a mismatch. See https://FILL IN#mismatch", n_Easier, ct, t), true
	}
	// User has updated creds.json but is still providing redundant information.
	return t, fmt.Sprintf("creds.json entry %q is valid. Please update dnsconfig.js. See https://FILL IN#dnsconfig", n_Easier), false

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
