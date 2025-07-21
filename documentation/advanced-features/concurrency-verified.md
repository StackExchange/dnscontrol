---
name: CONCURRENCY_VERIFIED
---

✅  - A checkmark means "this has been tested, and will run concurrently".  
❔  - The questionmark means "it hasn't been tested, safety unknown"  
❌  - The red "X" means "this has been tested, and it will _not_ work concurrently".

Concurrency in this context is about gathering zone data, as seen during the
preview stage when fetching current data for all zones.

### Here's how to perform a test:

1. Build dnscontrol with the -race flag: go build -race (this makes a special binary)
2. Run this special binary: `dnscontrol preview` with a configuration that lists at least 4 domains using the provider in question. More domains is better.
3. The special binary runs much slower (1/10th the speed) because it is doing a lot of checking.

If it reports problems, the error message will indicate where the problem is.
It might not be 100% accurate, but it will be in the right area.

### Development guidance

Loosely, each Provider has a top-level struct named `{providername}Provider`;
if this state is just some simple data-types initialized once before
processing starts, and then only read, then the provider is likely to be
concurrency-safe.

If this contains state which is updated (eg, caches) then the provider is
probably not concurrency-safe unless every routine accessing it is protected
by appropriate primitives.

Eg, URLs and auth credentials are fine.
An account ID determined on the first query?
That probably needs to be protected, even if every fetch returns the same data.

The `-race` build is a helpful hint but is not a guarantee.
