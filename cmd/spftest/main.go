package main

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/spf"
)

func main() {

	h := spf.NewResolverLive("spf-store.json")
	fmt.Println(h.GetTxt("_spf.google.com"))
	fmt.Println(h.GetTxt("spf-basic.fogcreek.com"))
	h.Close()

	i := spf.NewResolverPreloaded("spf-store.json")
	fmt.Println(i.GetTxt("_spf.google.com"))
	fmt.Println(i.GetTxt("spf-basic.fogcreek.com"))
	i.Close()

}
