var REG = NewRegistrar("Third-Party", "NONE");
var DNS = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// The example from docs/_functions/global/D_EXTEND.md

D("domain.tld", REG, DnsProvider(DNS),
  A("@", "127.0.0.1"), // domain.tld
  A("www", "127.0.0.2"), // www.domain.tld
  CNAME("a", "b") // a.domain.tld -> b.domain.tld
);
D_EXTEND("domain.tld",
  A("aaa", "127.0.0.3"), // aaa.domain.tld
  CNAME("c", "d") // c.domain.tld -> d.domain.tld
);
D_EXTEND("sub.domain.tld",
  A("bbb", "127.0.0.4"), // bbb.sub.domain.tld
  A("ccc", "127.0.0.5"), // ccc.sub.domain.tld
  CNAME("e", "f") // e.sub.domain.tld -> f.sub.domain.tld
);
D_EXTEND("sub.sub.domain.tld",
  A("ddd", "127.0.0.6"), // ddd.sub.sub.domain.tld
  CNAME("g", "h") // g.sub.domain.tld -> h.sub.domain.tld
);
D_EXTEND("sub.domain.tld",
  A("@", "127.0.0.7"), // sub.domain.tld
  CNAME("i", "j") // i.sub.domain.tld -> j.sub.domain.tld
);
