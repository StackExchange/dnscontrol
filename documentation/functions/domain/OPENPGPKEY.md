---
name: OPENPGPKEY
parameters:
  - local
  - digest
  - modifiers...
parameters_object: true
parameter_types:
  local: string
  digest: string
  "modifiers...": RecordModifier[]
---


This is a rough implementation of [RFC 7929](https://www.rfc-editor.org/rfc/rfc7929). Rough?

Rough: prior to the SHA256 hash step, a number of UTF8 normalization steps
are done, but not all.

`OPENPGPKEY({})` allows you to store a record of type `OPENPGPKEY`.

It currently takes two parameters in an object:

 * local - this is the local part of the email address (before the `@`) of the key
	- everything after the `@` is discarded
	- whitespace is removed
	- various forms of quotation marks are removed
	- several UTF normalization steps process the text to UTF8
	- the 28 octet truncated SHA256 hash of the UTF8 is produced
	- the hash is suffixed with `._openpgpkey` as specified in the RFC
 * digest - this is the base64 part of the key
	- `-----* PGP PUBLIC KEY BLOCK-----` lines are discarded
	- the CRC portion of any ASCII armored text block (radix64/base64) is discarded
	- whitespace and linebreaks are removed

 
This is an (ed25519) Open PGP key:
```text
-----BEGIN PGP PUBLIC KEY BLOCK-----

mDMEZCMu8xYJKwYBBAHaRw8BAQdAH4FTbN/H5SoMBl9Ez2cFQ1NuzymK894fq2ff
sYDvRkG0EWFsaWNlQGV4YW1wbGUuY29tiJYEExYKAD4CGwMFCwkIBwMFFQoJCAsF
FgIDAQACHgECF4AWIQRjw8oAQytQxDz5Q/Io7xpohfeBngUCZCMv5gUJAAk7ZgAK
CRAo7xpohfeBnlmVAP9k0slIpLwddCD1bZ9qVjqzNcS743OIDny7XuH6x02L2wEA
wxqAotO7/oUm0L4wyYR6hvGlhuGMSZXc9xMwZ1wVcA8=
=vHSO
-----END PGP PUBLIC KEY BLOCK-----
```

The `digest` portion is the base64 portion without the trailing CRC portion
(the last base64 line starting `=` - in this case - `=vHSO`) at the end.

In effect:

```text
mDMEZCMu8xYJKwYBBAHaRw8BAQdAH4FTbN/H5SoMBl9Ez2cFQ1NuzymK894fq2ff
sYDvRkG0EWFsaWNlQGV4YW1wbGUuY29tiJYEExYKAD4CGwMFCwkIBwMFFQoJCAsF
FgIDAQACHgECF4AWIQRjw8oAQytQxDz5Q/Io7xpohfeBngUCZCMv5gUJAAk7ZgAK
CRAo7xpohfeBnlmVAP9k0slIpLwddCD1bZ9qVjqzNcS743OIDny7XuH6x02L2wEA
wxqAotO7/oUm0L4wyYR6hvGlhuGMSZXc9xMwZ1wVcA8=
```

Example:


{% code title="dnsconfig.js" %}
```javascript
D("example.com","none"
  // hugh@example.com -> c93f1e400f26708f98cb19d936620da35eec8f72e57f9eec01c1afd6._openpgpkey.example.com -> testing123 -> dGVzdGluZzEyMw==
  , OPENPGPKEY({local:"hugh@", digest:"dGVzdGluZzEyMw=="})
  // 麻衣子@example.com -> 2bb5bc4202aaecd48dcb54967c8e7f1b7574a436f04e0d15534b20e5._openpgpkey.example.com
  , OPENPGPKEY({local:"麻衣子@", digest:"\
  mDMEZCMxgRYJKwYBBAHaRw8BAQdA/fgtlQjGflt2MUMWhRZRnH5Hg+BY9sQTeePm\
  qqUs+lK0Fem6u+iho+WtkEBleGFtcGxlLmNvbYiWBBMWCgA+AhsDBQsJCAcDBRUK\
  CQgLBRYCAwEAAh4BAheAFiEEIWsEkWx5wygGCb61+tJ3q3m88E0FAmQjMbMFCQAJ\
  OqwACgkQ+tJ3q3m88E0z4gEAtowKJMPefyV5YCW8VubgXK7Fa+hjwXOPSsHnEnJw\
  9pUBAL+VZvNZv/VZvyGGMd31Yivqerzl6q+VIkZ6XffVb2AB\
  =sRIg"})
);
```
{% endcode %}
