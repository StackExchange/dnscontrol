var REGISTRAR = NewRegistrar('none', 'NONE');    // No registrar.
var BIND = NewDnsProvider('bind', 'BIND');

D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
  PTR("1", 'foo.example.com.'),
  PTR("1.2.3.2", 'bar.example.com.'),
  PTR(REV("1.2.3.3"), 'baz.example.com.', {skip_fqdn_check:"true"}),
  PTR(REV("1.2.3.4"), 'blam.example.com.')
);
D_EXTEND(REV("1.2.3.5"), PTR("5", "silly.example.com."))
D_EXTEND(REV("1.2.3.6"), PTR("1.2.3.6", "willy.example.com."))
D_EXTEND(REV("1.2.3.7"), PTR(REV("1.2.3.7"), "billy.example.com."))
