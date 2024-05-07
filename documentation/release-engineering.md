# How to build and ship a release

These are the instructions for producing a release.

GitHub Actions (GHA) will do most of the work for you. You will need to edit the draft release notes and click a button to make the release public.

Please change the version number as appropriate.  Substitute (for example)
`v4.2.0` any place you see `$VERSION` in this doc.

## Step 1. Rebuild generated files

```shell
export VERSION=v4.x.0
git checkout main
git pull
go fmt ./...
go generate ./...
go mod tidy
git status
```

There should be no modified files. If there are, check them in then start over from the beginning:

```
git checkout -b gogenerate
git commit -a -m "Update generated files for $VERSION"
```

## Step 2. Tag the commit in main that you want to release

```shell
export VERSION=v4.x.0
git checkout main
git tag -m "Release $VERSION" -a $VERSION
git push origin HEAD --tags
```

Soon after
GitHub will start an [Action](https://github.com/StackExchange/dnscontrol/actions) Workflow called "draft release" which will build all release binaries and write the draft release notes.

## Step 3. Create the release notes

The draft release notes are created for you. In this step you'll edit them.

The GHA workflow uses [GoReleaser](https://goreleaser.com/) which produces the [GitHub Release](https://github.com/StackExchange/dnscontrol/releases) with Release Notes derived from the commit history between now and the last tag.
These notes are just a draft and needs considerable editing.
These release notes are used elsewhere, in particular the email step.

Release notes style guide:

* Entries in the bullet list should be phrased in the positive: "Feature FOO now does BAR".  This is often the opposite of the related issue, which was probably phrased, "Feature FOO is broken because of BAR".
* Every item should include the ID of the issue related to the change. If there was no issue, create one and close it.
* Sort the list most important/exciting changes earlier in the list.
* Items related to a specific provider should begin with the all-caps name of the provider, such as "ROUTE53: Added support for sandwiches (#100)"
* The `Deprecation warnings` section should just copy from `README.md`.  If you change one, change it in the README too (you can make that change in this PR).

See [https://github.com/StackExchange/dnscontrol/releases](https://github.com/StackExchange/dnscontrol/releases) for examples for recent release notes and copy that style.

## Step 4. Announce it via email

Email the release notes to the mailing list: (note the format of the Subject line and that the first line of the email is the URL of the release)

```text
To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v$VERSION

https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION

[insert the release notes here]
```

{% hint style="info" %}
**NOTE**: You won't be able to post to the mailing list unless you are on
it.  [Click here to join](https://groups.google.com/g/dnscontrol-discuss).
{% endhint %}

## Step 5. Announce it via chat

Mention on [https://gitter.im/dnscontrol/Lobby](https://gitter.im/dnscontrol/Lobby) that the new release has shipped.

```text
ANNOUNCEMENT: dnscontrol v$VERSION has been released! https://github.com/StackExchange/dnscontrol/releases/tag/v$VERSION
```

## Step 6. Get credit

Mention the fact that you did this release in your weekly accomplishments.

If you are at Stack Overflow:

* Add the release to your weekly snippets

## Tip: How to bump the major version

If you bump the major version, you need to change all the source
files.  The last time this was done (v3 -> v4) these two commands
were used. They're included her for reference.

```shell
#  Make all the changes:
sed -i.bak -e 's@github.com.StackExchange.dnscontrol.v3@github.com/StackExchange/dnscontrol/v4@g' go.* $(fgrep -lri --include '*.go' github.com/StackExchange/dnscontrol/v3 *)
# Delete the backup files:
find * -name \*.bak -delete
```

## Tip: Configuring GHA integration tests

### Overview

GHA is configured to run an integration test for any provider listed in the "provider" list. However the test is skipped if the `*_DOMAIN` variable is not set. For example, the Google Cloud provider integration test is only run if `GCLOUD_DOMAIN` is set.

* Q: Where is the list of providers to run integration tests on?
* A: In `.github/workflows/pr_test.yml`: (1) the "PROVIDERS" list, (2) the `integrtests-diff2` section.

* Q: Where are non-secret environment variables stored?
* A: GHA calls them "Variables". Update them here: https://github.com/StackExchange/dnscontrol/settings/variables/actions

* Q: Where are SECRET environment variables stored?
* A: GHA calls them "Secrets". Update them here: https://github.com/StackExchange/dnscontrol/settings/secrets/actions

### How do I add a single new integration test?

1. Edit `.github/workflows/pr_test.yml`
2. Add the `FOO_DOMAIN` variable name of the provider to the "PROVIDERS" list.
3. Set the `FOO_DOMAIN` variables in GHA via https://github.com/StackExchange/dnscontrol/settings/variables/actions
4. All other variables should be stored as secrets (for consistency).  Add them to the `integration-tests` section.
Set them in GHA via https://github.com/StackExchange/dnscontrol/settings/secrets/actions

### How do I add a "bring your own keys" integration test?

Overview: You will fork the repo and add any secrets to your fork.  For security reasons you won't have access to the secrets from the main repository.

1. [Fork StackExchange/dnscontrol](https://github.com/StackExchange/dnscontrol/fork) in GitHub.

    If you already have a fork, be sure to use the "sync fork" button on the main page to sync with the upstream.

2. In your fork, set the `${DOMAIN}_DOMAIN` variable in GHA via Settings :: Secrets and variables :: Actions :: Variables.

3. In your fork, set any secrets in GHA via Settings :: Secrets and variables :: Actions :: Secrets.

5. Start a build


## Tip: How to rebuild flattener

Rebuilding flatter requires go1.17.1 and the gopherjs compiler.

Install go1.17.1:

```shell
go install golang.org/dl/go1.17.1@latest
go1.17.1 download
```

Install [GopherJS](https://github.com/gopherjs/gopherjs):

```shell
go install github.com/gopherjs/gopherjs@latest
```

Build the software:

{% hint style="info" %}
**NOTE**: GOOS can't be Darwin because GOPHERJS doesn't support it.
{% endhint %}

```shell
cd docs/flattener/
export GOPHERJS_GOROOT="$(go1.17.1 env GOROOT)"
export GOOS=linux
gopherjs build
```

## Tip: How to update modules

List out-of-date modules and update any that seem worth updating:

```shell
go install github.com/oligot/go-mod-upgrade@latest
go-mod-upgrade
go mod tidy
```

OLD WAY:

```shell
go install github.com/psampaz/go-mod-outdated@latest
go list -mod=mod -u -m -json all | go-mod-outdated -update -direct

# If any are out of date, update via:

go get module/path

# Once the updates are complete, tidy up:

go mod tidy
```
