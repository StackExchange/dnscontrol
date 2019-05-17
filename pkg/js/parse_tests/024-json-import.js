var domains = require('./domain-ip-map.json')

var domain = "foo.com"
var ip = domains["foo.com"]

D(domain,"none",
    A("@",ip)
);
