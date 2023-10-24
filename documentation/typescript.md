# Using TypeScript with DNSControl (experimental)

## What is this?

Would you like your editor to support auto-completion and other advanced IDE
features when editing `dnsconfig.js`? Yes you can!

While DNSControl does not support TypeScript syntax in `dnsconfig.js`, you can
still use TypeScript’s features in editors which support it.

If you’re using Visual Studio Code (or another editor that supports TypeScript), you
should now be able to see the type information in your `dnsconfig.js` file as
you type. Hover over record names to read their documentation without having to
open the documentation website!

## How to activate auto-completion

To set up TypeScript support in Visual Studio Code, follow these steps:

1. Run this command to generate the file `types-dnscontrol.d.ts`.

```shell
dnscontrol write-types
```

This file has all the information your editor or IDE needs.  It must be in the same directory as the `dnsconfig.js` file you are editing.

{% hint style="info" %}
**NOTE**: Re-run the `dnscontrol write-types` command any time you upgrade
DNSControl. Because it is generated from the command, it will always be correct
for the version of DNSControl you are using.
{% endhint %}

2. Tell your editor

At this point some features (autocomplete) will work. However to get the full experience, including
type checking (i.e. red squiggly underlines when you misuse APIs), there is one more step.

Add these comments to the top of your `dnsconfig.js` file:

{% code title="dnsconfig.js" %}
```javascript
// @ts-check
/// <reference path="types-dnscontrol.d.ts" />
```
{% endcode %}


That should be all you need to do!

If your editor requires extra steps, please [file a bug](https://github.com/StackExchange/dnscontrol/issues) and we'll update this page.

### Bugs?

{% hint style="warning" %}
**BUGS**: Not all features of DNSControl work perfectly at the moment. Please report bugs and feature requests on https://github.com/StackExchange/dnscontrol/issues
{% endhint %}

{% hint style="info" %}
**NOTE**: This feature is currently experimental. We might change the installation instructions as we find better ways to enable this.
{% endhint %}

## Known bugs/issue

### Bug: `CLI_DEFAULTS` not implemented

Values passed to `CLI_DEFAULTS` (and the corresponding `-v` command-line option) don’t show up as global variables

Workaround: create a new `.d.ts` file in the same folder as your `dnsconfig.js` file. In that file, add the following line for each variable you want to use (replacing `VARIABLE_NAME` with the name of the variable).

{% code title=".d.ts" %}
```javascript
declare const VARIABLE_NAME: string;
```
{% endcode %}


This will tell TypeScript that the variable exists, and that it’s a string.

### Known issue: `FETCH` not always accurate

`FETCH` is always shown as available, even if you don’t run DNSControl with the `--allow-fetch` flag.
