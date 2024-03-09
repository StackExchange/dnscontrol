package commands

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/StackExchange/dnscontrol/v4/models"
	"github.com/StackExchange/dnscontrol/v4/pkg/bindserial"
	"github.com/StackExchange/dnscontrol/v4/pkg/credsfile"
	"github.com/StackExchange/dnscontrol/v4/pkg/normalize"
	"github.com/StackExchange/dnscontrol/v4/pkg/notifications"
	"github.com/StackExchange/dnscontrol/v4/pkg/printer"
	"github.com/StackExchange/dnscontrol/v4/pkg/zonerecs"
	"github.com/StackExchange/dnscontrol/v4/providers"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
	"golang.org/x/net/idna"
)

var _ = cmd(catMain, func() *cli.Command {
	var args PPreviewArgs
	return &cli.Command{
		Name:  "ppreview",
		Usage: "read live configuration and identify changes to be made, without applying them",
		Action: func(ctx *cli.Context) error {
			return exit(PPreview(args))
		},
		Flags: args.flags(),
	}
}())

// PPreviewArgs contains all data/flags needed to run preview, independently of CLI
type PPreviewArgs struct {
	GetDNSConfigArgs
	GetCredentialsArgs
	FilterArgs
	Notify      bool
	WarnChanges bool
	NoPopulate  bool
	DePopulate  bool
	Full        bool
}

// ReportItem is a record of corrections for a particular domain/provider/registrar.
//type ReportItem struct {
//	Domain      string `json:"domain"`
//	Corrections int    `json:"corrections"`
//	Provider    string `json:"provider,omitempty"`
//	Registrar   string `json:"registrar,omitempty"`
//}

func (args *PPreviewArgs) flags() []cli.Flag {
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
		Usage:       `Do not auto-create zones at the provider`,
	})
	flags = append(flags, &cli.BoolFlag{
		Name:        "depopulate",
		Destination: &args.NoPopulate,
		Usage:       `Delete unknown zones at provider (dangerous!)`,
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
	var args PPushArgs
	return &cli.Command{
		Name:  "ppush",
		Usage: "identify changes to be made, and perform them",
		Action: func(ctx *cli.Context) error {
			return exit(PPush(args))
		},
		Flags: args.flags(),
	}
}())

// PPushArgs contains all data/flags needed to run push, independently of CLI
type PPushArgs struct {
	PPreviewArgs
	Interactive bool
	Report      string
}

