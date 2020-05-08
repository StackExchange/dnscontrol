// Package all is simply a container to reference all known provider implementations for easy import into other packages
package all

import (
	// Define all known providers here. They should each register themselves with the providers package via init function.
	_ "github.com/StackExchange/dnscontrol/v3/providers/activedir"
	_ "github.com/StackExchange/dnscontrol/v3/providers/axfrddns"
	_ "github.com/StackExchange/dnscontrol/v3/providers/azuredns"
	_ "github.com/StackExchange/dnscontrol/v3/providers/bind"
	_ "github.com/StackExchange/dnscontrol/v3/providers/cloudflare"
	_ "github.com/StackExchange/dnscontrol/v3/providers/cloudns"
	_ "github.com/StackExchange/dnscontrol/v3/providers/desec"
	_ "github.com/StackExchange/dnscontrol/v3/providers/digitalocean"
	_ "github.com/StackExchange/dnscontrol/v3/providers/dnsimple"
	_ "github.com/StackExchange/dnscontrol/v3/providers/exoscale"
	_ "github.com/StackExchange/dnscontrol/v3/providers/gandi_v5"
	_ "github.com/StackExchange/dnscontrol/v3/providers/gcloud"
	_ "github.com/StackExchange/dnscontrol/v3/providers/hexonet"
	_ "github.com/StackExchange/dnscontrol/v3/providers/internetbs"
	_ "github.com/StackExchange/dnscontrol/v3/providers/linode"
	_ "github.com/StackExchange/dnscontrol/v3/providers/namecheap"
	_ "github.com/StackExchange/dnscontrol/v3/providers/namedotcom"
	_ "github.com/StackExchange/dnscontrol/v3/providers/netcup"
	_ "github.com/StackExchange/dnscontrol/v3/providers/ns1"
	_ "github.com/StackExchange/dnscontrol/v3/providers/octodns"
	_ "github.com/StackExchange/dnscontrol/v3/providers/opensrs"
	_ "github.com/StackExchange/dnscontrol/v3/providers/ovh"
	_ "github.com/StackExchange/dnscontrol/v3/providers/route53"
	_ "github.com/StackExchange/dnscontrol/v3/providers/softlayer"
	_ "github.com/StackExchange/dnscontrol/v3/providers/vultr"
)
