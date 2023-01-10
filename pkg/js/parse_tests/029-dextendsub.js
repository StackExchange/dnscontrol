var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// Zone that gets extended by subdomain
D("foo.net", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "10.1.1.1"),
  A("www", "10.2.2.2")
);
D_EXTEND("bar.foo.net",
  A("@", "10.3.3.3"),
  A("www", "10.4.4.4")
);

// Zone and subdomain zone, each get extended.
D("foo.tld", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "20.5.5.5"),
  A("www", "20.6.6.6")
);
D("bar.foo.tld", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "30.7.7.7"),
  A("www", "30.8.8.8")
);
D_EXTEND("bar.foo.tld",
  A("a", "30.9.9.9")
);
D_EXTEND("foo.tld",
  A("a", "20.10.10.10")
);

// Zone and subdomain zone, each get extended by a subdomain.
D("foo.help", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "40.12.12.12"),
  A("www", "40.12.12.12")
);
D("bar.foo.help", REG, DnsProvider(CF),
  DefaultTTL(300),
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

// Zone extended by a subdomain and sub-subdomain.
D("foo.here", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "60.19.19.19"),
  A("www", "60.20.20.20")
);
D_EXTEND("bar.foo.here",
  A("@", "60.21.21.21"),
  A("www", "60.22.22.22")
);
D_EXTEND("baz.bar.foo.here",
  A("@", "60.23.23.23"),
  A("www", "60.24.24.24")
);

// Zone extended by a sub-subdomain.
D_EXTEND("a.long.path.of.sub.domains.foo.net",
  A("@", "10.25.25.25"),
  A("www", "10.26.26.26")
);

// ASCII zone
D("example.com", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "10.0.0.1"),
  A("www", "10.0.0.2")
);
// … extended by an IDN subdomain
D_EXTEND("düsseldorf.example.com",
  A("@", "10.0.0.3"),
  A("www", "10.0.0.4")
);
// … extended by a one-character IDN subdomain
D_EXTEND("ü.example.com",
  A("@", "10.0.0.5"),
  A("www", "10.0.0.6")
);

// IDN zone
D("düsseldorf.example.net", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "10.0.0.7"),
  A("www", "10.0.0.8")
);
// … extended by an ASCII subdomain
D_EXTEND("subdomain.düsseldorf.example.net",
  A("@", "10.0.0.9"),
  A("www", "10.0.0.10")
);
// … extended by an IDN subdomain
D_EXTEND("düsseltal.düsseldorf.example.net",
  A("@", "10.0.0.11"),
  A("www", "10.0.0.12")
);
// … extended by a one-character IDN subdomain
D_EXTEND("ü.düsseldorf.example.net",
  A("@", "10.0.0.13"),
  A("www", "10.0.0.14")
);

// One-character IDN zone
D("ü.example.net", REG, DnsProvider(CF),
  DefaultTTL(300),
  A("@", "10.0.0.15"),
  A("www", "10.0.0.16")
);
// … extended by an ASCII subdomain
D_EXTEND("subdomain.ü.example.net",
  A("@", "10.0.0.17"),
  A("www", "10.0.0.18")
);
// … extended by an IDN subdomain
D_EXTEND("düsseldorf.ü.example.net",
  A("@", "10.0.0.19"),
  A("www", "10.0.0.20")
);
// … extended by a one-character IDN subdomain
D_EXTEND("ü.ü.example.net",
  A("@", "10.0.0.21"),
  A("www", "10.0.0.22")
);

// Zone extended by a subdomain, with absolute and relative CNAME targets
D("example.tld", REG, DnsProvider(CF), DefaultTTL(300));
D_EXTEND("sub.example.tld",
    CNAME("a", "b"), // a.sub.example.tld -> b.sub.example.tld
    CNAME("b", "@"), // a.sub.example.tld -> sub.example.tld
    CNAME("c", "sub.example.tld."), // a.sub.example.tld -> sub.example.tld
    //CNAME("d", "x.y"), // Error. Contains dot but doesn't end with dot.
    CNAME("e", "otherdomain.tld.") // a.sub.example.tld -> otherdomain.tld
    // Also test for MX, NS, ANAME, SRV.
    // Not sure if PTR needs any special treatment. Haven't thought about it much.
);
