package commands

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/DNSControl/dnscontrol/v4/pkg/credsfile"
	"github.com/DNSControl/dnscontrol/v4/pkg/providers"
	"github.com/urfave/cli/v3"
)

var _ = cmd(catMain, func() *cli.Command {
	var args InitArgs
	return &cli.Command{
		Name:  "init",
		Usage: "Interactively create a creds.json and starter dnsconfig.js",
		Description: "Walks you through picking a registrar and DNS provider, " +
			"entering their credentials, and writing a creds.json plus a minimal " +
			"dnsconfig.js so a fresh setup can run `dnscontrol preview` immediately.",
		Action: func(ctx context.Context, c *cli.Command) error {
			return exit(Init(args))
		},
		Flags: args.flags(),
	}
}())

// InitArgs carries the flag values for the `init` subcommand.
type InitArgs struct {
	CredsFile  string
	ConfigFile string
	SkipConfig bool
}

func (args *InitArgs) flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "creds",
			Value:       "creds.json",
			Usage:       "Output path for the credentials file",
			Destination: &args.CredsFile,
		},
		&cli.StringFlag{
			Name:        "config",
			Value:       "dnsconfig.js",
			Usage:       "Output path for the starter DNSControl config",
			Destination: &args.ConfigFile,
		},
		&cli.BoolFlag{
			Name:        "no-config",
			Value:       false,
			Usage:       "Do not write a starter dnsconfig.js",
			Destination: &args.SkipConfig,
		},
	}
}

// Init runs the interactive onboarding flow described by InitArgs.
func Init(args InitArgs) error {
	return runInit(args, surveyAsker{})
}

// runInit is the test friendly entry point. It takes an Asker so tests
// can stub the interactive prompts.
func runInit(args InitArgs, asker Asker) error {
	fmt.Println("Welcome to dnscontrol init.")
	fmt.Println("This wizard creates a creds.json and a starter dnsconfig.js.")

	existingCreds, err := loadExistingCreds(args.CredsFile)
	if err != nil {
		return err
	}

	regType, dnsType, sameAccount, err := pickProviders(asker)
	if err != nil {
		return err
	}

	entries, choice, err := collectEntries(asker, regType, dnsType, sameAccount)
	if err != nil {
		return err
	}

	if !args.SkipConfig {
		choice.Domains, err = askDomains(asker)
		if err != nil {
			return err
		}
	}

	credsBytes, err := renderCredsJSON(existingCreds, entries)
	if err != nil {
		return err
	}
	var configBytes []byte
	if !args.SkipConfig {
		configBytes = renderDnsconfigJS(choice)
	}

	if err := confirmAndWrite(asker, args, existingCreds, entries, credsBytes, configBytes); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Done.")
	return offerFollowUps(asker, args, entries, choice)
}

// pickProviders walks the user through choosing a DNS provider and a
// registrar. It returns the chosen registrar TYPE, DNS provider TYPE,
// and whether the registrar should reuse the DNS provider's credentials.
func pickProviders(asker Asker) (regType, dnsType string, sameAccount bool, err error) {
	// DNS first because most users think in terms of where their records
	// live. NONE defers the choice. The picker only lists providers
	// whose maintainers have registered onboarding metadata so the
	// wizard can drive the prompts. Other providers should be set up
	// from the documentation.
	dnsOptions := providersWithMetadata(keysOf(providers.DNSProviderTypes))
	dnsOptions = append([]string{"NONE"}, dnsOptions...)
	fmt.Println()
	fmt.Println("A DNS provider hosts the records (A, MX, TXT, CNAME, and so on) for your zones.")
	fmt.Println("Pick NONE if you want to defer this choice.")
	fmt.Println("Providers not listed below can be configured from their documentation page at https://docs.dnscontrol.org/provider/.")
	dnsType, err = pickProvider(asker, "Which DNS service provider do you want to configure?", dnsOptions)
	if err != nil {
		return "", "", false, err
	}

	// If the chosen DNS provider can also act as a registrar, offer to
	// reuse it; otherwise ask which registrar to use, with NONE as the
	// default.
	if dnsType != "NONE" {
		if _, alsoRegistrar := providers.RegistrarTypes[dnsType]; alsoRegistrar {
			meta, _ := providers.GetCredsMetadata(dnsType)
			sameAccount, err = asker.Confirm(
				fmt.Sprintf("Use the same %s account for the registrar role too?", displayName(meta.TypeName)),
				true,
			)
			if err != nil {
				return "", "", false, err
			}
			if sameAccount {
				return dnsType, dnsType, true, nil
			}
		}
	}

	fmt.Println()
	fmt.Println("A registrar is where the domain itself is registered. DNSControl updates the NS delegation there.")
	fmt.Println("Pick NONE if you manage the registrar outside DNSControl.")
	fmt.Println("Registrars not listed below can be configured from their documentation page at https://docs.dnscontrol.org/provider/.")
	regType, err = pickProvider(asker, "Which registrar do you want to configure?",
		providersWithMetadata(keysOf(providers.RegistrarTypes)))
	if err != nil {
		return "", "", false, err
	}
	return regType, dnsType, false, nil
}

