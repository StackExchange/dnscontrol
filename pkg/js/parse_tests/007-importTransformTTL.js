D("foo1.com", "reg",
    A("bar", "1.1.1.1"),
    A("foo", "5.5.5.5"),
);

var TRANSFORM_BASE = [{
    low: "1.1.1.0",
    high: "1.1.1.100",
    newBase: "4.4.4.100"
}, {
    low: "5.5.5.2",
    high: "5.5.5.100",
    newBase: "6.6.6.0"
}, ];

D("inny", "reg",
    IMPORT_TRANSFORM(TRANSFORM_BASE, "foo1.com", 60),
);

var TRANSFORM_NEWIP = [{
    low: "5.5.5.0",
    high: "6.0.0.0",
    newIP: "7.7.7.7"
}];

D("com.inny", "reg",
    IMPORT_TRANSFORM_STRIP(TRANSFORM_NEWIP, "foo1.com", 99, "com"),
);
