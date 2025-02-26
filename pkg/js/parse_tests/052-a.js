// This tests "A" records.
// This tests D_EXTEND()'s ability to generate proper labels.
var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

D("foo.com", REG, DnsProvider(CF),
    A("@", "1.1.1.1"),
    A("www", "2.2.2.2"),
);

D_EXTEND("foo.com",
    A("@", "3.3.3.3"),
    A("ewww", "4.4.4.4"),
);

D_EXTEND("sub.foo.com",
    A("@", "5.5.5.5"),
    A("swww", "6.6.6.6"),
);
