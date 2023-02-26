D("example.com", "none",
  A("normal", "1.1.1.1"),
  A("helper", "2.2.2.2", ENSURE_ABSENT_REC()),
  //ENSURE_ABSENT(A("wrapped", "3.3.3.3")),
  {});
