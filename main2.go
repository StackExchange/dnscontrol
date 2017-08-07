package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/StackExchange/dnscontrol/models"
	"github.com/StackExchange/dnscontrol/providers"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
)

//go:generate go run build/generate/generate.go

var credsFile = flag.String("creds", "creds.json", "Provider credentials JSON file")
var jsonFile = flag.String("json", "", "File containing intermediate JSON")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	command := flag.Arg(0)
	if command == "version" {
		fmt.Println(versionString())
		return
	}

	var dnsConfig *models.DNSConfig

	switch command {
	case "create-domains":
		for _, domain := range dnsConfig.Domains {
			fmt.Println("*** ", domain.Name)
			for prov := range domain.DNSProviders {
				dsp, ok := dsps[prov]
				if !ok {
					log.Fatalf("DSP %s not declared.", prov)
				}
				if creator, ok := dsp.(providers.DomainCreator); ok {
					fmt.Println("  -", prov)
					err := creator.EnsureDomainExists(domain.Name)
					if err != nil {
						fmt.Printf("Error creating domain: %s\n", err)
					}
				}
			}
		}
	case "preview", "push":

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