func (args *PPushArgs) flags() []cli.Flag {
	flags := args.PPreviewArgs.flags()
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

// PPreview implements the preview subcommand.
func PPreview(args PPreviewArgs) error {
	return prun(args, false, false, printer.DefaultPrinter, "")
}

// PPush implements the push subcommand.
func PPush(args PPushArgs) error {
	return prun(args.PPreviewArgs, true, args.Interactive, printer.DefaultPrinter, args.Report)
}

var pobsoleteDiff2FlagUsed = false

// // create a WaitGroup with the length of domains for the anonymous functions (later goroutines) to wait for
// var wg sync.WaitGroup
// wg.Add(len(cfg.Domains))
// var reportItems []ReportItem
// // For each domain in dnsconfig.js...

/*

	foreach includedZone:
	    includedProviders = (for all providers, shouldRunProvider)
	    foreach includedProvider:
		    go
			    ZoneLister doesn't exists: goto next_step.
		        (mutex protected) get ZoneList if we don't already have it.
				if zone not in zonelist:
			    	If Populate:
						if ZoneCreator doesn't exist: goto next_step
				    	if not push: output "would have created zone"; then next_step.
			    		create zone.
						If error, (mutex protected: anyError = true) then done.
						(mutex protected) add zone to zonelist
				next_step:
		        	domain+provider.Nameservers = GetNameServers(domain) (zone not exist [return empty] vs. other error [return error])
					domain+provider.Records = GetZoneRecords() (if error, output errror and store nil)
					make NS and Rec corrections

		WaitGroup.

	foreach includedZone:
	    includedProviders = (for all providers, shouldRunProvider)
	    foreach includedProvider:
		    BeginProvider
			RunZoneCorrections (preview or push)
			RunParentCorrections (preview or push)
		    EndProvider
*/

// run is the main routine common to preview/push
func prun(args PPreviewArgs, push bool, interactive bool, out printer.CLI, report string) error {

	// TODO: make truly CLI independent. Perhaps return results on a channel as they occur

	// This is a hack until we have the new printer replacement.
	printer.SkinnyReport = !args.Full

	if pobsoleteDiff2FlagUsed {
		printer.Println("WARNING: Please remove obsolete --diff2 flag. This will be an error in v5 or later. See https://github.com/StackExchange/dnscontrol/issues/2262")
	}

	fmt.Printf("Reading dnsconfig.js or equiv.\n")
	cfg, err := GetDNSConfig(args.GetDNSConfigArgs)
	if err != nil {
		return err
	}

	fmt.Printf("Reading creds.json or equiv.\n")
	providerConfigs, err := credsfile.LoadProviderConfigs(args.CredsFile)
	if err != nil {
		return err
	}

	fmt.Printf("Creating an in-memory model of 'desired'...\n")
	notifier, err := PInitializeProviders(cfg, providerConfigs, args.Notify)
	if err != nil {
		return err
	}

	fmt.Printf("Normalizing and validating 'desired'..\n")
	errs := normalize.ValidateAndNormalizeConfig(cfg)
	if PrintValidationErrors(errs) {
		return fmt.Errorf("exiting due to validation errors")
	}

	fmt.Printf("Iterating over the zones...\n")

	// Loop over all (or some) zones:
	zonesToProcess := whichZonesToProcess(cfg.Domains, args.Domains)
	if false {
		for _, zone := range zonesToProcess {
			fmt.Printf("ZONE: %q\n", zone.Name)
			oneDomain(zone, args)
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(len(zonesToProcess))
		for _, zone := range zonesToProcess {
			fmt.Printf("ZONE: %q\n", zone.Name)
			go func(zone *models.DomainConfig, args PPreviewArgs) {
				defer wg.Done()
				oneDomain(zone, args)
			}(zone, args)
		}
		wg.Wait()
	}

	// Now we know what to do, print or do the tasks.
	fmt.Printf("CORRECTIONS:\n")
	for _, zone := range zonesToProcess {
		fmt.Printf("ZONE: %q\n", zone.Name)
		providersToProcess := whichProvidersToProcess(zone.DNSProviderInstances, args.Providers)
		for _, provider := range providersToProcess {
			fmt.Printf("    PROVIDER: %q\n", provider.Name)
			corrections := zone.GetCorrections(provider.Name)
			pprintOrRunCorrections(zone.Name, provider.Name, corrections, out, push, interactive, notifier, report)
		}
		corrections := zone.GetCorrections(zone.RegistrarInstance.Name)
		//pprintOrRunCorrections(zone.Name, zone.RegistrarInstance.Name, corrections, out, push, interactive, notifier, *report != "")
		pprintOrRunCorrections(zone.Name, zone.RegistrarInstance.Name, corrections, out, push, interactive, notifier, report)
	}
	return nil
}

func oneDomain(zone *models.DomainConfig, args PPreviewArgs) {
	// Loop over the (selected) providers configured for that zone:
	providersToProcess := whichProvidersToProcess(zone.DNSProviderInstances, args.Providers)
	for _, provider := range providersToProcess {

		// Populate the zones at the provider (if desired/needed/able):
		if !args.NoPopulate {
			populateCorrections := generatePopulateCorrections(provider, zone.Name)
			zone.StoreCorrections(provider.Name, populateCorrections)
		}

		// Update the zone's records at the provider:
		zoneCor, rep := generateZoneCorrections(zone, provider)
		zone.StoreCorrections(provider.Name, rep)
		zone.StoreCorrections(provider.Name, zoneCor)
	}

	// // Fix the parent zone's delegation: (if able/needed)
	// if parentSupportsDelegations(zone.RegistrarInstance) {
	delegationCorrections := generateDelegationCorrections(zone)
	zone.StoreCorrections(zone.RegistrarInstance.Name, delegationCorrections)
	// }
}

func whichZonesToProcess(domains []*models.DomainConfig, filter string) []*models.DomainConfig {
	if filter == "" || filter == "all" {
		return domains
	}

	permitList := strings.Split(filter, ",")
	var picked []*models.DomainConfig
	for _, domain := range domains {
		if domainInList(domain.Name, permitList) {
			picked = append(picked, domain)
		}
	}
	return picked
}

func whichProvidersToProcess(providers []*models.DNSProviderInstance, filter string) []*models.DNSProviderInstance {

	if filter == "all" { // all
		return providers
	}

	permitList := strings.Split(filter, ",")
	var picked []*models.DNSProviderInstance

	// Just the default providers:
	if filter == "" {
		for _, provider := range providers {
			if provider.IsDefault {
				picked = append(picked, provider)
			}
		}
		return picked
	}

	// Just the exact matches:
	for _, provider := range providers {
		for _, filterItem := range permitList {
			if provider.Name == filterItem {
				picked = append(picked, provider)
			}
		}
	}
	return picked
}

func generatePopulateCorrections(provider *models.DNSProviderInstance, zoneName string) []*models.Correction {

	lister, ok := provider.Driver.(providers.ZoneLister)
	if !ok {
		return nil // We can't generate a list. No corrections are possible.
	}

	zones, err := lister.ListZones()
	if err != nil {
		return []*models.Correction{{Msg: fmt.Sprintf("Provider %q ListZones returned: %s", provider.Name, err)}}
	}

	aceZoneName, _ := idna.ToASCII(zoneName)
	if slices.Contains(zones, aceZoneName) {
		return nil // zone exists. Nothing to do.
	}

	creator, ok := provider.Driver.(providers.ZoneCreator)
	if !ok {
		return []*models.Correction{{Msg: fmt.Sprintf("Zone %q does not exist. Can not create because %q does not implement ZoneCreator", aceZoneName, provider.Name)}}
	}

	return []*models.Correction{{
		Msg: fmt.Sprintf("Create zone '%s' in the '%s' profile", aceZoneName, provider.Name),
		F:   func() error { return creator.EnsureZoneExists(aceZoneName) },
	}}
}

func generateZoneCorrections(zone *models.DomainConfig, provider *models.DNSProviderInstance) ([]*models.Correction, []*models.Correction) {
	reports, zoneCorrections, err := zonerecs.CorrectZoneRecords(provider.Driver, zone)
	if err != nil {
		return []*models.Correction{{Msg: fmt.Sprintf("Domain %q provider %s Error: %s", zone.Name, provider.Name, err)}}, nil
	}
	return zoneCorrections, reports
}

func generateDelegationCorrections(zone *models.DomainConfig) []*models.Correction {
	if len(zone.Nameservers) == 0 && zone.Metadata["no_ns"] != "true" {
		return []*models.Correction{{Msg: fmt.Sprintf("No nameservers declared for domain %q; skipping registrar. Add {no_ns:'true'} to force", zone.Name)}}
	}

	corrections, err := zone.RegistrarInstance.Driver.GetRegistrarCorrections(zone)
	if err != nil {
		return msg(fmt.Sprintf("zone %q; registrar %q; Error: %s", zone.Name, zone.RegistrarInstance.Name, err))
	}
	return corrections
}

func pprintOrRunCorrections(zoneName string, providerName string, corrections []*models.Correction, out printer.CLI, push bool, interactive bool, notifier notifications.Notifier, report string) bool {
	if len(corrections) == 0 {
		return false
	}
	var anyErrors bool
	for _, correction := range corrections {
		// 		out.PrintCorrection(i, correction)
		var err error
		if push {
			if interactive && !out.PromptToRun() {
				continue
			}
			if correction.F != nil {
				err = correction.F()
				// 				out.EndCorrection(err)
				if err != nil {
					anyErrors = true
				}
			}
		}
		notifier.Notify(zoneName, providerName, correction.Msg, err, !push)
	}

	_ = report // File name to write report to.
	return anyErrors
}

func msg(s string) []*models.Correction {
	return []*models.Correction{{Msg: s}}
}

// includedZones, _ := slices.Filter(cfg.Domains, func(d *models.DomainConfig) (bool, error) { return args.shouldRunDomain(d.GetUniqueName()), nil })
// // TODO(tlim): Improve performance by rewriting shouldRunDomain to not split on comma for every run.
// fmt.Printf("len(includedZones) = %d\n", len(includedZones))

// for _, zone := range includedZones {
// 	fmt.Printf("zone: %s\n", zone.Name)

// 	// REGISTRAR CORRECTIONS
// 	fmt.Printf("    registrar = %s run=%v\n", zone.RegistrarName, run)

// 	var regCorrections []*models.Correction
// 	if args.shouldRunProvider(zone.RegistrarName, zone) {
// 		regCorrections, err = zone.RegistrarInstance.Driver.GetRegistrarCorrections(zone)
// 		fmt.Printf("    len(regCorrections) = %d\n", len(regCorrections))
// 		if err != nil {
// 			//anyErrors = true
// 		}
// 	}

// 	// DSP CORRECTIONS
// 	for _, provider := range zone.DNSProviderInstances {

// 		canListZones, zones, err := getZoneList(provider.Driver)
// 		fmt.Printf("    canListZones=%v err=%v zones=%v\n", canListZones, err, zones)

// 		if canListZones {
// 			if slices.Contains(zones, zone) {
// 				if creator, ok := provider.Driver.(providers.ZoneCreator); ok {
// 					if push {
// 						fmt.Printf("    PUSH creating zone %q in %q: %v\n", zone.Name, provider.Name, creator)
// 					} else {
// 						fmt.Printf("    preview creating zone %q in %q\n", zone.Name, provider.Name)
// 					}
// 				}
// 			}
// 		}
// 	}
// 	//aceZoneName, _ := idna.ToASCII(domain.Name)
// }

// for _, domain := range cfg.Domains {
// 	uniquename := domain.GetUniqueName()
// 	if !args.shouldRunDomain(uniquename) {
// 		skip
// 	}
// }

// 	// Run preview or push operations per domain as anonymous function, in preparation for the later use of goroutines.
// 	// For now running this code is still sequential.
// 	// Please note that at the end of this anonymous function there is a } (domain) which executes this function actually
// 	func(domain *models.DomainConfig) {
// 		defer wg.Done() // defer notify WaitGroup this anonymous function has finished

// 		err = domain.Punycode()
// 		if err != nil {
// 			return
// 		}

// 		// Correct the domain...

// 		out.StartDomain(uniquename)
// 		var providersWithExistingZone []*models.DNSProviderInstance
// 		/// For each DSP...
// 		for _, provider := range domain.DNSProviderInstances {
// 			if !args.NoPopulate {
// 				// preview run: check if zone is already there, if not print a warning
// 				if lister, ok := provider.Driver.(providers.ZoneLister); ok && !push {
// 					zones, err := lister.ListZones()
// 					if err != nil {
// 						out.Errorf("ERROR: %s\n", err.Error())
// 						return
// 					}
// 					aceZoneName, _ := idna.ToASCII(domain.Name)

// 					if !slices.Contains(zones, aceZoneName) {
// 						//out.Warnf("DEBUG: zones: %v\n", zones)
// 						//out.Warnf("DEBUG: Name: %v\n", domain.Name)

// 						out.Warnf("Zone '%s' does not exist in the '%s' profile and will be added automatically.\n", domain.Name, provider.Name)
// 						continue // continue with next provider, as we can not determine corrections without an existing zone
// 					}
// 				} else if creator, ok := provider.Driver.(providers.ZoneCreator); ok && push {
// 					// this is the actual push, ensure domain exists at DSP
// 					if err := creator.EnsureZoneExists(domain.Name); err != nil {
// 						out.Warnf("Error creating domain: %s\n", err)
// 						anyErrors = true
// 						continue // continue with next provider, as we couldn't create this one
// 					}
// 				}
// 			}
// 			providersWithExistingZone = append(providersWithExistingZone, provider)
// 		}

// 		// Correct the registrar...

// 		nsList, err := nameservers.DetermineNameserversForProviders(domain, providersWithExistingZone)
// 		if err != nil {
// 			out.Errorf("ERROR: %s\n", err.Error())
// 			return
// 		}
// 		domain.Nameservers = nsList
// 		nameservers.AddNSRecords(domain)

// 		for _, provider := range providersWithExistingZone {

// 			shouldrun := args.shouldRunProvider(provider.Name, domain)
// 			out.StartDNSProvider(provider.Name, !shouldrun)
// 			if !shouldrun {
// 				continue
// 			}

// 			reports, corrections, err := zonerecs.CorrectZoneRecords(provider.Driver, domain)
// 			out.EndProvider(provider.Name, len(corrections), err)
// 			if err != nil {
// 				anyErrors = true
// 				return
// 			}
// 			totalCorrections += len(corrections)
// 			pprintReports(domain.Name, provider.Name, reports, out, push, notifier)
// 			reportItems = append(reportItems, ReportItem{
// 				Domain:      domain.Name,
// 				Corrections: len(corrections),
// 				Provider:    provider.Name,
// 			})
// 			anyErrors = pprintOrRunCorrections(domain.Name, provider.Name, corrections, out, push, interactive, notifier) || anyErrors
// 		}

// 		//
// 		run := args.shouldRunProvider(domain.RegistrarName, domain)
// 		out.StartRegistrar(domain.RegistrarName, !run)
// 		if !run {
// 			return
// 		}
// 		if len(domain.Nameservers) == 0 && domain.Metadata["no_ns"] != "true" {
// 			out.Warnf("No nameservers declared; skipping registrar. Add {no_ns:'true'} to force.\n")
// 			return
// 		}

// 		totalCorrections += len(corrections)
// 		reportItems = append(reportItems, ReportItem{
// 			Domain:      domain.Name,
// 			Corrections: len(corrections),
// 			Registrar:   domain.RegistrarName,
// 		})
// 		anyErrors = pprintOrRunCorrections(domain.Name, domain.RegistrarName, corrections, out, push, interactive, notifier) || anyErrors
// 	}(domain)
// }
// wg.Wait() // wait for all anonymous functions to finish

// if os.Getenv("TEAMCITY_VERSION") != "" {
// 	fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
// }
// notifier.Done()
// out.Printf("Done. %d corrections.\n", totalCorrections)
// if anyErrors {
// 	return fmt.Errorf("completed with errors")
// }
// if totalCorrections != 0 && args.WarnChanges {
// 	return fmt.Errorf("there are pending changes")
// }
// if report != nil && *report != "" {
// 	f, err := os.OpenFile(*report, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()
// 	b, err := json.MarshalIndent(reportItems, "", "  ")
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := f.Write(b); err != nil {
// 		return err
// 	}
// }

// func getZoneList(driver models.DNSProvider) (bool, []string, error) {
// 	lister, ok := driver.(providers.ZoneLister)
// 	if !ok {
// 		return false, nil, nil
// 	}
// 	zones, err := lister.ListZones()
// 	return true, zones, err
// }

// func zoneMissing(name string, zones []string) bool {
// 	return false
// }

// PInitializeProviders takes (fully processed) configuration and instantiates all providers and returns them.
func PInitializeProviders(cfg *models.DNSConfig, providerConfigs map[string]map[string]string, notifyFlag bool) (notify notifications.Notifier, err error) {
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
	msgs, err := ppopulateProviderTypes(cfg, providerConfigs)
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

// pproviderTypeFieldName is the name of the field in creds.json that specifies the provider type id.
const pproviderTypeFieldName = "TYPE"

// ppurl is the documentation URL to list in the warnings related to missing provider type ids.
const purl = "https://docs.dnscontrol.org/commands/creds-json"

// ppopulateProviderTypes scans a DNSConfig for blank provider types and fills them in based on providerConfigs.
// That is, if the provider type is "-" or "", we take that as an flag
// that means this value should be replaced by the type found in creds.json.
func ppopulateProviderTypes(cfg *models.DNSConfig, providerConfigs map[string]map[string]string) ([]string, error) {
	var msgs []string

	for i := range cfg.Registrars {
		pType := cfg.Registrars[i].Type
		pName := cfg.Registrars[i].Name
		nt, warnMsg, err := prefineProviderType(pName, pType, providerConfigs[pName], "NewRegistrar")
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
		nt, warnMsg, err := prefineProviderType(pName, pType, providerConfigs[pName], "NewDnsProvider")
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
			nt, warnMsg, err := prefineProviderType(pName, pType, providerConfigs[pName], "NewDnsProvider")
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
		nt, warnMsg, err := prefineProviderType(pName, pType, providerConfigs[pName], "NewRegistrar")
		p.ProviderType = nt
		if warnMsg != "" {
			msgs = append(msgs, warnMsg)
		}
		if err != nil {
			return msgs, err
		}
	}

	return puniqueStrings(msgs), nil
}

// puniqueStrings takes an unsorted slice of strings and returns the
// unique strings, in the order they first appeared in the list.
func puniqueStrings(stringSlice []string) []string {
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

func prefineProviderType(credEntryName string, t string, credFields map[string]string, source string) (replacementType string, warnMsg string, err error) {

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
				credEntryName, pproviderTypeFieldName, t,
				purl,
			), nil
		}

		switch ct := credFields[pproviderTypeFieldName]; ct {
		case "":
			// Warn the user to update creds.json in preparation for 4.0:
			// In 4.0 this should be an error.
			return t, fmt.Sprintf(`WARNING: For future compatibility, update the %q entry in creds.json by adding: %q: %q, (See %s#missing)`,
				credEntryName,
				pproviderTypeFieldName, t,
				purl,
			), nil
		case "-":
			// This should never happen. The user is specifying "-" in a place that it shouldn't be used.
			return "-", "", fmt.Errorf(`ERROR: creds.json entry %q has invalid %q value %q (See %s#hyphen)`,
				credEntryName, pproviderTypeFieldName, ct,
				purl,
			)
		case t:
			// creds.json file is compatible with and dnsconfig.js can be updated.
			return ct, fmt.Sprintf(`INFO: In dnsconfig.js %s(%q, %q) can be simplified to %s(%q) (See %s#cleanup)`,
				source, credEntryName, t,
				source, credEntryName,
				purl,
			), nil
		default:
			// creds.json lists a TYPE but it doesn't match what's in dnsconfig.js!
			return t, "", fmt.Errorf(`ERROR: Mismatch found! creds.json entry %q has %q set to %q but dnsconfig.js specifies %s(%q, %q) (See %s#mismatch)`,
				credEntryName,
				pproviderTypeFieldName, ct,
				source, credEntryName, t,
				purl,
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
			credEntryName, pproviderTypeFieldName, "FILL_IN_PROVIDER_TYPE",
			purl,
		)
	}

	// New-style, dnsconfig.js doesn't specifies the type. It will be
	// looked up in creds.json.
	switch ct := credFields[pproviderTypeFieldName]; ct {
	case "":
		return ct, "", fmt.Errorf(`ERROR: creds.json entry %q is missing: %q: %q, (See %s#fixcreds)`,
			credEntryName,
			pproviderTypeFieldName, "FILL_IN_PROVIDER_TYPE",
			purl,
		)
	case "-":
		// This should never happen. The user is confused and specified "-" in the wrong place!
		return "-", "", fmt.Errorf(`ERROR: creds.json entry %q has invalid %q value %q (See %s#hyphen)`,
			credEntryName,
			pproviderTypeFieldName, ct,
			purl,
		)
	default:
		// use the value in creds.json (this should be the normal case)
		return ct, "", nil
	}

}

// func pprintReports(domain string, provider string, reports []*models.Correction, out printer.CLI, push bool, notifier notifications.Notifier) (anyErrors bool) {
// 	anyErrors = false
// 	if len(reports) == 0 {
// 		return false
// 	}
// 	for i, report := range reports {
// 		out.PrintReport(i, report)
// 		notifier.Notify(domain, provider, report.Msg, nil, !push)
// 	}
// 	return anyErrors
// }