// confirmAndWrite shows the rendered files, warns the user about any
// pre existing files that will be merged or replaced, asks for
// confirmation, and writes the files when accepted. It also runs the
// per provider PostWrite hooks and validates that the resulting
// creds.json still parses.
func confirmAndWrite(asker Asker, args InitArgs, existingCreds map[string]map[string]string, entries []InitCredsEntry, credsBytes, configBytes []byte) error {
	fmt.Println()
	fmt.Printf("--- %s ---\n", args.CredsFile)
	fmt.Println(string(credsBytes))
	if !args.SkipConfig {
		fmt.Printf("--- %s ---\n", args.ConfigFile)
		fmt.Println(string(configBytes))
	}

	credsExists := len(existingCreds) > 0
	configExists := false
	if !args.SkipConfig {
		if _, err := os.Stat(args.ConfigFile); err == nil {
			configExists = true
		}
	}
	if credsExists || configExists {
		fmt.Println()
		if credsExists {
			fmt.Printf("NOTE: %s already exists; new entries are merged in.\n", args.CredsFile)
		}
		if configExists {
			fmt.Printf("NOTE: %s already exists and will be replaced.\n", args.ConfigFile)
		}
	}

	confirm, err := asker.Confirm("Write these files?", true)
	if err != nil {
		return err
	}
	if !confirm {
		return errInitAborted
	}

	if err := writeFile(args.CredsFile, credsBytes); err != nil {
		return err
	}
	if !args.SkipConfig {
		if err := writeFile(args.ConfigFile, configBytes); err != nil {
			return err
		}
	}
	runPostWriteHooks(entries)

	// Round trip: confirm credsfile can still load the result.
	if _, err := credsfile.LoadProviderConfigs(args.CredsFile); err != nil {
		return fmt.Errorf("wrote %s but it failed to parse: %w", args.CredsFile, err)
	}
	return nil
}

