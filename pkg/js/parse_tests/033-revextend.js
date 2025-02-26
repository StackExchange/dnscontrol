var REGISTRAR = NewRegistrar("none", "NONE"); // No registrar.
var BIND = NewDnsProvider("bind", "BIND");

// Delegating reverse zones
D(REV("9.8.0.0/16"), REGISTRAR,
    DnsProvider(BIND),
    NS(REV("9.8.2.1"), "ns1.example.com."),
);
D_EXTEND(REV("9.8.7.0/24"),
    NS(REV("9.8.7.6"), "ns2.example.org."),
);

// Forward zone
D("example.com", REGISTRAR,
    DnsProvider(BIND),
    NS("foo", "ns1.fooexample.com."),
);
D_EXTEND("lego.example.com",
    NS("more", "ns1.example.com."),
    NS("short", "ns1"),
);
