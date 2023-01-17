---
layout: default
title: Using TypeScript with DNSControl
---

# Using TypeScript with DNSControl (experimental)

## What is this?

Would you like your editor to support auto-completion and other advanced IDE
features when editing `dnsconfig.js`? Yes you can!

While DNSControl does not support TypeScript syntax in `dnsconfig.js`, you can
still use TypeScript’s features in editors which support it.

If you’re using Visual Studio Code (or another editor that supports TypeScript), you
should now be able to see the type information in your `dnsconfig.js` file as
you type. Hover over record names to read their documentation without having to
open the website!

## How to activate

To set up TypeScript support in Visual Studio Code, follow these steps:

1. Run this command to generate the file `types-dnscontrol.d.ts`.

```bash
dnscontrol write-types
```

This file has all the information your editor or IDE needs.  It must be in the same directory as the `dnsconfig.js` file you are editing.

NOTE: Re-run the `dnscontrol write-types` command any time you upgrade
dnscontrol. Because it is generated from the command, it will always be correct
for the version of DNSControl you are using.

2. Tell your editor

At this point some features (autocomplete) will work. However to get the full experience, including
type checking (i.e. red squiggly underlines when you misuse APIs), there is one more step.

Add this comment to the top of your `dnsconfig.js` file:

```
// @ts-check
```

That should be all you need to do!

If your edit requires extra steps, please [file a bug](https://github.com/StackExchange/dnscontrol/issues) and we'll update this page.

### Bugs?

**Bugs?**  Not all features of DNSControl work perfectly at the moment. Please report bugs and feature requests on https://github.com/StackExchange/dnscontrol/issues

**This is experimental.** This feature is currently experimental. We might change the installation instructions as we find better ways to enable this.

## Known bugs

## Bug: `CLI_DEFAULTS` not implemented

Bug: Values passed to `CLI_DEFAULTS` (and the corresponding `-v` command-line option) don’t show up as global variables

Workaround: create a new `.d.ts` file in the same folder as your `dnsconfig.js` file. In that file, add the following line for each variable you want to use (replacing `VARIABLE_NAME` with the name of the variable).

```
declare const VARIABLE_NAME: string;
```

This will tell TypeScript that the variable exists, and that it’s a string.

## Bug: `FETCH` not always accurate

Bug: `FETCH` is always shown as available, even if you don’t run DNSControl with the `--allow-fetch` flag.

Workaround: None
