D("foo.com", "none",
    MIKROTIK_FWD("@", "8.8.8.8", {
        match_subdomain: "true",
        address_list: "vpn-list"
    }),
    MIKROTIK_NXDOMAIN("blocked"),
    MIKROTIK_FORWARDER("corp.example.com", "10.0.0.53")
);
