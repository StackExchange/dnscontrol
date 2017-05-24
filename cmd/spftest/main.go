package main

import (
	"fmt"

	"github.com/StackExchange/dnscontrol/dnsresolver"
)

func main() {

	h := dnsresolver.NewResolverLive("spf-store.json")
	fmt.Println(h.GetTxt("_spf.google.com"))
	fmt.Println(h.GetTxt("spf-basic.fogcreek.com"))
	h.Close()

	i, err := dnsresolver.NewResolverPreloaded("spf-store.json")
	if err != nil {
		panic(err)
	}
	fmt.Println(i.GetTxt("_spf.google.com"))
	fmt.Println(i.GetTxt("spf-basic.fogcreek.com"))
	fmt.Println(i.GetTxt("wontbefound"))

}
