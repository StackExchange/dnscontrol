D("foo.com","none",
    // Allow letsencrypt to issue certificate for this domain
    CAA("@","issue","letsencrypt.org"),
    // Allow no CA to issue wildcard certificate for this domain
    CAA("@","issuewild",";"),
    // Report all violation to test@example.com. If CA does not support
    // this record then refuse to issue any certificate
    CAA("@", "iodef", "mailto:test@example.com", CAA_CRITICAL),
    // Optionally report violation to http://example.com
    CAA("@", "iodef", "http://example.com"),
    // Report violation to https://example.com
    CAA("@", "iodef", "https://example.com", CAA_CRITICAL)
);
