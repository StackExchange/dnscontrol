# Documentation Style

The DNSControl documentation uses Markdown and is rendered by [GitBook](https://docs.gitbook.com). This allows us to use some additional render rendering.

{% hint style="info" %}
You can learn more about Markdown itself by visiting [Common Mark](https://commonmark.org/help/).
{% endhint %}

## Text formatting

We support all the classic inline Markdown formatting:

| Formatting    | Markdown version  | Result          |
|---------------|-------------------|-----------------|
| Bold          | `**bold**`        | **bold**        |
| Italic        | `_italic_`        | _italic_        |
| Strikethrough | `~strikethrough~` | ~strikethrough~ |

## Titles

* Heading 1: `# A first-level title`
* Heading 2: `## A second-level title`
* Heading 3: `### A third-level title`

## Links

You can insert three different types of links:

1. [Relative links](#relative-links)
2. [Absolute links](#absolute-links)
3. [Email address mailto links](#email-address-mailto-links)

### Relative links

Relative links are links created by linking pages that already exist in the GitHub repository.

{% code title="Markdown syntax" %}
```markdown
[Documentation Style](documentation-style.md)
```
{% endcode %}

### Absolute links

Absolute links are for external links.

{% hint style="info" %}
External links will always open in a new tab.
{% endhint %}

{% code title="Markdown syntax" %}
```markdown
[GitHub releases](https://github.com/StackExchange/dnscontrol/releases/latest)
```
{% endcode %}

### Email address mailto links

Email address `mailto` links are useful when you want click on a link that will open up the default email client, fill in TO with the email address of your link, and allow them to write an email to send out.

{% code title="Markdown syntax" %}
```markdown
[example@example.com](mailto:example@example.com)
```
{% endcode %}

## Lists

* Begin a line with `-` or `*` to start a [bullet list](#unordered-list).
* Being a line with `1.` to start a [ordered list](#ordered-list).
* Begin a line with `- [ ]` to start a [task list](#task-list).

### Unordered list

Unordered lists are great for making a series of points that do not necessarily need to be made in a particular order.

- Item
    - Nested item
        - Another nested item
    - Yet another nested item
- Another item
- Yet another item

{% code title="Markdown syntax" %}
```markdown
- Item
   - Nested item
      - Another nested item
   - Yet another nested item
- Another item
- Yet another item
```
{% endcode %}

### Ordered list

Ordered or numbered lists help you prioritize items or create a list of steps.

1. Item 1
    1. Nested item 1.1
        1. Nested item 1.1.1
    2. Nested item 1.2
2. Item 2
3. Item 3

{% code title="Markdown syntax" %}
```markdown
1. Item 1
   1. Nested item 1.1
      1. Nested item 1.1.1
   2. Nested item 1.2
2. Item 2
3. Item 3
```
{% endcode %}

### Task list

Task lists allow you to create a list of items with checkboxes that you can check or uncheck. This is useful for tracking project items, shopping lists, create playbooks and more.

- [ ] Here's a task that hasn't been done
    - [x] And here's a subtask that has been done, indented using `tab`
    - [ ] Aaaaand, here's a subtask that hasn't been done.
- [ ] Finally, an item, unidented using `shift` + `tab`.

{% code title="Markdown syntax" %}
```markdown
- [ ] Here's a task that hasn't been done
  - [x] And here's a subtask that has been done, indented using `tab`
  - [ ] Aaaaand, here's a subtask that hasn't been done.
- [ ] Finally, an item, unidented using `shift` + `tab`.
```
{% endcode %}

## Quotes

Begin a line with `>` to create a block quote.

### Example of a quote

> "No human ever steps in the same river twice, for it's not the same river and they are not the same human." — _Heraclitus_

{% code title="Markdown syntax" %}
```markdown
> "No human ever steps in the same river twice, for it's not the same river and they are not the same human." — _Heraclitus_
```
{% endcode %}


## Hint

Hints are a great way to bring the reader's attention to specific elements in your documentation.

There are 4 different types of hints, and both inline content and formatting are supported.

### Example of a hint

{% hint style="info" %}
**Info hints** are great for showing general information, or providing tips and tricks.
{% endhint %}

{% hint style="success" %}
**Success hints** are good for showing positive actions or achievements.
{% endhint %}

{% hint style="warning" %}
**Warning hints** are good for showing important information or non-critical warnings.
{% endhint %}

{% hint style="danger" %}
**Danger hints** are good for highlighting destructive actions or raising attention to critical information.
{% endhint %}

{% hint style="info" %}
### This is a heading

This is a line

This is a second <mark style="color:white;background-color:green;">line</mark>
{% endhint %}

{% code title="Markdown syntax" %}
```markdown
{% hint style="info" %}
**Info hints** are great for showing general information, or providing tips and tricks.
{% endhint %}

{% hint style="success" %}
**Success hints** are good for showing positive actions or achievements.
{% endhint %}

{% hint style="warning" %}
**Warning hints** are good for showing important information or non-critical warnings.
{% endhint %}

{% hint style="danger" %}
**Danger hints** are good for highlighting destructive actions or raising attention to critical information.
{% endhint %}

{% hint style="info" %}
### This is a heading

This is a line

This is a second <mark style="color:white;background-color:green;">line</mark>
{% endhint %}
```
{% endcode %}

## Code block

You can show code using code blocks by placing triple backticks ` ``` ` before and after the code block. You can choose to set the [syntax](#syntax) and show a [caption](#caption). We recommend placing a blank line before and after code blocks to make the raw formatting easier to read.

### Options

#### Syntax

You can set the syntax to any of the supported languages and that will enable syntax highlighting in that language.

{% hint style="info" %}
GitBook use [Prism](https://github.com/PrismJS/prism) for syntax highlighting. Here's an easy way to check which languages Prism supports: [Test Drive Prism](https://prismjs.com/test.html#language=markup).
{% endhint %}

#### Caption

A code block can have a caption. The caption is often the name of a file as shown in our example, but it can be used as a title, or anything else you'd like.

### Example of a code block

{% code title="dnsconfig.js" %}
```javascript
D('example.com', REG, DnsProvider('R53'),
    A('@', '1.2.3.4')
);
```
{% endcode %}

{% code title="Markdown syntax" %}
````markdown
{% code title="dnsconfig.js" %}
```javascript
D('example.com', REG, DnsProvider('R53'),
    A('@', '1.2.3.4')
);
```
{% endcode %}
````
{% endcode %}
