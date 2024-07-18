D("foo.com","none",
    A("name1", "1.2.3.4", { meta: "value" } ),
    CF_SINGLE_REDIRECT("name1", 301, "when1", "then1"),
    CF_SINGLE_REDIRECT("name2", 302, "when2", "then2"),
    CF_SINGLE_REDIRECT("name3", "301", "when3", "then3"),
    CF_SINGLE_REDIRECT("namettl", 302, "whenttl", "thenttl", TTL(999)),
    CF_SINGLE_REDIRECT("namemeta", 302, "whenmeta", "thenmeta", { metastr: "stringy"}, { metanum: 22 } )
);
