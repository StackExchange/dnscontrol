var REG_NONE = NewRegistrar("none");
var DSP_BIND = NewDnsProvider("bind", "BIND");

D(REV("2011:abcd::/32"), REG_NONE, DnsProvider(DSP_BIND),
    PTR("2011:abcd::11", "another11.example.com."),
    PTR(REV("2011:abcd::11"), "another22.example.com."),
);

D(REV("2001:db8::/32"), REG_NONE, DnsProvider(DSP_BIND),
    PTR("2001:db8::11", "reverse11.example.com."),
    PTR(REV("2001:db8::22"), "reverse22.example.com."),
);

D_EXTEND("d.c.b.a.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa",
    PTR("2001:db8::abcd", "extend.example.com.")
);