// offerFollowUps asks the user whether to compare configured domains
// with zones at the provider and whether to run `dnscontrol preview`
// immediately.
func offerFollowUps(asker Asker, args InitArgs, entries []InitCredsEntry, choice InitDnsconfigChoice) error {
	binary := dnscontrolBinary()

	if sample, ok := dnsSample(entries); ok && len(choice.Domains) > 0 {
		prompt := fmt.Sprintf("Compare domains in dnsconfig.js with zones at %s?", displayName(sample.TypeName))
		compare, err := asker.Confirm(prompt, true)
		if err != nil {
			return err
		}
		if compare {
			// get-zones writes its own diagnostics to stderr, so a
			// non nil error here adds no information beyond what the
			// user already saw. Keep going.
			_ = compareZones(binary, args, sample, choice.Domains)
		}
	}

	run, err := asker.Confirm("Run `dnscontrol preview` now?", true)
	if err != nil {
		return err
	}
	if run {
		cmd := exec.Command(binary, buildPreviewArgs(args)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := runSubcommand(cmd); err != nil {
			fmt.Fprintf(os.Stderr, "dnscontrol preview failed: %v\n", err)
		}
	}

	printCommunityWelcome()
	return nil
}

// printCommunityWelcome closes the init flow with a pointer to the
// GitHub community and the online documentation.
func printCommunityWelcome() {
	fmt.Println()
	fmt.Println("================================================================================")
	fmt.Println()
	fmt.Println("Welcome to the DNSControl community!")
	fmt.Println()
	fmt.Println("Questions, feedback or ideas are always welcome:")
	fmt.Println("  Discussions: https://github.com/StackExchange/dnscontrol/discussions")
	fmt.Println("  Issues:      https://github.com/StackExchange/dnscontrol/issues")
	fmt.Println()
	fmt.Println("Learn more:")
	fmt.Println("  Getting started: https://docs.dnscontrol.org/getting-started/getting-started")
	fmt.Println("  Examples:        https://docs.dnscontrol.org/getting-started/examples")
	fmt.Println("  Migrating zones: https://docs.dnscontrol.org/getting-started/migrating")
	fmt.Println()
	fmt.Println("Want to stay up to date? Releases and the monthly DNSControl community video call")
	fmt.Println("are announced at https://github.com/StackExchange/dnscontrol/discussions/categories/announcements")
}

// compareZones fetches the live zone list from the provider via
// fetchZoneNames and prints how it lines up with the domains the user
// just placed in dnsconfig.js.
func compareZones(binary string, args InitArgs, sample InitCredsEntry, configured []string) error {
	zones, err := fetchZoneNames(binary, args, sample)
	if err != nil {
		return err
	}
	both, onlyConfig, onlyProvider := diffDomains(configured, zones)
	fmt.Println()
	fmt.Printf("Zones at %s compared with dnsconfig.js:\n", displayName(sample.TypeName))
	fmt.Printf("  In both          : %s\n", formatList(both))
	fmt.Printf("  Only in config   : %s\n", formatList(onlyConfig))
	fmt.Printf("  Only at provider : %s\n", formatList(onlyProvider))
	return nil
}

// fetchZoneNames invokes `dnscontrol get-zones --format=nameonly` as a
// subprocess and returns the zones printed on stdout. Stderr is
// streamed straight through so errors land in front of the user.
func fetchZoneNames(binary string, args InitArgs, sample InitCredsEntry) ([]string, error) {
	argv := []string{"get-zones", "--format=nameonly"}
	if args.CredsFile != "creds.json" {
		argv = append(argv, "--creds", args.CredsFile)
	}
	argv = append(argv, "--", sample.Name, "-", "all")

	cmd := exec.Command(binary, argv...)
	var stdout bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := runSubcommand(cmd); err != nil {
		return nil, err
	}
	return parseNameOnlyOutput(stdout.String()), nil
}

// parseNameOnlyOutput extracts zone names from the `get-zones
// --format=nameonly` stdout. Empty lines and surrounding whitespace are
// ignored; CRLF line endings are handled. Lines that do not look like
// zone names (containing whitespace) are skipped to be safe against
// future format additions.
func parseNameOnlyOutput(output string) []string {
	output = strings.ReplaceAll(output, "\r\n", "\n")
	var zones []string
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.ContainsAny(line, " \t") {
			continue
		}
		zones = append(zones, line)
	}
	return zones
}

// diffDomains partitions the configured and provider sets.
func diffDomains(configured, atProvider []string) (both, onlyConfig, onlyProvider []string) {
	toSet := func(domains []string) map[string]bool {
		set := make(map[string]bool, len(domains))
		for _, domain := range domains {
			set[strings.ToLower(domain)] = true
		}
		return set
	}
	configuredSet := toSet(configured)
	providerSet := toSet(atProvider)
	for _, name := range configured {
		if providerSet[strings.ToLower(name)] {
			both = append(both, name)
		} else {
			onlyConfig = append(onlyConfig, name)
		}
	}
	for _, name := range atProvider {
		if !configuredSet[strings.ToLower(name)] {
			onlyProvider = append(onlyProvider, name)
		}
	}
	sort.Strings(both)
	sort.Strings(onlyConfig)
	sort.Strings(onlyProvider)
	return
}

