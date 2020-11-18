---
layout: default
title: How to build and ship a release
---

# How to build and ship a release

These are the instructions for producing a release.
Please change the version number as appropriate.


## Step 1. Tools check

Make sure you are using the latest version of `go`
(listed on [https://golang.org/dl/](https://golang.org/dl/))

```
go version
```


## Step 2. Check unit and integration tests

There's a GitHub Actions [workflow](https://github.com/StackExchange/dnscontrol/actions?query=workflow%3Abuild) which builds the code and runs a set of unit and integration tests. Make sure all tests are passing, including the integration tests for all DNS providers.


## Step 3. Bump the version number

Edit the "Version" variable in `pkg/version/version.go` and commit.

```
export PREVVERSION=3.3.0     <<< Change to the previous version
export VERSION=3.4.0         <<< Change to the new release version
git checkout master
vi main.go                   <<< Change "Version" to new release version
git commit -m'Release v'"$VERSION" main.go
git tag v"$VERSION"
git push origin tag v"$VERSION"
```

NOTE: If you bump the major version, you need to change all the source
files.  The last time this was done (v2 -> v3) these two commands
automated all that:

```
#  Make all the changes:
sed -i.bak -e 's@github.com.StackExchange.dnscontrol.v2@github.com/StackExchange/dnscontrol/v3@g' go.* $(fgrep -lri --include '*.go' github.com/StackExchange/dnscontrol/v2 *)
# Delete the backup files:
find * -name \*.bak -delete
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
  * Target: master (this should be the default; and disappears when
    you enter the tag)

Release title: Release v$VERSION

Fill in the text box with the release notes written above.

(DON'T click SAVE until the next step is complete!)

(DO use the "preview" tab to proofread the text.)


## Step 6. Publish the release

a. Publish the release.

Make sure the "This is a pre-release" checkbox is UNchecked. Then click "Publish Release".

b. Wait for workflow to complete 

There's a GitHub Actions [workflow](https://github.com/StackExchange/dnscontrol/actions?query=workflow%3Arelease) which automatically builds and attaches
all 3 binaries to the release. Refresh the page after a few minutes and you'll
see dnscontrol-Darwin, dnscontrol-Linux, and dnscontrol.exe attached as assets.


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
ANNOUNCEMENT: dnscontrol $VERSION has been released! https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION
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
go list -mod=mod -u -m -json all | go-mod-outdated -update -direct 
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
