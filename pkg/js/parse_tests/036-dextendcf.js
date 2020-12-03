var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

D("foo.com", REG, DnsProvider(CF));
D_EXTEND("sub.foo.com",
    A("test1.foo.com","10.2.3.1"),
    A("test2.foo.com","10.2.3.2"),
    CF_REDIRECT("test1.foo.com","https://goo.com/$1"),
    CF_TEMP_REDIRECT("test2.foo.com","https://goo.com/$1")
);
