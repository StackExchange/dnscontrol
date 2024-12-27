var CLOUDFLARE = NewRegistrar("Cloudflare", "CLOUDFLAREAPI");

D("foo.com", CLOUDFLARE,
    A("@", "1.2.3.4", {
        "cloudflare_proxy": "ON"
    })
);
