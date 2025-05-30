---
name: FRAME
parameters:
  - name
  - target
  - modifiers...
parameter_types:
  name: string
  target: string
  "modifiers...": RecordModifier[]
---

{% hint style="info" %}
This is provider specific type of record and not a DNS standard. It may behave differently for each provider that handles it.
{% endhint %}

### Namecheap

This is a URL Redirect record with a type of "Masked", it creates a framed HTML page to the target.

You can read more at the [Namecheap documentation](https://www.namecheap.com/support/knowledgebase/article.aspx/385/2237/how-to-set-up-a-url-redirect-for-a-domain/).
