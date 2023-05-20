# Documentation Coding Style

## Where are the docs?

TL;DR version: "documentation" is the docs.dnscontrol.org website.  "docs" is
the marketing website. (Yes, the names are backwards!)

* **The two websites:**

* https://dnscontrol.org/
  * The main website
  * Source code: `docs`
  * Mostly "marketing" for the project.
  * Rarely changes.  Updated via Github "pages" feature.
* https://docs.dnscontrol.org/
  * Project documentation
  * Source code: `documentation`
  * Users and developer documentation
  * Changes frequently.  Updated via [GitBook](https://www.gitbook.com/)

* **The directory structure:**

Within the git repo, docs are grouped:

* `documentation/`: general docs
* `documentation/providers/`: One file per provider
* `documentation/functions/`: One file per dnsconfig.js language feature
* `documentation/assets/FOO/`: Images for page FOO(PNGs only, please!)

## How to add a new page?

* 1. Add the page to the `documentation` (or a sub folder)
* 2. List the page in `SUMMARY.md` so that it will appear in the table of contents, sidebar, etc.

## Documentation previews

NOTE TO SELF: Ask Cafferata how to preview a draft.

## Formatting tips

### General

Break lines every 80 chars.

Include a blank line between paragraphs.

Leave one blank line before and after a heading.

Javascript code should use double quotes (`"`) for strings, not single quotes
(`'`).  They are equivalent but consistency is good.

### Headings

```
#  Title of the page

## Heading

At least one paragraph.

## Subheadings

At least one paragraph.

* **Step 1: Foo**

Description of the step.

* **Step 2: Bar**

Description of the step.

(further sub sub headings are discouraged encouraged)
```

### Code Snippets

Long example: (with filename)

{% code %}
```
    {% code title="dnsconfig.js" %}
    ```javascript
    The code goes here.
    ```
    {% endcode %}
```
{% endcode %}

Long example: (without filename)

{% code title="dnsconfig.js" %}
```
    {% code title="dnsconfig.js" %}
    ```javascript
    The code goes here.
    ```
    {% endcode %}
```
{% endcode %}

### Technical references

* Mentioning language features:

Not every mention to A, CNAME, or function
needs to be a link to the manual for that record type.
However, the first mention on a page should always
be a link.  Others are at the authors digression.

```
The [`PTR`](functions/domain/PTR.md) feature is helpful in LANs.
```

* Mentioning functions from the source code:

```
The function `GetRegistrarCorrections()` returns...
```

### Links

* Internal links:

```
Blah blah blah [M365_BUILDER](functions/record/M365_BUILDER.md)
```

NOTE: The `.md` is required.

* Link to another website:

Just list the URL.

```
Blah blah blah https://www.google.com blah blah.
```

* Link with anchor text:

```
Blah blah blah [a search engine](https://www.google.com) blah blah.
```

## Proofreading

Please spellcheck documents before submitting a PR.

Don't be surprised if Tom rewrites your text.  He often does that to keep the
documentation consistent and make it more approachable by new users.  It's not
[because he has a big ego](https://www.amazon.com/stores/author/B004J0QIVM).
Well, not usually.
