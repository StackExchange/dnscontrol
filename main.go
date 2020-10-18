package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/StackExchange/dnscontrol/v3/commands"
	"github.com/StackExchange/dnscontrol/v3/pkg/version"
	_ "github.com/StackExchange/dnscontrol/v3/providers/_all"
)

//go:generate go run build/generate/generate.go build/generate/featureMatrix.go

// Version management. Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build/build.go, more specific build information gets put in.
// Update the number here manually each release, so at least we have a range for go-get people.
var (
	SHA       = ""
	Version   = "3.4.2"
	BuildTime = ""
)

func main() {
	version.SHA = SHA
	version.Semver = Version
	version.BuildTime = BuildTime

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if info, ok := debug.ReadBuildInfo(); !ok && info == nil {
		fmt.Fprint(os.Stderr, "Warning: dnscontrol was built without Go modules. See https://github.com/StackExchange/dnscontrol#from-source for more information on how to build dnscontrol correctly.\n\n")
	}
	os.Exit(commands.Run("dnscontrol " + version.Banner()))
}
