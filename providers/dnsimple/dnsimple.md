## Dnsimple Provider

### required config

In your providers config json file you must provide your DNSimple account access token:

```
"yourDnsimpleProviderName":{
  "token": "your-dnsimple-account-access-token"
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
