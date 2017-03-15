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
var DNSIMPLE = NewRegistrar("dnsimple","DNSIMPLE");
var DNSIMPLEDSP = NewDSP("dnsimple","DNSIMPLE")

D("exammple.tld", DNSIMPLE, DNSIMPLEDSP,
   //ns[1-4].dnsimple.com used by default as nameservers
      
   A("test","1.2.3.4"),
   
   //override ttl for one record only
   CNAME("foo","some.otherdomain.tld.",TTL(100))
)
```
