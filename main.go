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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if info, ok := debug.ReadBuildInfo(); !ok && info == nil {
		fmt.Fprint(os.Stderr, "Warning: dnscontrol was built without Go modules. See https://github.com/StackExchange/dnscontrol#from-source for more information on how to build dnscontrol correctly.\n\n")
	}
	os.Exit(commands.Run("dnscontrol " + version.VersionString()))
}
