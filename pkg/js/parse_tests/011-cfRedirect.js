D("foo.com", "none",
    CF_REDIRECT("test1.foo.com", "https://goo.com/$1"),
    CF_TEMP_REDIRECT("test2.foo.com", "https://goo.com/$1"),
);
