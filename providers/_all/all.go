//Package all is simply a container to reference all known provider implementations for easy import into other packages
package all

import (
	//Define all known providers here. They should each register themselves with the providers package via init function.
	_ "github.com/StackExchange/dnscontrol/providers/activedir"
	_ "github.com/StackExchange/dnscontrol/providers/bind"
	_ "github.com/StackExchange/dnscontrol/providers/cloudflare"
	_ "github.com/StackExchange/dnscontrol/providers/gandi"
	_ "github.com/StackExchange/dnscontrol/providers/google"
	_ "github.com/StackExchange/dnscontrol/providers/namecheap"
	_ "github.com/StackExchange/dnscontrol/providers/namedotcom"
	_ "github.com/StackExchange/dnscontrol/providers/route53"
)
