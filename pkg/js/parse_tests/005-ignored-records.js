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
