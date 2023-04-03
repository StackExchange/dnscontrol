---
name: require_glob
parameters:
  - path
  - recursive
parameter_types:
  path: string
  recursive: boolean
---

`require_glob()` can recursively load `.js` files, optionally non-recursive as well.

Possible parameters are:

- Path as string, where you would like to start including files. Mandatory. Pattern matching possible, see [GoLand path/filepath/#Match docs](https://golang.org/pkg/path/filepath/#Match).
- If being recursive. This is a boolean if the search should be recursive or not. Define either `true` or `false`. Default is `true`.

Example to load `.js` files recursively:

{% code title="dnsconfig.js" %}
```javascript
require_glob("./domains/");
```
{% endcode %}

Example to load `.js` files only in `domains/`:

{% code title="dnsconfig.js" %}
```javascript
require_glob("./domains/", false);
```
{% endcode %}

One more important thing to note: `require_glob()` is as smart as `require()` is. It loads files always relative to the JavaScript
file where it's being executed in. Let's go with an example, as it describes it better:

{% code title="dnsconfig.js" %}
```javascript
require("domains/index.js");
```
{% endcode %}

{% code title="domains/index.js" %}
```javascript
require_glob("./user1/");
```
{% endcode %}

This will now load files being present underneath `./domains/user1/` and **NOT** at below `./domains/`, as `require_glob()`
is called in the subfolder `domains/`.
