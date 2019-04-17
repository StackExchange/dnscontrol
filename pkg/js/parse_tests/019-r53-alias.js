D("foo.com", "none",
  R53_ALIAS("mxtest", "MX", "foo.com."),
  R53_ALIAS("atest", "A", "foo.com."),
  R53_ALIAS("atest", "A", "foo.com.", R53_ZONE("Z2FTEDLFRTF")),
  R53_ALIAS("aaaatest", "AAAA", "foo.com."),
  R53_ALIAS("aaaatest", "AAAA", "foo.com.", R53_ZONE("ERERTFGFGF")),
  R53_ALIAS("cnametest", "CNAME", "foo.com."),
  R53_ALIAS("ptrtest", "PTR", "foo.com."),
  R53_ALIAS("txttest", "TXT", "foo.com."),
  R53_ALIAS("srvtest", "SRV", "foo.com."),
  R53_ALIAS("spftest", "SPF", "foo.com."),
  R53_ALIAS("caatest", "CAA", "foo.com."),
  R53_ALIAS("naptrtest", "NAPTR", "foo.com.")
);