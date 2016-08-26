package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/StackExchange/dnscontrol/js"
	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/normalize"
	"github.com/StackExchange/dnscontrol/providers"
	"github.com/StackExchange/dnscontrol/providers/config"
	"github.com/StackExchange/dnscontrol/web"

	//Define all known providers here. They should each register themselves with the providers package via init function.
	_ "github.com/StackExchange/dnscontrol/providers/activedir"
	_ "github.com/StackExchange/dnscontrol/providers/bind"
	_ "github.com/StackExchange/dnscontrol/providers/cloudflare"
	_ "github.com/StackExchange/dnscontrol/providers/gandi"
	_ "github.com/StackExchange/dnscontrol/providers/namecheap"
	_ "github.com/StackExchange/dnscontrol/providers/namedotcom"
	_ "github.com/StackExchange/dnscontrol/providers/route53"
)

//go:generate esc -modtime 0 -o js/static.go -pkg js -ignore go -prefix js js
//go:generate esc -modtime 0 -o web/static.go -pkg web -include=bundle\.js -ignore node_modules -prefix web web

// One of these config options must be set.
var jsFile = flag.String("js", "dnsconfig.js", "Javascript file containing dns config")
var stdin = flag.Bool("stdin", false, "Read domain config JSON from stdin")
var jsonInput = flag.String("json", "", "Read domain config from specified JSON file.")

var jsonOutputPre = flag.String("debugrawjson", "", "Write JSON intermediate to this file pre-normalization.")
var jsonOutputPost = flag.String("debugjson", "", "During preview, write JSON intermediate to this file instead of stdout.")

var configFile = flag.String("creds", "creds.json", "Provider credentials JSON file")
var devMode = flag.Bool("dev", false, "Use helpers.js from disk instead of embedded")

var flagProviders = flag.String("providers", "", "Providers to enable (comma seperated list); default is all-but-bind. Specify 'all' for all (including bind)")
var domains = flag.String("domains", "", "Comma seperated list of domain names to include")

var interactive = flag.Bool("i", false, "Confirm or Exclude each correction before they run")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	command := flag.Arg(0)
	if command == "version" {
		fmt.Println(versionString())
		return
	}
	if command == "web" {
		runWebServer()
		return
	}

	var dnsConfig *models.DNSConfig
	if *stdin {
		log.Fatal("Read from stdin not implemented yet.")
	} else if *jsonInput != "" {
		log.Fatal("Direct JSON read not implemented")
	} else if *jsFile != "" {
		text, err := ioutil.ReadFile(*jsFile)
		if err != nil {
			log.Fatalf("Error reading %v: %v\n", *jsFile, err)
		}
		dnsConfig, err = js.ExecuteJavascript(string(text), *devMode)
		if err != nil {
			log.Fatalf("Error executing javasscript in (%v): %v", *jsFile, err)
		}
	}

	if dnsConfig == nil {
		log.Fatal("No config specified.")
	}

	if flag.NArg() != 1 {
		fmt.Println("Usage: dnscontrol [options] cmd")
		fmt.Println("        cmd:")
		fmt.Println("           preview: Show changed that would happen.")
		fmt.Println("           push:    Make changes for real.")
		fmt.Println("           version: Print program version string.")
		fmt.Println("           print:   Print compiled data.")
		fmt.Println("")
		flag.PrintDefaults()
		return
	}
	if *jsonOutputPre != "" {
		dat, _ := json.MarshalIndent(dnsConfig, "", "  ")
		err := ioutil.WriteFile(*jsonOutputPre, dat, 0644)
		if err != nil {
			panic(err)
		}
	}

	errs := normalize.NormalizeAndValidateConfig(dnsConfig)
	if len(errs) > 0 {
		fmt.Printf("%d Validation errors:\n", len(errs))
		for i, err := range errs {
			fmt.Printf("%d: %s\n", i+1, err)
		}
	}

	if command == "print" {
		dat, _ := json.MarshalIndent(dnsConfig, "", "  ")
		if *jsonOutputPost == "" {
			fmt.Println("While running JS:", string(dat))
		} else {
			err := ioutil.WriteFile(*jsonOutputPost, dat, 0644)
			if err != nil {
				panic(err)
			}
		}
		return
	}

	providerConfigs, err := config.LoadProviderConfigs(*configFile)
	if err != nil {
		log.Fatalf("error loading provider configurations: %s", err)
	}
	registrars, err := providers.CreateRegistrars(dnsConfig, providerConfigs)
	if err != nil {
		log.Fatalf("Error creating registrars: %v\n", err)
	}
	dsps, err := providers.CreateDsps(dnsConfig, providerConfigs)
	if err != nil {
		log.Fatalf("Error creating dsps: %v\n", err)
	}

	fmt.Printf("Initialized %d registrars and %d dns service providers.\n", len(registrars), len(dsps))
	anyErrors, totalCorrections := false, 0
	switch command {
	case "preview", "push":
	DomainLoop:
		for _, domain := range dnsConfig.Domains {
			if !shouldRunDomain(domain.Name) {
				continue
			}
			fmt.Printf("******************** Domain: %s\n", domain.Name)
			for pi, prov := range domain.Dsps {

				dc, err := domain.Copy()
				if err != nil {
					log.Fatal(err)
				}
				shouldrun := shouldRunProvider(prov)
				if shouldrun {
					fmt.Printf("----- DNS Provider: %s\n", prov)
				} else {
					if pi == 0 {
						fmt.Printf("----- DNS Provider: %s (read-only)\n", prov)
					} else {
						fmt.Printf("----- DNS Provider: %s (skipping)\n", prov)
					}
				}
				dsp, ok := dsps[prov]
				if !ok {
					log.Fatalf("DSP %s not declared.", prov)
				}
				corrections, err := dsp.GetDomainCorrections(dc)
				if err != nil {
					anyErrors = true
					fmt.Printf("Error getting corrections: %s\n", err)
					continue DomainLoop
				}
				storeNameservers(dc, domain)
				if !shouldrun {
					continue
				}
				totalCorrections += len(corrections)
				anyErrors = printOrRunCorrections(corrections, command) || anyErrors
			}
			if !shouldRunProvider(domain.Registrar) {
				continue
			}
			fmt.Printf("----- Registrar: %s\n", domain.Registrar)
			reg, ok := registrars[domain.Registrar]
			if !ok {
				log.Fatalf("Registrar %s not declared.", reg)
			}
			if len(domain.Nameservers) == 0 {
				//fmt.Printf("No nameservers declared; skipping registrar.\n")
				continue
			}
			dc, err := domain.Copy()
			if err != nil {
				log.Fatal(err)
			}
			corrections, err := reg.GetRegistrarCorrections(dc)
			if err != nil {
				fmt.Printf("Error getting corrections: %s\n", err)
				anyErrors = true
				continue
			}
			totalCorrections += len(corrections)
			anyErrors = printOrRunCorrections(corrections, command) || anyErrors
		}
	default:
		log.Fatalf("Unknown command %s", command)
	}
	if os.Getenv("TEAMCITY_VERSION") != "" {
		fmt.Fprintf(os.Stderr, "##teamcity[buildStatus status='SUCCESS' text='%d corrections']", totalCorrections)
	}
	fmt.Printf("Done. %d corrections.\n", totalCorrections)
	if anyErrors {
		os.Exit(1)
	}
}

