D("foo.com", "none"
  , IGNORE("", "", "targetGlob1")
  , IGNORE("", "CNAME", "")
  , IGNORE("", "A", "targetGlob3")
  , IGNORE("lab4")
  , IGNORE("notype", "", "targetGlob5")
  , IGNORE("lab6", "A, CNAME")
  , IGNORE("lab7", "TXT", "targetGlob7")
);
