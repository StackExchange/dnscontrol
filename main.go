package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/StackExchange/dnscontrol/cmd"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
)

//go:generate go run build/generate/generate.go

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cmd.Run(versionString())
}

// Version management. 2 Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build.sh, more specific build information gets put in.
// Update the number here manually each release, so at least we have a range for go-get people.
var (
	SHA       = ""
	Version   = "0.2.0"
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
