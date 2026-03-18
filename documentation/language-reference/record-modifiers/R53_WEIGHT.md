---
name: R53_WEIGHT
parameters:
  - weight
  - set_identifier
parameter_types:
  weight: number
  set_identifier: string
ts_return: RecordModifier
provider: ROUTE53
---

`R53_WEIGHT` configures [Route 53 weighted routing](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy-weighted.html) for a record.

Weighted routing lets you associate multiple resources with a single domain name and control the proportion of traffic that is routed to each resource.

- `weight`: An integer between 0 and 255. Route 53 distributes traffic proportionally based on the weights assigned to each record with the same name and type.
- `set_identifier`: A unique string that differentiates this record from other records with the same name and type. Each weighted record in a group must have a unique set identifier.

{% code title="dnsconfig.js" %}
```javascript
var REG_NONE = NewRegistrar("none");
var DSP_R53 = NewDnsProvider("r53_main");

D("example.com", REG_NONE, DnsProvider(DSP_R53),
    // 70% of traffic goes to 1.2.3.4, 30% to 5.6.7.8
    A("www", "1.2.3.4", R53_WEIGHT(70, "web-east")),
    A("www", "5.6.7.8", R53_WEIGHT(30, "web-west")),
);
```
{% endcode %}

`R53_WEIGHT` can be used with any record type supported by Route 53 weighted routing, including `A`, `AAAA`, `CNAME`, `TXT`, and [`R53_ALIAS()`](../domain-modifiers/R53_ALIAS.md).

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_NONE, DnsProvider(DSP_R53),
    // Weighted CNAME records
    CNAME("cdn", "east.cdn.example.com.", R53_WEIGHT(70, "cdn-east")),
    CNAME("cdn", "west.cdn.example.com.", R53_WEIGHT(30, "cdn-west")),

    // Weighted R53_ALIAS records
    R53_ALIAS("api", "A", "alb-east.us-east-1.elb.amazonaws.com.",
        R53_WEIGHT(60, "api-east"),
        R53_ZONE("Z35SXDOTRQ7X7K"),
    ),
    R53_ALIAS("api", "A", "alb-west.us-west-2.elb.amazonaws.com.",
        R53_WEIGHT(40, "api-west"),
        R53_ZONE("Z1H1FL5HABSF5"),
    ),
);
```
{% endcode %}

You can optionally add a health check using [`R53_HEALTH_CHECK_ID()`](R53_HEALTH_CHECK_ID.md) to remove unhealthy endpoints from the rotation.

{% code title="dnsconfig.js" %}
```javascript
D("example.com", REG_NONE, DnsProvider(DSP_R53),
    A("www", "1.2.3.4",
        R53_WEIGHT(70, "web-east"),
        R53_HEALTH_CHECK_ID("12345678-1234-1234-1234-123456789012"),
    ),
    A("www", "5.6.7.8",
        R53_WEIGHT(30, "web-west"),
        R53_HEALTH_CHECK_ID("87654321-4321-4321-4321-210987654321"),
    ),
);
```
{% endcode %}
