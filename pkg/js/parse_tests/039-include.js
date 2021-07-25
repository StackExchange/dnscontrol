
var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

D("foo.com!external",REG,DnsProvider(CF),
    A("@","1.2.3.4")
);

D("foo.com!internal",REG,DnsProvider(CF),
  INCLUDE("foo.com!external"),
  A("local","127.0.0.1")
);
