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

```javascript
// dnsconfig.js
require('kubernetes/clusters.js');

D("mydomain.net", REG, PROVIDER,
    IncludeKubernetes()
);
```

```javascript
// kubernetes/clusters.js
require('./clusters/prod.js');
require('./clusters/dev.js');

function IncludeKubernetes() {
    return [includeK8Sprod(), includeK8Sdev()];
}
```

```javascript
// kubernetes/clusters/prod.js
function includeK8Sprod() {
    return [
        // ...
    ];
}
```

```javascript
// kubernetes/clusters/dev.js
function includeK8Sdev() {
    return [
        // ...
    ];
}
```

You can also use it to require JSON files and initialize variables with it:

```javascript
// dnsconfig.js
var domains = require('./domain-ip-map.json')

for (var domain in domains) {
    D(domain, REG, PROVIDER,
        A("@", domains[domain])
    );
}
```

```javascript
// domain-ip-map.json
{
    "mydomain.net": "1.1.1.1",
    "myotherdomain.org": "5.5.5.5"
}
```

# Future

It might be better to rename the function to something like
`include()` instead, (leaving `require` as a deprecated alias) because
by analogy it is *much* closer to PHP's `include()` function than it
is to node's `require()`.  After all, the reason node.js calls it
"require" is because it's a declarative statement saying the file is
needed, and so should be loaded if it hasn't already been loaded.

In contrast, dnscontrol's require is actually an imperative command to
load the file and execute the code or parse the data from it.  (So if
two files both `require("./tools.js")`, for example, then it will be
loaded twice, whereas in node.js it would only be loaded once.)
