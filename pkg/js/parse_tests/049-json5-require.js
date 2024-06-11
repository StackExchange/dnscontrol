var domains = require('./domain-ip-map.json5')

var domain = "foo.com"
var ip = domains["foo.com"]

D(domain,"none",
    A("@",ip)
);
