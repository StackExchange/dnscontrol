DefaultTTL(301);
var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("bind", "BIND")
D("example.tld",REG,DnsProvider(CF),
    A("foo","1.2.3.4")
);
