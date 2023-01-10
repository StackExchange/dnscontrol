var REGISTRAR = NewRegistrar('none', 'NONE');    // No registrar.
var BIND = NewDnsProvider('bind', 'BIND');

// Delegating reverse zones
D(REV('1.3.0.0/16'), REGISTRAR, DnsProvider(BIND),
	NS(REV('1.3.1.0/24'), "ns1.example.com.")
);
D_EXTEND(REV("1.3.2.0/24"), NS(REV("1.3.2.0/24"), "ns2.example.org."))
