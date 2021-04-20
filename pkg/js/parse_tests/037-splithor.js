var REG = NewRegistrar("Third-Party", "NONE");
var DNS_MAIN = NewDnsProvider("otherconfig", "CLOUDFLAREAPI");
var DNS_INSIDE = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");
var DNS_OUTSIDE = NewDnsProvider("bind", "BIND");

D("example.com", REG, DnsProvider(DNS_MAIN),
  A("main", "3.3.3.3")
);

D("example.com!inside", REG, DnsProvider(DNS_INSIDE),
  A("main", "1.1.1.1")
);

D("example.com!outside", REG, DnsProvider(DNS_OUTSIDE),
  A("main", "8.8.8.8")
);

D_EXTEND("example.com",
  A("www", "33.33.33.33")
);

D_EXTEND("example.com!inside",
  A("main", "11.11.11.11")
);
