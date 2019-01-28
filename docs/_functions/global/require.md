---
name: require
parameters:
  - path
---

`require(...)` behaves similarly to its equivalent in node.js. You can use it
to split your configuration across multiple files. If the path starts with a
`.`, it is calculated relative to the current file. For example:

{% include startExample.html %}
{% highlight js %}

// dnsconfig.js
require('kubernetes/clusters.js');

D("mydomain.net", REG, PROVIDER, 
    IncludeKubernetes()
);

{%endhighlight%}

{% highlight js %}

// kubernetes/clusters.js
require('./clusters/prod.js');
require('./clusters/dev.js');

function IncludeKubernetes() {
    return [includeK8Sprod(), includeK8Sdev()];
}

{%endhighlight%}

{% highlight js %}

// kubernetes/clusters/prod.js
function includeK8Sprod() {
    return [ /* ... */ ];
}

{%endhighlight%}

{% highlight js %}

// kubernetes/clusters/dev.js
function includeK8Sdev() {
    return [ /* ... */ ];
}

{%endhighlight%}
{% include endExample.html %}
