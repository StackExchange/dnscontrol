var TRANSFORM_NEWIP = [{
    low: "0.0.0.0",
    high: "1.1.1.1",
    newIP: "2.2.2.2"
}];
var TRANSFORM_BASE = [{
    low: "0.0.0.0",
    high: "1.1.1.1",
    newBase: "4.4.4.4"
}, {
    low: "7.7.7.7",
    high: "8.8.8.8",
    newBase: "9.9.9.9"
},
];

D("foo1.com", "reg");

D("foo2.com", "reg",
    IMPORT_TRANSFORM(TRANSFORM_BASE, "foo1.com", 60)
);

D("foo3.com", "reg",
    IMPORT_TRANSFORM_STRIP(TRANSFORM_NEWIP, "foo1.com", 99, ".com")
);
