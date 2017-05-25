package main

import (
	"fmt"
	"strings"

	"github.com/StackExchange/dnscontrol/dnsresolver"
	"github.com/StackExchange/dnscontrol/spflib"
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

	fmt.Println()
	fmt.Println("---------------------")
	fmt.Println()

	res := dnsresolver.NewResolverLive("preload-dns.json")
	//res := dnsresolver.NewResolverPreloaded("preload-dns.json")

	rec, err := spflib.Parse(strings.Join([]string{"v=spf1",
		"ip4:198.252.206.0/24",
		"ip4:192.111.0.0/24",
		"include:_spf.google.com",
		"include:mailgun.org",
		"include:spf-basic.fogcreek.com",
		"include:mail.zendesk.com",
		"include:servers.mcsv.net",
		"include:sendgrid.net",
		"include:spf.mtasv.net",
		"~all"}, " "), res)
	if err != nil {
		panic(err)
	}
	spflib.DumpSPF(rec, "")

	fmt.Println()
	fmt.Println("---------------------")
	fmt.Println()

	var spfs []string
	spfs, err = spflib.Lookup("stackex.com", res)
	if err != nil {
		panic(err)
	}
	rec, err = spflib.Parse(strings.Join(spfs, " "), res)
	if err != nil {
		panic(err)
	}
	spflib.DumpSPF(rec, "")

}
