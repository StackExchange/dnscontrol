---
name: require
parameters:
  - path
ts_ignore: true
---

`require(...)` loads the specified JavaScript or JSON file, allowing
to split your configuration across multiple files.

If the supplied `path` string ends with `.js`, the file is interpreted
as JavaScript code, almost as though its contents had been included in
the currently-executing file.  If  the path string ends with `.json`,
`require()` returns the `JSON.parse()` of the file's contents.

If the path string begins with a `.`, it is interpreted relative to
the currently-loading file (which may not be the file where the
`require()` statement is, if called within a function), otherwise it
is interpreted relative to the program's working directory at the time
of the call.

{% code title="dnsconfig.js" %}
```javascript
require('kubernetes/clusters.js');

D("mydomain.net", REG, PROVIDER,
    IncludeKubernetes()
);
```
{% endcode %}

{% code title="kubernetes/clusters.js" %}
```javascript
require('./clusters/prod.js');
require('./clusters/dev.js');

function IncludeKubernetes() {
    return [includeK8Sprod(), includeK8Sdev()];
}
```
{% endcode %}

{% code title="kubernetes/clusters/prod.js" %}
```javascript
function includeK8Sprod() {
    return [
        // ...
    ];
}
```
{% endcode %}

{% code title="kubernetes/clusters/dev.js" %}
```javascript
function includeK8Sdev() {
    return [
        // ...
    ];
}
```
{% endcode %}

You can also use it to require JSON files and initialize variables with it:

{% code title="dnsconfig.js" %}
```javascript
var domains = require('./domain-ip-map.json')

for (var domain in domains) {
    D(domain, REG, PROVIDER,
        A("@", domains[domain])
    );
}
```
{% endcode %}

{% code title="domain-ip-map.json" %}
```javascript
{
    "mydomain.net": "1.1.1.1",
    "myotherdomain.org": "5.5.5.5"
}
```
{% endcode %}

# Future

It might be better to rename the function to something like
`include()` instead, (leaving `require` as a deprecated alias) because
by analogy it is *much* closer to PHP's `include()` function than it
is to node's `require()`.  After all, the reason node.js calls it
"require" is because it's a declarative statement saying the file is
needed, and so should be loaded if it hasn't already been loaded.

In contrast, DNSControl's `require()` is actually an imperative command to
load the file and execute the code or parse the data from it.  (So if
two files both `require("./tools.js")`, for example, then it will be
loaded twice, whereas in node.js it would only be loaded once.)
