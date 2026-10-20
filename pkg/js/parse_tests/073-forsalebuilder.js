// Test FORSALE_BUILDER (RFC 10023): no options, every content tag, a list,
// a label, TTL() at the call, and D_EXTEND().

D("domain.tld", "none",
    FORSALE_BUILDER("@"),
    FORSALE_BUILDER("exco", {
        fcod: ["EXCO-ZGVhZGJlZWYx", "XYZ1-MTExLTIyMi0zMzMtNDQ0"],
        ftxt: "This domain name is for sale",
        furi: "https://fs.example.com/",
        fval: "EUR500",
    }, TTL("1h")),
);

D_EXTEND("sub.domain.tld",
    FORSALE_BUILDER("@", {
        furi: "mailto:hq@example.com?subject=foo",
    }),
);
