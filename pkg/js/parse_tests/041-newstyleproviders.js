// Test old-style and new-style New*() functions.

var REG1 = NewRegistrar("foo1");
var CF1 = NewDnsProvider("dns1");

var REG2a = NewRegistrar("foo2a", "NONE");
var CF2a = NewDnsProvider("dns2a", "CLOUDFLAREAPI");

var REG2b = NewRegistrar("foo2b", {
    regmetakey: "reg2b"
});
var CF2b = NewDnsProvider("dns2b", {
    dnsmetakey: "dns2b"
});

var REG3 = NewRegistrar("foo3", "MANUAL", {
    regmetakey: "reg3"
});
var CF3 = NewDnsProvider("dns3", "CLOUDFLAREAPI", {
    dnsmetakey: "dns3"
});

var REG1h = NewRegistrar("foo1h", "-");
var CF1h = NewDnsProvider("dns1h", "-");

var REG2bh = NewRegistrar("foo2bh", "-", {
    regmetakey: "reg2bh"
});
var CF2bh = NewDnsProvider("dns2bh", "-", {
    dnsmetakey: "dns2bh"
});
