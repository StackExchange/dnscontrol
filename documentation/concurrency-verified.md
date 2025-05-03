---
name: CONCURRENCY_VERIFIED
---

✅  - A checkmark means "this has been tested, and will run concurrently".  
❔  - The questionmark means "it hasn't been tested, safety unknown"  
❌  - The red "X" means "this has been tested, and it will _not_ work concurrently".  


### Here's how to perform a test:

1. Build dnscontrol with the -race flag: go build -race (this makes a special binary)
2. Run this special binary: `dnscontrol preview` with a configuration that lists at least 4 domains using Porkbun. More domains is better.
3. The special binary runs much slower (1/10th the speed) because it is doing a lot of checking. 
If it reports problems, the error message will indicate where the problem is. 
It might not be 100% accurate, but it will be in the right area.
