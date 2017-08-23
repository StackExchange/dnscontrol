# How to build a release

Here are my notes from producing the v0.1.5 release.

0.  Run the integration tests (if you are at StackOverflow, thisis "DNS > Integration Tests"


1. Edit the "Version" variable in `main.go`

```
vi main.go
git commit -m'Release v1.5' main.go 

go build
git tag v0.1.5
git push origin tag v0.1.5


https://github.com/StackExchange/dnscontrol/releases/new

Pick the v0.1.5 tag
Release title: Release v0.1.5

Review the git log and make the release notes:

    git log v0.1.5...v0.1.0

Create the binaries:

    go run build/build.go 


Submit the release.

Email the mailing list:

To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v0.1.5

https://github.com/StackExchange/dnscontrol/releases/tag/v0.1.5

So many new providers and features! Plus, a new testing framework that makes it easier to add big features without fear of breaking old ones.

* list
* of 
* major
* changes


```

Add this release to your weekly accomplishments:
 https://github.com/StackExchange/dnscontrol/releases/tag/v0.1.5

