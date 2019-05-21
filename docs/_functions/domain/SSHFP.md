---
name: SSHFP
parameters:
  - name
  - algorithm
  - type
  - value
  - modifiers...
---

SSHFP contains a fingerprint of a SSH server which can be validated before SSH clients are establishing the connection.

**Algorithm** (type of the key)
| ID | Algorithm |
|----|-----------|
| 0  | reserved  |
| 1  | RSA       |
| 2  | DSA       |
| 3  | ECDSA     |
| 4  | ED25519   |

**Type** (fingerprint format)
| ID | Algorithm |
|----|-----------|
| 0  | reserved  |
| 1  | SHA-1     |
| 2  | SHA-256   |

`value` is the fingerprint as a string.

{% include startExample.html %}
{% highlight js %}

SSHFP('@', 1, 1, '00yourAmazingFingerprint00'),

{%endhighlight%}
{% include endExample.html %}
