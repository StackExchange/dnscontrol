var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("bind", "BIND")
D("example.tld",REG,DnsProvider(CF),
    DefaultTTL(302),
    A("foo","1.2.3.4"),
    A("foo","1.2.3.5")
);
