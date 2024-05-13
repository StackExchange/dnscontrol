{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DNS_BIND = NewDnsProvider("bind");

D("example.com", REG_NONE, DnsProvider(DNS_BIND),
    A("@", "1.2.3.4"),
END);
```
{% endcode %}
