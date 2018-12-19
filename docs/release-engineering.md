---
layout: default
title: How to build and ship a release
---

# How to build and ship a release

Here are my notes from producing the v0.2.8 release.  Change the version number as appropriate.

## Step 1. Run the integration tests

* If you are at StackOverflow, this is in TC as "DNS > Integration Tests".
* Otherwise:
  * Run "go test" (documented in [Creating new DNS Resource Types](adding-new-rtypes))
  * Run the integration tests (documented in [Writing new DNS providers](writing-providers)


## Step 2. Bump the verison number

Edit the "Version" variable in `main.go` and commit.

```
export PREVVERSION=2.7
export VERSION=2.8
git checkout master
vi main.go
git commit -m'Release v'"$VERSION" main.go
git tag v0."$VERSION"
git push origin tag v0."$VERSION"
```

## Step 3. Make the draft release.

[On github.com, click on "Draft a new release"](https://github.com/StackExchange/dnscontrol/releases/new)

Fill in the `Tag version` @ `Target` with:

  * Tag version: v0.$VERSION (this should be the first tag listed)
  * Target: master (this should be the default)

Release title: Release v0.$VERSION

Fill in the text box with something friendly like, "So many new features!" then make a bullet list of major new functionality.

Review the git log using this command:

    git log v0."$VERSION"...v0."$PREVVERSION"

(Don't click SAVE until the next step is complete!)

Create the binaries and attach them to the release:

    go run build/build.go

NOTE: This command creates binaries with the version number and git hash embedded. It also builds the releases for all supported platforms (i.e. creates a .exe for Windows even if you are running on Linux.  Isn't Go amazing?)

WARNING: if there are unchecked in files, the version string will have "dirty" appended.

This is what it looks like when you did it right:

```
$ ./dnscontrol-Darwin version
dnscontrol 0.2.8 ("ee5208bd5f19b9e5dd0bdba8d0e13403c43a469a") built 19 Dec 18 11:16 EST
```

This is what it looks like when there was a file that should have been checked in:

```
$ ./dnscontrol-Darwin version
dnscontrol 0.2.8 ("ee5208bd5f19b9e5dd0bdba8d0e13403c43a469a[dirty]") built 19 Dec 18 11:14 EST
                                                            ^^^^^
                                                            ^^^^^
                                                            ^^^^^
```

## Step 4. Attach the binaries and release.

a. Drag and drop binaries into the web form.

There is a box labeled "Atttach binaries by dropping them here or
selecting them".  Drag dnscontrol-Darwin, dnscontrol-Linux, and
dnscontrol.exe onto that box (one at a time or all at once). This
will upload the binaries.

b. Submit the release.

Make sure the "This is a pre-release" checkbox is UNchecked. Then click "Publish Release".


## Step 5. Announce it via email

Email the mailing list: (note the format of the Subject line and that the first line of the email is the URL of the release)

```
To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v0.$VERSION

https://github.com/StackExchange/dnscontrol/releases/tag/v0.$VERSION

So many new providers and features! Plus, a new testing framework that makes it easier to add big features without fear of breaking old ones.

* list
* of
* major
* changes
```


## Step 6. Announce it via chat

Mention on [https://gitter.im/dnscontrol/Lobby](https://gitter.im/dnscontrol/Lobby) that the new release has shipped.

```
dnscontrol $VERSION has been released! https://github.com/StackExchange/dnscontrol/releases/tag/v0.$VERSION
```


## Step 7. Get credit!

Mention the fact that you did this release in your weekly accomplishments.

If you are at Stack Overflow:

  * Add the release to your weekly snippets
  * Run this build: `dnscontrol_embed - Promote most recent artifact into ExternalDNS repo`
