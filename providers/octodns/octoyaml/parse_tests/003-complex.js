var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("bind", "BIND")
D("example.tld",REG,DnsProvider(CF),
    DefaultTTL(303),
    A("one","1.2.3.3"),
    A("foo","1.2.3.4"),
    A("foo","1.2.3.5"),
    MX("foo", 10, "mx1.example.com."),
    MX("foo", 10, "mx2.example.com.")
);
