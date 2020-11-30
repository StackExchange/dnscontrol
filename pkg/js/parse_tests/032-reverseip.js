var REGISTRAR = NewRegistrar('none', 'NONE');    // No registrar.
var BIND = NewDnsProvider('bind', 'BIND');

D(REV('1.2.3.0/24'), REGISTRAR, DnsProvider(BIND),
  PTR("1", 'foo.example.com.'),
  PTR("1.2.3.2", 'bar.example.com.'),
  PTR(REV("1.2.3.3"), 'baz.example.com.', {skip_fqdn_check:"true"})
);
D_EXTEND(REV("1.2.3.4"), PTR("4", "silly.example.com."))
D_EXTEND(REV("1.2.3.5"), PTR("1.2.3.5", "willy.example.com."))
D_EXTEND(REV("1.2.3.6"), PTR(REV("1.2.3.6"), "billy.example.com."))

D_EXTEND(REV("1.2.3.0/24"), PTR("7", "my.example.com."))
D_EXTEND(REV("1.2.3.0/24"), PTR("1.2.3.8", "fair.example.com."))
D_EXTEND(REV("1.2.3.0/24"), PTR(REV("1.2.3.9/32"), "lady.example.com.", {skip_fqdn_check:"true"}))

// Expected zone: 3.2.1.in-addr.arpa.zone
// $TTL 300
//; generated with dnscontrol 2020-11-30T12:56:28-05:00
//@                IN SOA   DEFAULT_NOT_SET. DEFAULT_NOT_SET. 2020113000 3600 600 604800 1440
//1                IN PTR   foo.example.com.
//2                IN PTR   bar.example.com.
//3                IN PTR   baz.example.com.
//4                IN PTR   silly.example.com.
//5                IN PTR   willy.example.com.
//6                IN PTR   billy.example.com.
//7                IN PTR   my.example.com.
//8                IN PTR   fair.example.com.
//9                IN PTR   lady.example.com.
