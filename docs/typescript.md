---
layout: default
title: '[EXPERIMENTAL] Using TypeScript with DNSControl'
---

# [EXPERIMENTAL] Using TypeScript with DNSControl

> **NOTE**: This feature is currently experimental. Type information may change in future releases, and any release (even a patch release) could introduce new type-checking errors into your `dnsconfig.js` file. If you plan to use TypeScript, please file bug reports for any issues you encounter, and avoid manually running type-checking as part of your deployment process until the “experimental” label is removed.

While DNSControl does not support TypeScript syntax in `dnsconfig.js`, you can still use TypeScript’s features in editors which support it (including Visual Studio Code). To set up TypeScript support in Visual Studio Code, follow these steps:
First, you’ll need to grab the `dnscontrol.d.ts` file corresponding to your version of DNSControl. You can manually download this file off of GitHub, or you can use the following command:

```bash
dnscontrol write-types
```

When run, `dnscontrol write-types` will create a `dnscontrol.d.ts` file in the current directory, overwriting an existing file if it exists. The file will use the type information corresponding to the current version of `dnscontrol`, so you can be confident that everything in the type declarations are consistent with DNSControl functionality. That does mean that you should re-run `dnscontrol write-types` when you update DNSControl, though!

That should be all you need to do! If you’re using VS Code (or another editor that supports TypeScript), you should now be able to see the type information in your `dnsconfig.js` file as you type. Hover over record names to read their documentation without having to open the website!

## Type Checking

If you add the comment `// @ts-check` to the top of your `dnsconfig.js` file, you can enable _type checking_ for your DNSControl configuration. This will allow your editor’s integrated version of TypeScript to check your configuration for possible mistakes in addition to providing enhanced autocomplete. Note that not all features of DNSControl work perfectly at the moment.

Specifically:

-   Values passed to `CLI_DEFAULTS` (and the corresponding `-v` command-line option) don’t show up as global variables
    -   Workaround: create a new `.d.ts` file in the same folder as your `dnsconfig.js` file. In that file, add the following line: <code>declare const _variableName_: string;</code> for each variable you want to use. This will tell TypeScript that the variable exists, and that it’s a string.
-   `FETCH` is always shown as available, even if you don’t run DNSControl with the `--allow-fetch` flag.