var reader = bufio.NewReader(os.Stdin)

func printOrRunCorrections(corrections []*models.Correction, command string) (anyErrors bool) {
	anyErrors = false
	if len(corrections) == 0 {
		return anyErrors
	}
	for i, correction := range corrections {
		fmt.Printf("#%d: %s\n", i+1, correction.Msg)
		if command == "push" {
			if *interactive {
				fmt.Print("Run? (Y/n): ")
				txt, err := reader.ReadString('\n')
				run := true
				if err != nil {
					run = false
				}
				txt = strings.ToLower(strings.TrimSpace(txt))
				if txt != "y" {
					run = false
				}
				if !run {
					fmt.Println("Skipping")
					continue
				}
			}
			err := correction.F()
			if err != nil {
				fmt.Println("FAILURE!", err)
				anyErrors = true
			} else {
				fmt.Println("SUCCESS!")
			}
		}
	}
	return anyErrors
}

func shouldRunProvider(p string) bool {
	if *flagProviders == "all" {
		return true
	}
	if *flagProviders == "" {
		return p != "bind"
		// NOTE(tlim): Hardcoding bind is a hacky way to make it off by default.
		// As a result, bind only runs if you list it in -providers or use
		// -providers=all.
		// If you always want bind to run, call it something else in dnsconfig.js
		// for example `NewDSP('bindyes', 'BIND',`.
		// We don't want this hack, but we shouldn't need this in the future
		// so it doesn't make sense to write a lot of code to make it work.
		// In the future, the above `return p != "bind"` can become `return true`.
		// Alternatively we might want to add a complex system that permits
		// fancy whitelist/blacklisting of providers with defaults and so on.
		// In that case, all of this hack will go away.
	}
	for _, prov := range strings.Split(*flagProviders, ",") {
		if prov == p {
			return true
		}
	}
	return false
}

func shouldRunDomain(d string) bool {
	if *domains == "" {
		return true
	}
	for _, dom := range strings.Split(*domains, ",") {
		if dom == d {
			return true
		}
	}
	return false
}

func storeNameservers(from, to *models.DomainConfig) {
	if len(to.Nameservers) == 0 && len(from.Nameservers) > 0 {
		to.Nameservers = from.Nameservers
	}
}

func runWebServer() {
	fmt.Printf("Running Webserver on :8080 (js = %s , creds = %s)", *jsFile, *configFile)
	web.Serve(*jsFile, *configFile, *devMode)
}

// Version management. 2 Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build.sh, more specific build information gets put in.
// Update the number here manually each release, so at least we have a range for go-get people.
var (
	SHA       = ""
	Version   = "0.1.0"
	BuildTime = ""
)

// printVersion prints the version banner.
func versionString() string {
	var version string
	if SHA != "" {
		version = fmt.Sprintf("%s (%s)", Version, SHA)
	} else {
		version = fmt.Sprintf("%s-dev", Version) //no SHA. '0.x.y-dev' indeicates it is run form source without build script.
	}
	if BuildTime != "" {
		i, err := strconv.ParseInt(BuildTime, 10, 64)
		if err == nil {
			tm := time.Unix(i, 0)
			version += fmt.Sprintf(" built %s", tm.Format(time.RFC822))
		}
	}
	return fmt.Sprintf("dnscontrol %s", version)
}
