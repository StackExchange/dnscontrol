// This tests PTR records, REV(), and PTR label magic.
// This tests D_EXTEND()'s ability to generate proper labels when REV() is used as a label.
var REGISTRAR = NewRegistrar('none', 'NONE'); // No registrar.
var BIND = NewDnsProvider('bind', 'BIND');

D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
    PTR("1", 'foo.example.com.'),
    PTR("1.2.3.2", 'bar.example.com.'),
    PTR(REV("1.2.3.3"), 'baz.example.com.', {
        skip_fqdn_check: "true"
    }),
);
D_EXTEND(REV("1.2.3.4"),
    PTR("@", "silly.example.com."),
);
D_EXTEND(REV("1.2.3.5/32"),
    PTR("1.2.3.5", "willy.example.com."),
);
D_EXTEND(REV("1.2.3.6"),
    PTR(REV("1.2.3.6"), "billy.example.com."),
);

D_EXTEND(REV("1.2.3.0/24"),
    PTR("7", "my.example.com."),
);
D_EXTEND(REV("1.2.3.0/24"),
    PTR("1.2.3.8", "fair.example.com."),
);
D_EXTEND(REV("1.2.3.0/24"),
    PTR(REV("1.2.3.9/32"), "lady.example.com.", {
        skip_fqdn_check: "true"
    }),
);
