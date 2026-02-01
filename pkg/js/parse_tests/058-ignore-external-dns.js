// Test IGNORE_EXTERNAL_DNS domain modifier
// Default usage (no prefix)
D("extdns-default.com", "none", IGNORE_EXTERNAL_DNS());

// With custom prefix
D("extdns-custom.com", "none", IGNORE_EXTERNAL_DNS("extdns-"));

// Combined with other records
D("extdns-combined.com", "none",
    IGNORE_EXTERNAL_DNS(),
    A("www", "1.2.3.4"),
    CNAME("api", "www")
);
