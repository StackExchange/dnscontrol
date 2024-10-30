var TRANSFORM_NEWIP = [{
    low: "0.0.0.0",
    high: "1.1.1.1",
    newIP: "2.2.2.2"
}];
var TRANSFORM_BASE = [{
    low: "0.0.0.0",
    high: "1.1.1.1",
    newBase: "4.4.4.4"
}];

D("foo1.com", "reg");

D("foo2.com", "reg",
    IMPORT_TRANSFORM(TRANSFORM_BASE, "int2.com", 60)
);

D("foo3.com", "reg",
    IMPORT_TRANSFORM_STRIP(TRANSFORM_NEWIP, "int3.com", 99, ".com")
);
