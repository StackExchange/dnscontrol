var REGISTRAR = NewRegistrar('none', 'NONE');    // No registrar.
var BIND = NewDnsProvider('bind', 'BIND');

// Delegating reverse zones
D(REV('1.3.0.0/16'), REGISTRAR, DnsProvider(BIND),
	NS(REV('1.3.1.0/24'), "ns1.example.com.")
);
D_EXTEND(REV("1.3.2.0/24"), NS(REV("1.3.2.0/24"), "ns2.example.org."))

// Expeccted zone: 3.1.in-addr.arpa.zone
// $TTL 300
// ; generated with dnscontrol 2020-11-30T12:58:47-05:00
// @                IN SOA   DEFAULT_NOT_SET. DEFAULT_NOT_SET. 2020113000 3600 600 604800 1440
// 1                IN NS    ns1.example.com.
// 2                IN NS    ns2.example.org.
