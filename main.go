package main

import (
	"log"

	"github.com/StackExchange/dnscontrol/cmd"
	_ "github.com/StackExchange/dnscontrol/providers/_all"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cmd.Run()
}
