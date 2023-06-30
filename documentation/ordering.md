# Ordering of DNS records

DNSControl tries to automatically reorder the pending changes based on the dependencies of the records.

Some providers explicitly require the targets of certain records like CNAMEs to exist, and source records to be valid. This makes it not always possible to "just" apply the pending changes in any order. This is why reordering the records based on the dependencies and the type of operation is required.

## Practical example

```js
D('example.com', REG_NONE, DnsProvider(DNS_BIND),
    CNAME('foo', 'bar')
    A('bar', '1.2.3.4'),
);
```

`foo` requires `bar` to exist. Thus `bar` needs to exist before `foo`. But when deleting these records, `foo` needs to be deleted before `bar`.

## Unresolved records

DNSControl can produce a warning stating it found `unresolved records` this is most likely because of a cycle in the targets of your records. For instance in the code sample below both `foo` and `bar` depend on each other and thus will produce the warning.

But have no fear, DNSControl will still try to push both `foo` and `bar` to your provider, it just can't determine the order and thus will append all these unresolved records as last records to update.

```js
D('example.com', REG_NONE, DnsProvider(DNS_BIND),
    CNAME('foo', 'bar')
    CNAME('bar', 'foo'),
);
```


## Disabling ordering

If you encounter any issues with the reordering don't hesitate to [open an issue](https://github.com/StackExchange/dnscontrol/issues). On top of that you can opt-out of the reordering with the `--disableordering` global flag.
