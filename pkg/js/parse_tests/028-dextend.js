var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// Zone and subdomain Zone:
D("foo.com", REG, DnsProvider(CF),
  A("@", "10.1.1.1"),
  A("www", "10.2.2.2")
);
D("bar.foo.com", REG, DnsProvider(CF),
  A("@", "10.3.3.3"),
  A("www", "10.4.4.4")
);

// Zone that gets extended
D("foo.edu", REG, DnsProvider(CF),
  A("@", "10.5.5.5"),
  A("www", "10.6.6.6")
);
D_EXTEND("foo.edu",
  A("more1", "10.7.7.7"),
  A("more2", "10.8.8.8")
);
