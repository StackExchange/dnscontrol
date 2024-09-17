package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/idna"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/bindserial"
	"github.com/StackExchange/dnscontrol/v4/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v4/pkg/nameservers"
	"github.com/StackExchange/dnscontrol/v4/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v4/pkg/notifications"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/pkg/rfc4183"
	"github.com/StackExchange/dnscontrol/v4/pkg/zonerecs"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
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
	NoPopulate  bool
	Full        bool
}

// ReportItem is a record of corrections for a particular domain/provider/registrar.
type ReportItem struct {
	Domain      string `json:"domain"`
	Corrections int    `json:"corrections"`
	Provider    string `json:"provider,omitempty"`
	Registrar   string `json:"registrar,omitempty"`
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
	flags = append(flags, &cli.BoolFlag{
		Name:        "no-populate",
		Destination: &args.NoPopulate,
		Usage:       `Use this flag to not auto-create non-existing zones at the provider`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "full",
		Destination: &args.Full,
		Usage:       `Add headings, providers names, notifications of no changes, etc`,
	})
	flags = append(flags, &cli.IntFlag{
		Name:   "reportmax",
		Hidden: true,
		Usage:  `Limit the IGNORE/NO_PURGE report to this many lines (Expermental. Will change in the future.)`,
		Action: func(ctx *cli.Context, max int) error {
			printer.MaxReport = max
			return nil
		},
	})
	flags = append(flags, &cli.Int64Flag{
		Name:        "bindserial",
		Destination: &bindserial.ForcedValue,
		Usage:       `Force BIND serial numbers to this value (for reproducibility)`,
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
	Report      string
}

func (args *PushArgs) flags() []cli.Flag {
	flags := args.PreviewArgs.flags()
	flags = append(flags, &cli.BoolFlag{
		Name:        "i",
		Destination: &args.Interactive,
		Usage:       "Interactive. Confirm or Exclude each correction before they run",
	})
	flags = append(flags, &cli.StringFlag{
		Name:        "report",
		Destination: &args.Report,
		Usage:       `Generate a machine-parseable report of performed corrections.`,
	})
	return flags
}

// Preview implements the preview subcommand.
func Preview(args PreviewArgs) error {
	return run(args, false, false, printer.DefaultPrinter, nil)
}

// Push implements the push subcommand.
func Push(args PushArgs) error {
	return run(args.PreviewArgs, true, args.Interactive, printer.DefaultPrinter, &args.Report)
}

var obsoleteDiff2FlagUsed = false

// run is the main routine common to preview/push
func run(args PreviewArgs, push bool, interactive bool, out printer.CLI, report *string) error {
	// TODO: make truly CLI independent. Perhaps return results on a channel as they occur

	// This is a hack until we have the new printer replacement.
	printer.SkinnyReport = !args.Full

	if obsoleteDiff2FlagUsed {
		printer.Println("WARNING: Please remove obsolete --diff2 flag. This will be an error in v5 or later. See https://github.com/StackExchange/dnscontrol/issues/2262")
	}

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

	// create a WaitGroup with the length of domains for the anonymous functions (later goroutines) to wait for
	var wg sync.WaitGroup
	wg.Add(len(cfg.Domains))
	var reportItems []ReportItem
	// For each domain in dnsconfig.js...
	for _, domain := range cfg.Domains {
		// Run preview or push operations per domain as anonymous function, in preparation for the later use of goroutines.
		// For now running this code is still sequential.
		// Please note that at the end of this anonymous function there is a } (domain) which executes this function actually
		func(domain *models.DomainConfig) {
			defer wg.Done() // defer notify WaitGroup this anonymous function has finished

			uniquename := domain.GetUniqueName()
			if !args.shouldRunDomain(uniquename) {
				return
			}

			err = domain.Punycode()
			if err != nil {
				return
			}

			// Correct the domain...

			out.StartDomain(uniquename)
			var providersWithExistingZone []*models.DNSProviderInstance
			/// For each DSP...
			for _, provider := range domain.DNSProviderInstances {
				if !args.NoPopulate {
					// preview run: check if zone is already there, if not print a warning
					if lister, ok := provider.Driver.(providers.ZoneLister); ok && !push {
						zones, err := lister.ListZones()
						if err != nil {
							out.Errorf("ERROR: %s\n", err.Error())
							return
						}
						aceZoneName, _ := idna.ToASCII(domain.Name)

						if !slices.Contains(zones, aceZoneName) {
							//out.Warnf("DEBUG: zones: %v\n", zones)
							//out.Warnf("DEBUG: Name: %v\n", domain.Name)

							out.Warnf("Zone '%s' does not exist in the '%s' profile and will be added automatically.\n", domain.Name, provider.Name)
							continue // continue with next provider, as we can not determine corrections without an existing zone
						}
					} else if creator, ok := provider.Driver.(providers.ZoneCreator); ok && push {
						// this is the actual push, ensure domain exists at DSP
						if err := creator.EnsureZoneExists(domain.Name); err != nil {
							out.Warnf("Error creating domain: %s\n", err)
							anyErrors = true
							continue // continue with next provider, as we couldn't create this one
						}
					}
				}
				providersWithExistingZone = append(providersWithExistingZone, provider)
			}

			// Correct the registrar...

			nsList, err := nameservers.DetermineNameserversForProviders(domain, providersWithExistingZone, false)
			if err != nil {
				out.Errorf("ERROR: %s\n", err.Error())
				return
			}
			domain.Nameservers = nsList
			nameservers.AddNSRecords(domain)

			for _, provider := range providersWithExistingZone {

				shouldrun := args.shouldRunProvider(provider.Name, domain)
				out.StartDNSProvider(provider.Name, !shouldrun)
				if !shouldrun {
					continue
				}

				reports, corrections, actualChangeCount, err := zonerecs.CorrectZoneRecords(provider.Driver, domain)
				out.EndProvider(provider.Name, actualChangeCount, err)
				if err != nil {
					anyErrors = true
					return
				}
				totalCorrections += actualChangeCount
				printReports(domain.Name, provider.Name, reports, out, push, notifier)
				reportItems = append(reportItems, ReportItem{
					Domain:      domain.Name,
					Corrections: actualChangeCount,
					Provider:    provider.Name,
				})
				anyErrors = printOrRunCorrections(domain.Name, provider.Name, corrections, out, push, interactive, notifier) || anyErrors
			}

			//
			run := args.shouldRunProvider(domain.RegistrarName, domain)
			out.StartRegistrar(domain.RegistrarName, !run)
			if !run {
				return
			}
			if len(domain.Nameservers) == 0 && domain.Metadata["no_ns"] != "true" {
				out.Warnf("No nameservers declared; skipping registrar. Add {no_ns:'true'} to force.\n")
				return
			}

			corrections, err := domain.RegistrarInstance.Driver.GetRegistrarCorrections(domain)
			out.EndProvider(domain.RegistrarName, len(corrections), err)
			if err != nil {
				anyErrors = true
				return
			}
			totalCorrections += len(corrections)
			reportItems = append(reportItems, ReportItem{
				Domain:      domain.Name,
				Corrections: len(corrections),
				Registrar:   domain.RegistrarName,
			})
			anyErrors = printOrRunCorrections(domain.Name, domain.RegistrarName, corrections, out, push, interactive, notifier) || anyErrors
		}(domain)
	}
	wg.Wait() // wait for all anonymous functions to finish

	if os.Getenv("TEAMCITY_VERSION") != "" {
		fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
	}
	rfc4183.PrintWarning()
	notifier.Done()
	out.Printf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		return fmt.Errorf("completed with errors")
	}
	if totalCorrections != 0 && args.WarnChanges {
		return fmt.Errorf("there are pending changes")
	}
	if report != nil && *report != "" {
		f, err := os.OpenFile(*report, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		b, err := json.MarshalIndent(reportItems, "", "  ")
		if err != nil {
			return err
		}
		if _, err := f.Write(b); err != nil {
			return err
		}
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
const url = "https://docs.dnscontrol.org/commands/creds-json"

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
			if correction.F != nil {
				err = correction.F()
				out.EndCorrection(err)
				if err != nil {
					anyErrors = true
				}
			}
		}
		notifier.Notify(domain, provider, correction.Msg, err, !push)
	}
	return anyErrors
}

func printReports(domain string, provider string, reports []*models.Correction, out printer.CLI, push bool, notifier notifications.Notifier) (anyErrors bool) {
	anyErrors = false
	if len(reports) == 0 {
		return false
	}
	for i, report := range reports {
		out.PrintReport(i, report)
		notifier.Notify(domain, provider, report.Msg, nil, !push)
	}
	return anyErrors
}
