---
layout: default
title: CLI variables
---
# CLI variables

With dnscontrol you can pass variables from CLI into your `dnsconfig.js`.
This gives you the opportunity to run different code when a value is passed.

## 1. Passing variables
To pass a variable from CLI, just use the parameter `-v key=value` of the commands `preview` or `push`.

Example: `dnscontrol push -v testKey=testValue`

This would pass the variable with the name `testKey` and the value of `testValue` to `dnsconfig.js`

## 2. Define defaults
If you want to define some default values, that are used when no variable is passed from CLI,
you can do this with the following function:

```js
CLI_DEFAULTS({
  'testValue': 'defaultValue'
});
```

You need to define this defaults just once in your `dnsconfig.js`.
Define the defaults **before** using it.

_Please keep in mind, if there is no default value and you do not pass a variable, but you are using it in your `dnsconfig.js` it will fail!_

## 3. Use cases
See some use cases for CLI variables.

#### Different IPs for internal/external DNS
```js
CLI_DEFAULTS({
  'cliServer': 'external'
});
if (this.view == "internal") {
    var host01 = "192.168.0.16";
    var host02 = "192.168.0.17";
} else {
    var host01 = "10.0.0.16";
    var host02 = "10.0.0.17";
}

D("example.org", registrar, DnsProvider(public), DnsProvider(bind),
  A('sitea', host01, TTL(1800)),
  A('siteb', host01, TTL(1800)),
  A('sitec', host02, TTL(1800)),
  A('sited', host02, TTL(1800))
);
```
Running `dnscontrol push -v view=internal` would generate the zone for your internal dns.

Just running `dnscontrol push` would generate the zone for your external dns.

So you can use the same zone for external and internal resolution and there is no need to duplicate it.

#### ProTip
The cli variables functionality permits you to create very complex and
sophisticated configurations, but you shouldn't. Be nice to the next
person that edits the file, who may not be as expert as yourself.
