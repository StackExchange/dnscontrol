---
name: PURGE
---

PURGE is the default setting for all domains.  Therefore PURGE is
a no-op. It is included for completeness only.

A domain with a mixture of NO_PURGE and PURGE parameters will abide
by the last one.

These three examples all are equivalent.

PURGE is the default:

```javascript
D("example.com", .... ,
);
```

Purge is the default, but we set it anyway:

```javascript
D("example.com", .... ,
  PURGE,
);
```

Since the "last command wins", this is the same as `PURGE`:

```javascript
D("example.com", .... ,
  PURGE,
  NO_PURGE,
  PURGE,
  NO_PURGE,
  PURGE,
);
```