// formatList returns a comma separated list, or "(none)" for an empty
// slice.
func formatList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}
	return strings.Join(items, ", ")
}

// dnsSample picks the first non NONE entry. Used to suggest a
// `get-zones` command.
func dnsSample(entries []InitCredsEntry) (InitCredsEntry, bool) {
	for _, entry := range entries {
		if entry.TypeName != "" && entry.TypeName != "NONE" {
			return entry, true
		}
	}
	return InitCredsEntry{}, false
}

// buildPreviewArgs constructs the argv for `dnscontrol preview`,
// forwarding non default creds and config paths.
func buildPreviewArgs(args InitArgs) []string {
	argv := []string{"preview"}
	if args.CredsFile != "creds.json" {
		argv = append(argv, "--creds", args.CredsFile)
	}
	if !args.SkipConfig && args.ConfigFile != "dnsconfig.js" {
		argv = append(argv, "--config", args.ConfigFile)
	}
	return argv
}

// runSubcommand executes the prepared *exec.Cmd. Callers attach the
// stdio redirection (or a capture buffer) they need. It is exposed as a
// var so tests can replace it without spawning real subprocesses.
var runSubcommand = func(cmd *exec.Cmd) error {
	fmt.Printf("\n$ %s\n", strings.Join(cmd.Args, " "))
	return cmd.Run()
}

// dnscontrolBinary resolves the path of the currently running
// dnscontrol binary. os.Executable handles symlinks and `go run`
// correctly; a fallback to os.Args[0] is used when the syscall is not
// implemented (rare).
func dnscontrolBinary() string {
	if path, err := os.Executable(); err == nil {
		return path
	}
	return os.Args[0]
}

// collectEntries prompts for credentials for the chosen provider types and
// returns the entries plus the dnsconfig.js choice record. DNS is
// collected first because that is the primary workflow for most users.
func collectEntries(asker Asker, regType, dnsType string, sameAccount bool) ([]InitCredsEntry, InitDnsconfigChoice, error) {
	var entries []InitCredsEntry
	choice := InitDnsconfigChoice{}

	dnsEntryName := ""
	if dnsType != "NONE" && dnsType != "" {
		meta, ok := providers.GetCredsMetadata(dnsType)
		if !ok {
			meta = providers.CredsMetadata{TypeName: dnsType, DisplayName: dnsType}
		}
		fmt.Printf("\n== DNS provider: %s ==\n", displayName(meta.TypeName))
		fields, name, err := askEntry(asker, meta, defaultEntryName(dnsType))
		if err != nil {
			return nil, choice, err
		}
		entries = append(entries, InitCredsEntry{
			Name:     name,
			TypeName: dnsType,
			Fields:   fields,
		})
		dnsEntryName = name
		choice.DNSName = name
		choice.DNSVar = jsVarName("DNS", dnsType)
	}

	if regType == "" {
		return entries, choice, nil
	}

	if sameAccount && regType == dnsType && dnsEntryName != "" {
		// Reuse the DNS entry as the registrar.
		choice.RegistrarName = dnsEntryName
		choice.RegistrarVar = jsVarName("REG", regType)
		return entries, choice, nil
	}

	meta, ok := providers.GetCredsMetadata(regType)
	if !ok {
		meta = providers.CredsMetadata{TypeName: regType, DisplayName: regType}
	}
	fmt.Printf("\n== Registrar: %s ==\n", displayName(meta.TypeName))
	fields, name, err := askEntry(asker, meta, defaultEntryName(regType))
	if err != nil {
		return nil, choice, err
	}
	if regType != "NONE" || len(fields) > 0 {
		entries = append(entries, InitCredsEntry{
			Name:     name,
			TypeName: regType,
			Fields:   fields,
		})
	} else {
		// Minimal NONE entry so dnsconfig.js can reference it.
		entries = append(entries, InitCredsEntry{
			Name:     name,
			TypeName: "NONE",
			Fields:   map[string]string{},
		})
	}
	choice.RegistrarName = name
	choice.RegistrarVar = jsVarName("REG", regType)
	return entries, choice, nil
}

