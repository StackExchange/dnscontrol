var REG = NewRegistrar("Third-Party", "NONE");
var CF = NewDnsProvider("Cloudflare", "CLOUDFLAREAPI");

// This is a cloudflare-specific transform. It is no longer used because
// IMPORT_TRANSFORM works for all providers.

var TRANSFORM_INT = [{
    low: "0.0.0.0",
    high: "1.1.1.0",
    newBase: "2.2.2.2"
}, {
    low: "1.1.1.1",
    high: IP("2.2.2.2"),
    newIP: ["3.3.3.3", "4.4.4.4", IP("5.5.5.5")]
}];

D("foo.com", REG, DnsProvider(CF),
    A("@", "1.2.3.4", {
        transform: TRANSFORM_INT
    }),
);
