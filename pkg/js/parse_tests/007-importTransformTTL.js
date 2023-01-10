var TRANSFORM_INT = [
    {low: "0.0.0.0", high: "1.1.1.1", newBase: "2.2.2.2" }
]
D("foo2.com","reg");
D("foo.com","reg",IMPORT_TRANSFORM(TRANSFORM_INT,"foo2.com",60))
