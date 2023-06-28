/*
   dnsconfig.js: dnscontrol configuration file for ORGANIZATION NAME.
*/

// Providers:

var REG_NONE = NewRegistrar("none");    // No registrar.
var DNS_BIND = NewDnsProvider("bind");  // ISC BIND.

// Domains:

D("example.com", REG_NONE, DnsProvider(DNS_BIND),
    A("@", "1.2.3.4")
);
