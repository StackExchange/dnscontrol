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

## 3. Examples
See how to use CLI variables 

#### Switch statement
```js
CLI_DEFAULTS({
  'cliServer': 'dns'
});
var ip = "";

switch (this.cliServer) {
  case "webserver":
    ip = "1.2.3.4";
    break;
  case "dns":
    ip = "9.9.9.9";
    break;
}
```
Running `dnscontrol push -v cliServer=webserver` would set the `ip` to `1.2.3.4`.

Just running `dnscontrol push` would set the `ip` to `9.9.9.9` as it is using the default value.