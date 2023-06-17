D("foo.com", "none"
  , IGNORE_NAME("testignore")
  , IGNORE_NAME("testignore2", "A")
  , IGNORE_NAME("testignore3", "A, CNAME, TXT")
  , IGNORE_NAME("testignore4", "*")
  , IGNORE_TARGET("testtarget", "CNAME")
  , IGNORE("legacyignore")
  , IGNORE_NAME("@")
  , IGNORE_TARGET("@", "CNAME")
);
D("diff2.com", "none"
  , IGNORE("mylabel")
  , IGNORE("mylabel2", "")
  , IGNORE("mylabel3", "", "")
  , IGNORE("", "A")
  , IGNORE("", "A,AAAA")
  , IGNORE("", "", "mytarget")
  , IGNORE("labelc", "CNAME", "targetc")
  // Compatibility mode:
  , IGNORE_NAME("nametest")
  , IGNORE_TARGET("targettest1")
  , IGNORE_TARGET("targettest2", "A")
);