// askEntry prompts for the creds.json entry key and the credential values
// for a single provider.
func askEntry(asker Asker, meta providers.CredsMetadata, defaultName string) (map[string]string, string, error) {
	if err := openPortalHint(asker, meta); err != nil {
		return nil, "", err
	}

	fields := map[string]string{}
	if len(meta.Fields) > 0 {
		var err error
		fields, err = collectFields(asker, meta)
		if err != nil {
			return nil, "", err
		}
	}

	name, err := asker.Input(
		"creds.json entry name for this provider",
		"The top level key under which this provider appears in creds.json.",
		defaultName,
	)
	if err != nil {
		return nil, "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}
	return fields, name, nil
}

// pickProvider lets the user pick from a sorted list of provider names.
func pickProvider(asker Asker, question string, options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no providers available")
	}
	return asker.Select(question, "Start typing to filter the list.", options, options[0])
}

// askDomains prompts for one or more domain names. The first domain is
// required so the starter dnsconfig.js is never written with a stub.
func askDomains(asker Asker) ([]string, error) {
	first, err := askRequiredDomain(asker, "First domain name for dnsconfig.js",
		"For example example.com. You can add more later by editing dnsconfig.js.")
	if err != nil {
		return nil, err
	}
	domains := []string{first}
	for {
		more, err := asker.Confirm("Add another domain?", false)
		if err != nil {
			return nil, err
		}
		if !more {
			return domains, nil
		}
		next, err := askRequiredDomain(asker, "Next domain name", "")
		if err != nil {
			return nil, err
		}
		domains = append(domains, next)
	}
}

// askRequiredDomain prompts for a non empty domain name and re prompts
// until one is given.
func askRequiredDomain(asker Asker, message, help string) (string, error) {
	for {
		value, err := asker.Input(message, help, "")
		if err != nil {
			return "", err
		}
		value = strings.TrimSpace(value)
		if value != "" {
			return value, nil
		}
		fmt.Fprintln(os.Stderr, "A domain name is required.")
	}
}

// loadExistingCreds reads an existing creds.json, returning an empty map if
// the file does not exist. Parse errors are fatal so we never silently drop
// a broken file.
func loadExistingCreds(path string) (map[string]map[string]string, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return map[string]map[string]string{}, nil
	} else if err != nil {
		return nil, err
	}
	return credsfile.LoadProviderConfigs(path)
}

// writeFile writes data to path. The earlier "Write these files?"
// confirmation already covers user intent, and creds.json content is
// merged on top of any existing entries, so an additional per file
// overwrite prompt would just be noise.
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

// runPostWriteHooks lets each provider prepare local resources after
// creds.json has been written (for example BIND mkdir on its zone
// directory). Hooks are best effort; failures are reported but do not
// abort the wizard.
func runPostWriteHooks(entries []InitCredsEntry) {
	for _, entry := range entries {
		meta, ok := providers.GetCredsMetadata(entry.TypeName)
		if !ok || meta.PostWrite == nil {
			continue
		}
		if err := meta.PostWrite(entry.Fields); err != nil {
			fmt.Fprintf(os.Stderr, "warning: post write hook for %s: %v\n", entry.TypeName, err)
		}
	}
}

// providersWithMetadata keeps only the provider names for which
// CredsMetadata has been registered, sorted alphabetically.
func providersWithMetadata(names []string) []string {
	withMetadata := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := providers.GetCredsMetadata(name); ok {
			withMetadata = append(withMetadata, name)
		}
	}
	sort.Strings(withMetadata)
	return withMetadata
}

// keysOf returns the keys of any string keyed map.
func keysOf[V any](source map[string]V) []string {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	return keys
}
