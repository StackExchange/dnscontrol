var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// Zone that gets extended by subdomain
D("foo.net", REG, DnsProvider(CF),
  A("@", "10.1.1.1"),
  A("www", "10.2.2.2")
);
D_EXTEND("bar.foo.net",
  A("@", "10.3.3.3"),
  A("www", "10.4.4.4")
);

// Zone and subdomain zone, each get extended.
D("foo.tld", REG, DnsProvider(CF),
  A("@", "20.5.5.5"),
  A("www", "20.6.6.6")
);
D("bar.foo.tld", REG, DnsProvider(CF),
  A("@", "30.7.7.7"),
  A("www", "30.8.8.8")
);
D_EXTEND("bar.foo.tld",
  A("@", "30.9.9.9"),
  A("www", "30.10.10.10")
);
D_EXTEND("foo.tld",
  A("@", "20.11.11.11"),
  A("www", "20.11.11.11")
);

// Zone and subdomain zone, each get extended by a subdomain.
D("foo.help", REG, DnsProvider(CF),
  A("@", "40.12.12.12"),
  A("www", "40.12.12.12")
);
D("bar.foo.help", REG, DnsProvider(CF),
  A("@", "50.13.13.13"),
  A("www", "50.14.14.14")
);
D_EXTEND("zip.bar.foo.help",
  A("@", "50.15.15.15"),
  A("www", "50.16.16.16")
);
D_EXTEND("morty.foo.help",
  A("@", "40.17.17.17"),
  A("www", "40.18.18.18")
);

