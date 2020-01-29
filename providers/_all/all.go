// Package all is simply a container to reference all known provider implementations for easy import into other packages
package all

import (
	// Define all known providers here. They should each register themselves with the providers package via init function.
	_ "github.com/StackExchange/dnscontrol/v2/providers/activedir"
	_ "github.com/StackExchange/dnscontrol/v2/providers/azuredns"
	_ "github.com/StackExchange/dnscontrol/v2/providers/bind"
	_ "github.com/StackExchange/dnscontrol/v2/providers/cloudflare"
	_ "github.com/StackExchange/dnscontrol/v2/providers/cloudns"
	_ "github.com/StackExchange/dnscontrol/v2/providers/digitalocean"
	_ "github.com/StackExchange/dnscontrol/v2/providers/dnsimple"
	_ "github.com/StackExchange/dnscontrol/v2/providers/exoscale"
	_ "github.com/StackExchange/dnscontrol/v2/providers/gandi"
	_ "github.com/StackExchange/dnscontrol/v2/providers/gandi_v5"
	_ "github.com/StackExchange/dnscontrol/v2/providers/gcloud"
	_ "github.com/StackExchange/dnscontrol/v2/providers/hexonet"
	_ "github.com/StackExchange/dnscontrol/v2/providers/internetbs"
	_ "github.com/StackExchange/dnscontrol/v2/providers/linode"
	_ "github.com/StackExchange/dnscontrol/v2/providers/namecheap"
	_ "github.com/StackExchange/dnscontrol/v2/providers/namedotcom"
	_ "github.com/StackExchange/dnscontrol/v2/providers/ns1"
	_ "github.com/StackExchange/dnscontrol/v2/providers/octodns"
	_ "github.com/StackExchange/dnscontrol/v2/providers/opensrs"
	_ "github.com/StackExchange/dnscontrol/v2/providers/ovh"
	_ "github.com/StackExchange/dnscontrol/v2/providers/route53"
	_ "github.com/StackExchange/dnscontrol/v2/providers/softlayer"
	_ "github.com/StackExchange/dnscontrol/v2/providers/vultr"
)
