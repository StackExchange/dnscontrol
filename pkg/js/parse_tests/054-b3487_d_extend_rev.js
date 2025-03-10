var DSP_BIND = NewDnsProvider("bind");
var REG_CHANGEME = NewRegistrar("none");
D("6.10.in-addr.arpa", REG_CHANGEME,
    DnsProvider(DSP_BIND),
    PTR("31.104", "example.site.com."),
    PTR("206.104", "example2.site.com."),
);


D_EXTEND(REV("10.6.200.0/24"),
    PTR("50", "ip-10-6-200-50.example.com."),
    PTR("51", "ip-10-6-200-51.example.com."),
    PTR("52", "ip-10-6-200-52.example.com."),
    PTR("53", "ip-10-6-200-53.example.com."),
)

D_EXTEND(REV("10.6.119.0/27"),
    PTR("0", "ip-10-6-119-0.example.com."),
    PTR("1", "ip-10-6-119-1.example.com."),
    PTR("2", "ip-10-6-119-2.example.com."),
    PTR("3", "ip-10-6-119-3.example.com."),
)
