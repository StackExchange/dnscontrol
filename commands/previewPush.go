package commands

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/StackExchange/dnscontrol/v3/models"
	"github.com/StackExchange/dnscontrol/v3/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v3/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v3/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v3/pkg/notifications"
	"github.com/StackExchange/dnscontrol/v3/pkg/printer"
	"github.com/StackExchange/dnscontrol/v3/providers"
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
	providerConfigs, err := credsfile.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return err
	}
	notifier, err := InitializeProviders(cfg, providerConfigs, args.Notify)
	if err != nil {
		return err
	}

	errs := normalize.ValidateAndNormalizeConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("exiting due to validation errors")
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

// InitializeProviders takes (fully processed) configuration and instantiates all providers and returns them.
func InitializeProviders(cfg *models.DNSConfig, providerConfigs map[string]map[string]string, notifyFlag bool) (notify notifications.Notifier, err error) {
	var notificationCfg map[string]string
	defer func() {
		notify = notifications.Init(notificationCfg)
	}()
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

	// Populate provider type ids based on values from creds.json:
	msgs, err := populateProviderTypes(cfg, providerConfigs)
	if len(msgs) != 0 {
		fmt.Fprintln(os.Stderr, strings.Join(msgs, "\n"))
	}
	if err != nil {
		return
	}

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

// providerTypeFieldName is the name of the field in creds.json that specifies the provider type id.
const providerTypeFieldName = "TYPE"

// url is the documentation URL to list in the warnings related to missing provider type ids.
const url = "https://stackexchange.github.io/dnscontrol/creds-json"

// populateProviderTypes scans a DNSConfig for blank provider types and fills them in based on providerConfigs.
// That is, if the provider type is "-" or "", we take that as an flag
// that means this value should be replaced by the type found in creds.json.
func populateProviderTypes(cfg *models.DNSConfig, providerConfigs map[string]map[string]string) ([]string, error) {
	var msgs []string

	for i := range cfg.Registrars {
		pType := cfg.Registrars[i].Type
		pName := cfg.Registrars[i].Name
		nt, warnMsg, err := refineProviderType(pName, pType, providerConfigs[pName], "NewRegistrar")
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
		nt, warnMsg, err := refineProviderType(pName, pType, providerConfigs[pName], "NewDnsProvider")
		cfg.DNSProviders[i].Type = nt
		if warnMsg != "" {
			msgs = append(msgs, warnMsg)
		}
		if err != nil {
			return msgs, err
		}
	}

	// Update these fields set by // commands/commands.go:preloadProviders().
	// This is probably a layering violation.  That said, the
	// fundamental problem here is that we're storing the provider
	// instances by string name, not by a pointer to a struct.  We
	// should clean that up someday.
	for _, domain := range cfg.Domains { // For each domain..
		for _, provider := range domain.DNSProviderInstances { // For each provider...
			pName := provider.ProviderBase.Name
			pType := provider.ProviderBase.ProviderType
			nt, warnMsg, err := refineProviderType(pName, pType, providerConfigs[pName], "NewDnsProvider")
			provider.ProviderBase.ProviderType = nt
			if warnMsg != "" {
				msgs = append(msgs, warnMsg)
			}
			if err != nil {
				return msgs, err
			}
		}
		p := domain.RegistrarInstance
		pName := p.Name
		pType := p.ProviderType
		nt, warnMsg, err := refineProviderType(pName, pType, providerConfigs[pName], "NewRegistrar")
		p.ProviderType = nt
		if warnMsg != "" {
			msgs = append(msgs, warnMsg)
		}
		if err != nil {
			return msgs, err
		}
	}

	return uniqueStrings(msgs), nil
}

// uniqueStrings takes an unsorted slice of strings and returns the
// unique strings, in the order they first appeared in the list.
func uniqueStrings(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func refineProviderType(credEntryName string, t string, credFields map[string]string, source string) (replacementType string, warnMsg string, err error) {

	// t="" and t="-" are processed the same. Standardize on "-" to reduce the number of cases to check.
	if t == "" {
		t = "-"
	}

	// Use cases:
	//
	// type     credsType
	// ----     ---------
	// - or ""  GANDI        lookup worked. Nothing to say.
	// - or ""  - or ""      ERROR "creds.json has invalid or missing data"
	// GANDI    ""           WARNING "Working but.... Please fix as follows..."
	// GANDI    GANDI        INFO "working but unneeded: clean up as follows..."
	// GANDI    NAMEDOT      ERROR "error mismatched: please fix as follows..."

	// ERROR: Invalid.
	// WARNING: Required change to remain compatible with 4.0
	// INFO: Post-4.0 cleanups or other non-required changes.

	if t != "-" {
		// Old-style, dnsconfig.js specifies the type explicitly.
		// This is supported but we suggest updates for future compatibility.

		// If credFields is nil, that means there was no entry in creds.json:
		if credFields == nil {
			// Warn the user to update creds.json in preparation for 4.0:
			// In 4.0 this should be an error.  We could default to a
			// provider such as "NONE" but I suspect it would be confusing
			// to users to see references to a provider name that they did
			// not specify.
			return t, fmt.Sprintf(`WARNING: For future compatibility, add this entry creds.json: %q: { %q: %q }, (See %s#missing)`,
				credEntryName, providerTypeFieldName, t,
				url,
			), nil
		}

		switch ct := credFields[providerTypeFieldName]; ct {
		case "":
			// Warn the user to update creds.json in preparation for 4.0:
			// In 4.0 this should be an error.
			return t, fmt.Sprintf(`WARNING: For future compatibility, update the %q entry in creds.json by adding: %q: %q, (See %s#missing)`,
				credEntryName,
				providerTypeFieldName, t,
				url,
			), nil
		case "-":
			// This should never happen. The user is specifying "-" in a place that it shouldn't be used.
			return "-", "", fmt.Errorf(`ERROR: creds.json entry %q has invalid %q value %q (See %s#hyphen)`,
				credEntryName, providerTypeFieldName, ct,
				url,
			)
		case t:
			// creds.json file is compatible with and dnsconfig.js can be updated.
			return ct, fmt.Sprintf(`INFO: In dnsconfig.js %s(%q, %q) can be simplified to %s(%q) (See %s#cleanup)`,
				source, credEntryName, t,
				source, credEntryName,
				url,
			), nil
		default:
			// creds.json lists a TYPE but it doesn't match what's in dnsconfig.js!
			return t, "", fmt.Errorf(`ERROR: Mismatch found! creds.json entry %q has %q set to %q but dnsconfig.js specifies %s(%q, %q) (See %s#mismatch)`,
				credEntryName,
				providerTypeFieldName, ct,
				source, credEntryName, t,
				url,
			)
		}
	}

	// t == "-"
	// New-style, dnsconfig.js does not specify the type (t == "") or a
	// command line tool accepted "-" as a positional argument for
	// backwards compatibility.

	// If credFields is nil, that means there was no entry in creds.json:
	if credFields == nil {
		return "", "", fmt.Errorf(`ERROR: creds.json is missing an entry called %q. Suggestion: %q: { %q: %q }, (See %s#missing)`,
			credEntryName,
			credEntryName, providerTypeFieldName, "FILL_IN_PROVIDER_TYPE",
			url,
		)
	}

	// New-style, dnsconfig.js doesn't specifies the type. It will be
	// looked up in creds.json.
	switch ct := credFields[providerTypeFieldName]; ct {
	case "":
		return ct, "", fmt.Errorf(`ERROR: creds.json entry %q is missing: %q: %q, (See %s#fixcreds)`,
			credEntryName,
			providerTypeFieldName, "FILL_IN_PROVIDER_TYPE",
			url,
		)
	case "-":
		// This should never happen. The user is confused and specified "-" in the wrong place!
		return "-", "", fmt.Errorf(`ERROR: creds.json entry %q has invalid %q value %q (See %s#hyphen)`,
			credEntryName,
			providerTypeFieldName, ct,
			url,
		)
	default:
		// use the value in creds.json (this should be the normal case)
		return ct, "", nil
	}

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
