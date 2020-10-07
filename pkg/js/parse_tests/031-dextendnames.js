var REG = NewRegistrar("Third-Party", "NONE");
var DNS = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// Test the name matching algorithm

D("domain.tld", REG, DnsProvider(DNS),
  A("@", "127.0.0.1"),
  A("a", "127.0.0.2"),
  CNAME("b", "c")
);

D("sub.domain.tld", REG, DnsProvider(DNS),
  A("@", "127.0.1.1"),
  A("aa", "127.0.1.2"),
  CNAME("bb", "cc")
);


// Should match domain.tld
D_EXTEND("domain.tld",
  A("@", "127.0.0.3"),
  A("d", "127.0.0.4"), 
  CNAME("e", "f") 
);

// Should match domain.tld
D_EXTEND("ub.domain.tld",
  A("@", "127.0.0.5"), 
  A("g", "127.0.0.6"), 
  CNAME("h", "i") 
);

// Should match sub.domain.tld
D_EXTEND("sub.domain.tld",
  A("@", "127.0.1.3"), 
  A("dd", "127.0.1.4"), 
  CNAME("ee", "ff") 
);

// Should match domain.tld
D_EXTEND("ssub.domain.tld",
  A("@", "127.0.0.7"), 
  A("j", "127.0.0.8"), 
  CNAME("k", "l") 
);
