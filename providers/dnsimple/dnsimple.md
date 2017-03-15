## Dnsimple Provider

### required config

In your providers config json file you must provide your DNSimple account access token:

```
"dnsimple":{
  "token": "your-dnsimple-account-access-token"
}
```

You may also specify the baseurl to connect with sandbox:

"dnsimple":{
  "token": "your-sandbox-account-access-token",
  "baseurl": "https://api.sandbox.dnsimple.com"
}
```

### example config

```
var DNSIMPLE_REG = NewRegistrar("dnsimple","DNSIMPLE");
var DNSIMPLE = NewDnsProvider("dnsimple","DNSIMPLE");

D("exammple.tld", DNSIMPLE_REG, DnsProvider(DNSIMPLE),
   A("test","1.2.3.4"),
   CNAME("foo","some.otherdomain.tld.",TTL(100))
);
```
