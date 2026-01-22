var REG_NONE = NewRegistrar("none");

D("example.com", REG_NONE,
    ADMIN({
        registrar: "cloudflare_main",
        dnsproviders: {
            "gandi_myacct": 2,
            "bind": -1,
        },
        delegations: ["ns1", "ns2.example.com.", "ns4.whatever.com."],
        //delegated_signers: [DS("example.com", 2371, 13, 2, "ABCDEF"), ],
        glue: ["ns1", "ns2", "ns4.whatever.com."],
    }),
);

D("example.net", REG_NONE,
    ADMIN(
        A_REGISTRAR("cloudflare_main"),
        A_DNSPROVIDER("gandi_myacct", 2),
        A_DNSPROVIDER("bind"),
        A_DELEGATIONS("ns1", "ns2.example.com.", "ns4.whatever.com."),
        A_DELEGATED_SIGNERS(
            DS_NEW("example.com", 2371, 13, 2, "ABCDEF"),
            DS_NEW("example.net", 2371, 13, 2, "ABCDEF"),
        ), {
            glue: ["ns1", "ns2", "ns4.whatever.com."]
        },
        // ns1/ns2: are in the zone, must have 1 or more "A" record
        // ns2.whatever.com is not in the zone. that's an error.
    ),
);
