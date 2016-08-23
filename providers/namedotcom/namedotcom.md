## name.com Provider

### required config

In your providers config json file you must provide your name.com api username and access token:

```
 "yourNameDotComProviderName":{
      "apikey": "yourApiKeyFromName.com-klasjdkljasdlk235235235235",
      "apiuser": "yourUsername"
 }
```

In order to get api access you need to [apply for access](https://www.name.com/reseller/apply)

### example dns config js (registrar only):

```
var NAMECOM = NewRegistrar("myNameCom","NAMEDOTCOM");

var mynameServers = [
    NAMESERVER("bill.ns.cloudflare.com"),
    NAMESERVER("fred.ns.cloudflare.com")
];

D("example.tld",NAMECOM,myNameServers
    //records handled by another provider...
);
```

### example config (registrar and records managed by namedotcom)

```
var NAMECOM = NewRegistrar("myNameCom","NAMEDOTCOM");
var NAMECOMDSP = NewDSP("myNameCom","NAMEDOTCOM")

D("exammple.tld", NAMECOM, NAMECOMDSP,
   //ns[1-4].name.com used by default as nameservers
   
   //override default ttl of 300s
   DefaultTTL(3600),
   
   A("test","1.2.3.4"),
   
   //override ttl for one record only
   CNAME("foo","some.otherdomain.tld.",TTL(100))
)
```