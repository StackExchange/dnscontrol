
var REG = NewRegistrar("Third-Party","NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI")
D("foo.com",REG,DnsProvider(CF,2),
    A("@","1.2.3.4")
);
D("foo.com",REG);