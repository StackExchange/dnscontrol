# How to build and ship a release

Here are my notes from producing the v0.1.5 release.  Change the version number as appropriate.

## Step 1.  Run the integration tests

If you are at StackOverflow, this is in TC as "DNS > Integration Tests".

## Step 2. Bump the verison number

Edit the "Version" variable in `main.go` and commit.

```
vi main.go
git commit -m'Release v1.5' main.go 
git tag v0.1.5
git push origin tag v0.1.5
```

## Step 3. Make the draft release.

[On github.com, click on "Draft a new release"](https://github.com/StackExchange/dnscontrol/releases/new)

Pick the v0.1.5 tag

Release title: Release v0.1.5

Fill in the text box with something friendly like, "So many new features!" then make a bullet list of major new functionality.

Review the git log using this command:

    git log v0.1.5...v0.1.0

Create the binaries and attach them to the release:

    go run build/build.go 

NOTE: This command creates binaries with the version number and git hash embedded. It also builds the releases for all supported platforms (i.e. creates a .exe for Windows even if you are running on Linux.  Isn't Go amazing?)

WARNING: if there are unchecked in files, the version will have "dirty" added.

This is what it looks like when you did it right:

```
$ ./dnscontrol-Darwin version
dnscontrol 0.1.5 ("6fdf78997815055bbe119c0116c9e2d60310a515[dirty]") built 24 Aug 17 11:26 EDT
```

This is what it looks like when there was a file that should have been checked in:

```
$ ./dnscontrol-Darwin version
dnscontrol 0.1.5 ("6fdf78997815055bbe119c0116c9e2d60310a515[dirty]") built 24 Aug 17 11:27 EDT
                                                            ^^^^^
                                                            ^^^^^ See?
                                                            ^^^^^
```

## Step 4. Attach the binaries and release.

Drag and drop binaries into the web form.

Submit the release.

## Step 5. Announce it

Email the mailing list: (note the format of the Subject line and that the first line of the email is the URL of the release)

```
To: dnscontrol-discuss@googlegroups.com
Subject: New release: dnscontrol v0.1.5

https://github.com/StackExchange/dnscontrol/releases/tag/v0.1.5

So many new providers and features! Plus, a new testing framework that makes it easier to add big features without fear of breaking old ones.

* list
* of 
* major
* changes
```


## Step 6. Get credit!

Mention the fact that you did this release in your weekly accomplishments.
