D("example.com", "none",
  A("normal", "1.2.3.4"),
  A("helper", "1.2.3.4", ENSURE_ABSENT_HELPER()),
  //ENSURE_ABSENT(A("wrapped", "1.2.3.4")),
  {});
