var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("bind", "BIND")
D("example.tld",REG,DnsProvider(CF),
    DefaultTTL(301),
    A("foo","1.2.3.4", TTL(301))
);
