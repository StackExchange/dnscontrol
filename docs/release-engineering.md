---
layout: default
title: How to build and ship a release
---

# How to build and ship a release

Here are my notes from producing the v2.8 release.  Change the version number as appropriate.


## Step 1. Vendor the modules

Vendor the modules. The vendored files are not used (unless you change
the builds to use `-mod=vendor`). They are maintained simply to make
sure that we have a backup in the unlikely event of a disaster.

```
go mod vendor
git add vendor
git commit -m'vendor modules' vendor
```

TODO(Tom): build.go should verify that this was done, similar to
how it tests that gofmt was run.


## Step 2. Run the integration tests

* If you are at StackOverflow, this is in TC as "DNS > Integration Tests".
* Otherwise:
  * Run "go test" (documented in [Creating new DNS Resource Types](adding-new-rtypes))
  * Run the integration tests (documented in [Writing new DNS providers](writing-providers)


## Step 3. Bump the version number

Edit the "Version" variable in `main.go` and commit.

```
export PREVVERSION=2.10.0       <<< Change to the previous version
export VERSION=2.11.0           <<< Change to the new release version
git checkout master
vi main.go
git commit -m'Release v'"$VERSION" main.go
git tag v"$VERSION"
git push origin tag v"$VERSION"
```


## Step 4. Write the release notes.

The release notes that you write will be used in a few places.

To find items to write about, review the git log using this command:

    git log v"$VERSION"...v"$PREVVERSION"

Entries in the bullet list should be phrased in the positive: "Feature
FOO now does BAR".  This is often the opposite of the related issue,
which was probably phrased, "Feature FOO is broken because of BAR".

Every item should include the ID of the issue related to the change.
If there was no issue, create one and close it.

Sort the list most important/exciting changes earlier in the list.

Put the "[BREAKING CHANGE]" on any breaking change.

Items related to a specific provier should begin with the all-caps
name of the provider, such as "ROUTE53: Added support for sandwiches (#100)"


See [https://github.com/StackExchange/dnscontrol/releases for examples](https://github.com/StackExchange/dnscontrol/releases) for recent release notes and copy that style.

Example/template:

```
This release includes many new providers (JoeDNS and MaryDNS), dozens
of bug fixes, and a new testing framework that makes it easier to add
big features without fear of breaking old ones.

Major features:

* NEW PROVIDER: Providername (#issueid)
* Add FOO DNS record support (#issueid)
* Add SIP/JABBER labels to underscore exception list (#453)

Provider-specific changes:

* PROVIDER: New feature or thing (#issueid)
* PROVIDER: Another feature or bug fixed (#issueid)
* CLOUDFLARE: Fix CF trying to update non-changeable TTL (#issueid)
```

## Step 5. Make the draft release.

[On github.com, click on "Draft a new release"](https://github.com/StackExchange/dnscontrol/releases/new)

Fill in the `Tag version` @ `Target` with:

  * Tag version: v$VERSION (this should be the first tag listed)
  * Target: master (this should be the default)

Release title: Release v$VERSION

Fill in the text box with the release notes written above.

(Don't click SAVE until the next step is complete!)

Create the binaries and attach them to the release:

    go run build/build.go

NOTE: This command creates binaries with the version number and git hash embedded. It also builds the releases for all supported platforms (i.e. creates a .exe for Windows even if you are running on Linux.  Isn't Go amazing?)

WARNING: if there are files that haven't been checked in, the version string will have "dirty" appended.

This is what it looks like when you did it right:

```
$ ./dnscontrol-Darwin version
dnscontrol 2.8 ("ee5208bd5f19b9e5dd0bdba8d0e13403c43a469a") built 19 Dec 18 11:16 EST
```

This is what it looks like when there was a file that wasn't checked in:

```
$ ./dnscontrol-Darwin version
dnscontrol 2.8 ("ee5208bd5f19b9e5dd0bdba8d0e13403c43a469a[dirty]") built 19 Dec 18 11:14 EST
                                                          ^^^^^
                                                          ^^^^^
                                                          ^^^^^
```


## Step 6. Attach the binaries and release.

a. Drag and drop binaries into the web form.

There is a box labeled "Attach binaries by dropping them here or
selecting them".  Drag dnscontrol-Darwin, dnscontrol-Linux, and
dnscontrol.exe onto that box (one at a time or all at once). This
will upload the binaries.

b. Submit the release.

Make sure the "This is a pre-release" checkbox is UNchecked. Then click "Publish Release".


## Step 7. Announce it via email

Email the release notes to the mailing list: (note the format of the Subject line and that the first line of the email is the URL of the release)

```
To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v$VERSION

https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION

[insert the release notes here]
```

NOTE: You won't be able to post to the mailing list unless you are on
it.  [Click here to join](https://groups.google.com/forum/#!forum/dnscontrol-discuss).


## Step 8. Announce it via chat

Mention on [https://gitter.im/dnscontrol/Lobby](https://gitter.im/dnscontrol/Lobby) that the new release has shipped.

```
dnscontrol $VERSION has been released! https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION
```


## Step 9. Get credit!

Mention the fact that you did this release in your weekly accomplishments.

If you are at Stack Overflow:

  * Add the release to your weekly snippets
  * Run this build: `dnscontrol_embed - Promote most recent artifact into ExternalDNS repo`


# Tip: How to update modules

List out-of-date modules and update any that 

```
go get -u github.com/psampaz/go-mod-outdated
go list -u -m -json all | go-mod-outdated -update -direct 
```

To update a module, `get` it, then re-run the unit and integration tests.

```
go get -u
    or
go get -u module/path
```

Once the updates are complete, tidy up:

```
go mod tidy
```

When done, vendor the modules (see Step 1 above).
