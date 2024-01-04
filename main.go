package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/StackExchange/dnscontrol/v4/commands"
	_ "github.com/StackExchange/dnscontrol/v4/providers/_all"
	"github.com/fatih/color"
)

//go:generate go run build/generate/generate.go build/generate/featureMatrix.go build/generate/functionTypes.go build/generate/dtsFile.go

// Version management. Goals:
// 1. Someone who just does "go get" has at least some information.
// 2. If built with build/build.go, more specific build information gets put in.
// GoReleaser: version
var (
	version = "dev"
)

func main() {
	if os.Getenv("CI") == "true" {
		color.NoColor = false
	}
	if info, ok := debug.ReadBuildInfo(); !ok && info == nil {
		fmt.Fprint(os.Stderr, "Warning: dnscontrol was built without Go modules. See https://docs.dnscontrol.org/getting-started/getting-started#source for more information on how to build dnscontrol correctly.\n\n")
	}
	os.Exit(commands.Run("DNSControl " + version))
}
