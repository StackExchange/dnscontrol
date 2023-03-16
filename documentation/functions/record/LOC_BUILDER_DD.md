---
name: LOC_BUILDER_DD
parameters:
  - subdomain
  - decimal_degrees_x
  - decimal_degrees_y
  - altitude
  - ttl
parameter_types:
  subdomain: string
  decimal_degrees_x: float32
  decimal_degrees_y: float32
  altitude: float32
  ttl: int
---

`LOC_BUILDER_DD({})` actually takes an object with the mentioned properties.

A helper to build [`LOC`](../domain/LOC.md) records. Supply four parameters instead of 12.

Internally assumes some defaults for [`LOC`](../domain/LOC.md) records.


The cartesian coordinates are decimal degrees, like you typically find in e.g. Google Maps.

Examples.

Big Ben:
`51.50084265331501, -0.12462541415599787`

The White House:
`38.89775977858357, -77.03655125982903`


{% code title="dnsconfig.js" %}
```javascript
D("example.com","none"
  , LOC_BUILDER_DD({
    label: "big-ben",
    x: 51.50084265331501,
    y: -0.12462541415599787,
    alt: 6,
  })
  , LOC_BUILDER_DD({
    label: "white-house",
    x: 38.89775977858357,
    y: -77.03655125982903,
    alt: 19,
  })
  , LOC_BUILDER_DD({
    label: "white-house-ttl",
    x: 38.89775977858357,
    y: -77.03655125982903,
    alt: 19,
    ttl: "5m",
  })
);

```
{% endcode %}


Part of the series:
 * [`LOC()`](LOC.md) - build a `LOC` by supplying all 12 parameters
 * [`LOC_BUILDER_DD({})`](../record/LOC_BUILDER_DD.md) - accepts cartesian x, y
 * [`LOC_BUILDER_DMS_STR({})`](LOC_BUILDER_DMS_STR.md) - accepts DMS 33°51′31″S 151°12′51″E
 * [`LOC_BUILDER_DMM_STR({})`](LOC_BUILDER_DMM_STR.md) - accepts DMM 25.24°S 153.15°E
 * [`LOC_BUILDER_STR({})`](LOC_BUILDER_STR.md) - tries LOC_BUILDER_DM*STR()
