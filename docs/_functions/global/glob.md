---
name: glob
parameters:
  - path
  - recursive
  - fileExtension
---

`glob()` can recursively list files, for all kind of different usage scenarios like dynamically loading JavaScript files.

Possible parameters are:

- Path as string, where you would like to start searching files. Mandatory. Pattern matching possible, see [GoLand path/filepath/#Match docs](https://golang.org/pkg/path/filepath/#Match).
- If being recursive. This is a boolean if the search should be recursive or not. Define either `true` or `false`. Default is `true`.
- This is a file extension to filter for as a string. When not defined, `.js` is the default.

One more important thing to note: `glob()` is as smart as `require()` is. It lists files always relative to the JavaScript
file where it's being executed in. Let's go with an example, as it describes it better:

dnscontrol.js:
```
require("domains/index.js");
```

domains/index.js:
```
console.log( JSON.stringify(glob("./user1/"), null, 2) );
```

This will now show files being present underneath `./domains/user1/` and **NOT** at below `./domains/`, as `glob()`
is called in the subfolder `domains/`.

{% include startExample.html %}
A few examples:
{% highlight js %}
console.log( JSON.stringify(glob("./domains/"), null, 2) ); // default
console.log( JSON.stringify(glob("./domains/", false), null, 2) ); // not recursive
console.log( JSON.stringify(glob("./domains/", true, "*"), null, 2) ); // recursive and showing all files
console.log( JSON.stringify(glob("./domains/", true), null, 2) ); // recursive and only .js files (default)
{%endhighlight%}

Return might look like:
```
[
  "domains/user1/domain1.tld.js",
  "domains/user2/domain2.tld.js",
  "domains/test.js"
]
```

Further example, for loading JavaScript files recursively automatically:
{% highlight js %}
var load = glob("./domains/");
for (i = 0; i < load.length; i++) {
  console.log("Loading " + load[i] + "...")
  require(load[i]);
}
{%endhighlight%}
{% include endExample.html %}

In case you need more details about each specific file, you can use `eglob()` instead, which behaves pretty much the same.
The only difference is that `glob()` returns an array of strings with file paths, while `eglob()` returns an array of objects.

See an example:
{% include startExample.html %}
{% highlight js %}
```
[
  {
    "DirPath": "domains/user1/",
    "FileName": "domain1.tld.js",
    "IsDir": false,
    "ModTime": "2020-08-03T19:37:04+0100",
    "Mode": 666,
    "Size": 406
  },
  {
    "DirPath": "domains/user2/",
    "FileName": "domain2.tld.js",
    "IsDir": false,
    "ModTime": "2020-08-03T19:37:04+0100",
    "Mode": 666,
    "Size": 446
  }
]
```
{%endhighlight%}
{% include endExample.html %}
