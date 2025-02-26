// This tests MX records.
// This tests D_EXTEND()'s ability to generate proper labels and targets.
var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

D("foo.com", REG, DnsProvider(CF),
    MX("@", 1, "aaa"),
    MX("b", 2, "bbb"),
    MX("c", 3, "ccc.com."),
    MX("dwww.foo.com.", 4, "ddd"),
);

D_EXTEND("foo.com",
    MX("@", 5, "eee"),
    MX("fwww", 6, "fff"),
    MX("gwww", 7, "ggg.google.com."),
    MX("hwww.foo.com.", 8, "hhh"),
);

D_EXTEND("sub.foo.com",
    MX("@", 9, "iii"),
    MX("jwww", 10, "jjj"),
    MX("kwww", 11, "kk.bar.com."),
    MX("mwww.sub.foo.com.", 12, "mmm"),
);
