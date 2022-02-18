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

```bash
go version
```


## Step 2. Create a new release branch

From the "master" branch, run `bin/make-release.sh v3.xx.y` where
"v3.xx.y" should be the new release version.

NOTE: This warning can be ignored: `error: failed to push some refs to 'github.com:StackExchange/dnscontrol.git'`

The `make-release.sh` script will do the following:

1. Tag the current branch locally and remotely.
2. Update main.go with the new version string.
3. Create a file called draft-notes.txt which you will edit into the
   release notes.
4. Print instructions on how to create the release PR.


NOTE: If you bump the major version, you need to change all the source
files.  The last time this was done (v2 -> v3) these two commands
automated all that:

```bash
#  Make all the changes:
sed -i.bak -e 's@github.com.StackExchange.dnscontrol.v2@github.com/StackExchange/dnscontrol/v3@g' go.* $(fgrep -lri --include '*.go' github.com/StackExchange/dnscontrol/v2 *)
# Delete the backup files:
find * -name \*.bak -delete
```


## Step 3. Verify the version string

Verify the version string was updated:

```bash
grep Version main.go
```

(Make sure that it lists the new version number.)


## Step 4. Make the PR

The output of "make-release.sh" will mention a URL that will create the PR ("1. Create a PR:")

Load the URL and make the PR.


## Step 5. Write the release notes.

draft-notes.txt is just a draft and needs considerable editing.

Once complete, the contents of this file will be used in multiple
places (release notes, email announcements, etc.)

Entries in the bullet list should be phrased in the positive: "Feature
FOO now does BAR".  This is often the opposite of the related issue,
which was probably phrased, "Feature FOO is broken because of BAR".

Every item should include the ID of the issue related to the change.
If there was no issue, create one and close it.

Sort the list most important/exciting changes earlier in the list.

Items related to a specific provier should begin with the all-caps
name of the provider, such as "ROUTE53: Added support for sandwiches (#100)"

See [https://github.com/StackExchange/dnscontrol/releases for examples](https://github.com/StackExchange/dnscontrol/releases) for recent release notes and copy that style.

Example/template:

```text
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


## Step 6. Make the draft release.

[On github.com, click on "Draft a new release"](https://github.com/StackExchange/dnscontrol/releases/new)

Fill in the `Tag version` @ `Target` with:

  * Tag version: v$VERSION (this should be the first tag listed)
  * Target: master (this should be the default; and disappears when
    you enter the tag)

Release title: Release v$VERSION

Fill in the text box with the release notes written above.

(DON'T click SAVE until the next step is complete!)

(DO use the "preview" tab to proofread the text.)


## Step 7. Verify tests completed.

By now the tests should have started running.

Verify that the automated tests passed. If not, fix the problems
before you continue.


## Step 8. Merge

Merge the PR.

The tests will run again.

Once they pass, you are ready to create and promote the release (see next step).


## Step 9. Publish the release

a. Publish the release.

* Make sure the "This is a pre-release" checkbox is NOT checked.
* Click "Publish Release".

b. Wait for workflow to complete

There's a GitHub Actions [workflow](https://github.com/StackExchange/dnscontrol/actions?query=workflow%3Arelease) which automatically builds and attaches
all 3 binaries to the release. Refresh the page after a few minutes and you'll
see dnscontrol-Darwin, dnscontrol-Linux, and dnscontrol.exe attached as assets.


## Step 10. Announce it via email

Email the release notes to the mailing list: (note the format of the Subject line and that the first line of the email is the URL of the release)

```text
To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v$VERSION

https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION

[insert the release notes here]
```

NOTE: You won't be able to post to the mailing list unless you are on
it.  [Click here to join](https://groups.google.com/forum/#!forum/dnscontrol-discuss).


## Step 11. Announce it via chat

Mention on [https://gitter.im/dnscontrol/Lobby](https://gitter.im/dnscontrol/Lobby) that the new release has shipped.

```text
ANNOUNCEMENT: dnscontrol v$VERSION has been released! https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION
```


## Step 12. Get credit!

Mention the fact that you did this release in your weekly accomplishments.

If you are at Stack Overflow:

  * Add the release to your weekly snippets
  * Run this build: `dnscontrol_embed - Promote most recent artifact into ExternalDNS repo`


# Tip: How to update modules

List out-of-date modules and update any that seem worth updating:

```bash
go install github.com/oligot/go-mod-upgrade@latest
go-mod-upgrade
go mod tidy
```

OLD WAY:

```bash
go install github.com/psampaz/go-mod-outdated@latest
go list -mod=mod -u -m -json all | go-mod-outdated -update -direct

# If any are out of date, update via:

go get module/path

# Once the updates are complete, tidy up:

go mod tidy
```
