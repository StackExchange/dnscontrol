# DNSControl and Internationalized domain name

This is my proposal for how to make IDNs work better in DNSControl.
Basically, the UI will accept any format.  Early in the process
DNSControl will store labels/domains data 4 ways: As received from the
user (downcased), ASCII, Unicode, and in a "display" format that shows
both. The conversions already done ahead of time, providers can
access whatever format they need.  Output from the main program will
use the "display" format when possible.

# Problem Statement

DNSControl doesn't handle internationalized domain names (IDNs) very
well.  Coverage is uneven: They work better in some providers than
others. There are bugs and inconsistencies. Writing a provider that
handles IDNs properly requires doing most of the work in the provider
itself, which means every provider maintainer must be an expert in
IDNs, which is unreasonable.

# Background:

RFC 3490 recommends how applications should handle IDNs. My summary:
(1) the UI should accept a mix of Unicode and ASCII domains/labels.
(2) internally translate everything to ASCII (punycode) and do all
processing in that format, (3) when displaying output, display it as
the user input it, or Unicode, or ASCII, or give users a choice.

* IDNA: Internationalizing Domain Names in Applications
* IDNs: Internationalized domain names
* ACE Prefix: The `xn--` that means "Puny code follows"
* ASCII: A label or domain is output as ASCII with ACE prefix if needed.
* Unicode: A label or domain is output as Unicode.

Proposed Outcome:

1. Users should be able to input domains and labels in either ASCII (with ACE prefix if needed), Unicode, or a mix.  This holds for input via `dnsconfig.js` (domain names, labels, and targets); as well as flags such as `--domains`.
2a. Output should be Unicode or both with the ASCII being in parenthesis. Example: `рф.com (xn--p1ai.com)`
2b. Or maybe the reverse? Example: `xn--p1ai.com (рф.com)`
3. DNSControl's main code should create a "paved path" for providers to make it easy for them to do the right thing. It should be easier to do the right thing than the wrong thing.  The default (i.e. "lazy") path should result in the behavior we desire.

Here are some example outputs:

NOTE: Feedback needed!  Do you prefer "a" or "b"?  Is there an even better format I should consider?  Should we use `{}` instead of `()`?

Example 1a: CREATE unicode (ascii)

```
#1: + CREATE foo.рф.com (foo.xn--p1ai.com) MX 10 рф.com. (xn--p1ai.com.) (ttl=14400)
```

Example 1b: CREATE ascii (unicode)

```
#2: + CREATE foo.xn--p1ai.com (foo.рф.com) MX 10 xn--p1ai.com. (рф.com.) (ttl=14400)
```

Example 3a: MODIFY ascii (unicode) -> ascii (unicode)

```
#3: ± MODIFY foo.xn--p1ai.com (foo.рф.com) (10 xn--p1ai.com. (рф.com.) ttl=14400) -> (10 foo.xn--p1ai.com. (foo.рф.com.) ttl=14400)
```

Example 3b: MODIFY unicode (ascii) -> unicode (ascii)

```
#4: ± MODIFY foo.рф.com (foo.xn--p1ai.com) (10 рф.com. (xn--p1ai.com.) ttl=14400) -> (10 foo.рф.com. (foo.xn--p1ai.com.) ttl=14400)
```

Example 3c: MODIFY ascii

```
#5: ± MODIFY foo.рф.com (10 рф.com. ttl=14400) -> (10 foo.рф.com. ttl=14400)
```

Example 3d: MODIFY unicode

```
#6: ± MODIFY foo.xn--p1ai.com (10 xn--p1ai.com. ttl=14400) -> (10 foo.xn--p1ai.com. ttl=14400)
```

NOTE: When the ASCII and Unicode versions are the same (i.e.
everything is plain ASCII) the display would appear as before:

```
#7: + CREATE foo1.example.com MX 10 mxfoo.example.com. (ttl=14400)
#8: ± MODIFY foo2.example.com (10 example.com. ttl=14400) -> (10 foo.example.com. ttl=14400)
```

these examples are similar, but the targets are unicode:

```
#9: + CREATE foo3.example.com MX 10 xn--p1ai.com. (рф.com.) (ttl=14400)
#10: ± MODIFY foo4.example.com (10 xn--p1ai.com. (рф.com.) ttl=14400) -> (10 foo.example.com. ttl=14400)
#11: ± MODIFY foo5.example.com (10 example.com. ttl=14400) -> (10 xn--p1ai.com. (рф.com.) ttl=14400)
```

Now here are the same examples with `()` changed to `{}`:

```
#1: + CREATE foo.рф.com {foo.xn--p1ai.com} MX 10 рф.com. {xn--p1ai.com.} (ttl=14400)
#2: + CREATE foo.xn--p1ai.com {foo.рф.com} MX 10 xn--p1ai.com. {рф.com.} (ttl=14400)
#3: ± MODIFY foo.xn--p1ai.com {foo.рф.com} (10 xn--p1ai.com. {рф.com.} ttl=14400) -> (10 foo.xn--p1ai.com. {foo.рф.com.} ttl=14400)
#4: ± MODIFY foo.рф.com {foo.xn--p1ai.com} (10 рф.com. {xn--p1ai.com.} ttl=14400) -> (10 foo.рф.com. {foo.xn--p1ai.com.} ttl=14400)
#5: ± MODIFY foo.рф.com (10 рф.com. ttl=14400) -> (10 foo.рф.com. ttl=14400)
#6: ± MODIFY foo.xn--p1ai.com (10 xn--p1ai.com. ttl=14400) -> (10 foo.xn--p1ai.com. ttl=14400)
#7: + CREATE foo1.example.com MX 10 mxfoo.example.com (ttl=14400)
#8: ± MODIFY foo2.example.com (10 example.com. ttl=14400) -> (10 foo.example.com. ttl=14400)
#9: + CREATE foo3.example.com MX 10 xn--p1ai.com. {рф.com.} (ttl=14400)
#10: ± MODIFY foo4.example.com (10 xn--p1ai.com. {рф.com.} ttl=14400) -> (10 foo.example.com. ttl=14400)
#11: ± MODIFY foo5.example.com (10 example.com. ttl=14400) -> (10 xn--p1ai.com. {рф.com.} ttl=14400)
```

Now here are the same examples with `()` changed to `⟬⟭`:

```
#1: + CREATE foo.рф.com ⟬foo.xn--p1ai.com⟭ MX 10 рф.com. ⟬xn--p1ai.com.⟭ (ttl=14400)
#2: + CREATE foo.xn--p1ai.com ⟬foo.рф.com⟭ MX 10 xn--p1ai.com. ⟬рф.com.⟭ (ttl=14400)
#3: ± MODIFY foo.xn--p1ai.com ⟬foo.рф.com⟭ (10 xn--p1ai.com. ⟬рф.com.⟭ ttl=14400) -> (10 foo.xn--p1ai.com. ⟬foo.рф.com.⟭ ttl=14400)
#4: ± MODIFY foo.рф.com ⟬foo.xn--p1ai.com⟭ (10 рф.com. ⟬xn--p1ai.com.⟭ ttl=14400) -> (10 foo.рф.com. ⟬foo.xn--p1ai.com.⟭ ttl=14400)
#5: ± MODIFY foo.рф.com (10 рф.com. ttl=14400) -> (10 foo.рф.com. ttl=14400)
#6: ± MODIFY foo.xn--p1ai.com (10 xn--p1ai.com. ttl=14400) -> (10 foo.xn--p1ai.com. ttl=14400)
#7: + CREATE foo1.example.com MX 10 mxfoo.example.com (ttl=14400)
#8: ± MODIFY foo2.example.com (10 example.com. ttl=14400) -> (10 foo.example.com. ttl=14400)
#9: + CREATE foo3.example.com MX 10 xn--p1ai.com. ⟬рф.com.⟭ (ttl=14400)
#10: ± MODIFY foo4.example.com (10 xn--p1ai.com. ⟬рф.com.⟭ ttl=14400) -> (10 foo.example.com. ttl=14400)
#11: ± MODIFY foo5.example.com (10 example.com. ttl=14400) -> (10 xn--p1ai.com. ⟬рф.com.⟭ ttl=14400)
```

Now here are the same examples with `()` changed to `❮❯`:

```
#1: + CREATE foo.рф.com ❮foo.xn--p1ai.com❯ MX 10 рф.com. ❮xn--p1ai.com.❯ (ttl=14400)
#2: + CREATE foo.xn--p1ai.com ❮foo.рф.com❯ MX 10 xn--p1ai.com. ❮рф.com.❯ (ttl=14400)
#3: ± MODIFY foo.xn--p1ai.com ❮foo.рф.com❯ (10 xn--p1ai.com. ❮рф.com.❯ ttl=14400) -> (10 foo.xn--p1ai.com. ❮foo.рф.com.❯ ttl=14400)
#4: ± MODIFY foo.рф.com ❮foo.xn--p1ai.com❯ (10 рф.com. ❮xn--p1ai.com.❯ ttl=14400) -> (10 foo.рф.com. ❮foo.xn--p1ai.com.❯ ttl=14400)
#5: ± MODIFY foo.рф.com (10 рф.com. ttl=14400) -> (10 foo.рф.com. ttl=14400)
#6: ± MODIFY foo.xn--p1ai.com (10 xn--p1ai.com. ttl=14400) -> (10 foo.xn--p1ai.com. ttl=14400)
#7: + CREATE foo1.example.com MX 10 mxfoo.example.com (ttl=14400)
#8: ± MODIFY foo2.example.com (10 example.com. ttl=14400) -> (10 foo.example.com. ttl=14400)
#9: + CREATE foo3.example.com MX 10 xn--p1ai.com. ❮рф.com.❯ (ttl=14400)
#10: ± MODIFY foo4.example.com (10 xn--p1ai.com. ❮рф.com.❯ ttl=14400) -> (10 foo.example.com. ttl=14400)
#11: ± MODIFY foo5.example.com (10 example.com. ttl=14400) -> (10 xn--p1ai.com. ❮рф.com.❯ ttl=14400)
```


# Design

In general, DNSControl will store domains and labels in multiple
formats: (1) in the original
format the user specified (downcased), in ASCII, and in Unicode, and
in a format useful for displaying to users.  This way
providers do not have to do conversions.

When the user used Unicode:

* Original: рф.com
* ASCII: xn--p1ai.com
* Unicode: рф.com
* Display: xn--p1ai.com (рф.com)

When the user used ASCII:

* Original: xn--p1ai.com
* ASCII: xn--p1ai.com
* Unicode: рф.com
* Display: xn--p1ai.com (рф.com)

NOTE: User input is downcased.  If the user input is `D('xn--P1AI.COM')` the Original field would be `xn--p1ai.com` and so on.

Memory usage will be minimized by using Go's slices.  In the above
example, the Display string would be generated first, the others would
be slices of that string.


```
models.DomainConfig:

   .Name: the name from D() after downcased via unicode.ToLower()
   .NameASCII: The name stored after calling ToASCII() (with ACE prefix if any Unicode chars are present)
   .NameUnicode: The name stored after calling ToUnicode()
   .NameDisplay: if .NameASCII != .NameUnicode, store as "ascii (unicode)"
     Otherwise, the value is the same as .NameASCII

models.Nameserver:
  .Name will also be stored 4 ways, similar to models.DomainConfig

models.RecordConfig:

   .Name: the name downcased via unicode.ToLower()
   .NameASCII: The name stored after calling ToASCII() (with ACE prefix if any Unicode chars are present)
   .NameUnicode: The name stored after calling ToUnicode()
   .NameDisplay: if .NameASCII != .NameUnicode, store as "ascii (unicode)"
     Otherwise, the value is the same as .NameASCII

   .NameFQDN: the name downcased via unicode.ToLower()
   .NameFQDNASCII: The name stored after calling ToASCII() (with ACE prefix if any Unicode chars are present)
   .NameFQDNUnicode: The name stored after calling ToUnicode()
   .NameFQDNDisplay: if .NameFQDNASCII != .NameFQDNUnicode, store as "ascii (unicode)"
     Otherwise, the value is the same as .NameFQDNASCII

   .SubDomain: will be passed through unicode.ToLower() then ToASCII()

models.target:
   GetTargetField() returns .target
   GetTargetFieldASCII() returns .targetASCII
   GetTargetFieldUnicode() returns .targetUnicode
   GetTargetFieldDisplay() returns .targetDisplay
models.RecordConfig:
   .R53Alias: will be passed through unicode.ToLower() then ToASCII()
   .AzureAlias: will be passed through unicode.ToLower() then ToASCII()
```

# Code changes

Since the labels/domains have been pre-converted, providers no longer
need to do the conversion themselves.

1. After compiling `dnsconfig.js`, but before calling
   ValidateAndNormalizeConfig(), the function NormalizeIDN() will be
   called. NormaliIDN() will do all the conversions listed above
   (.Name, .NameASCII, .NameUnicode, .NameDisplay and so on)

2. All calls to `dc.Punycode()` will be removed. They are no longer
   needed.

3. Providers should no longer need to do their own conversions.
Calls to the idna module currently exist in
domainnameshop, cloudflare, vultr, and  hostingde.  These will require
special attention.

4. Testing will be required for all providers.  A PR with a checklist
   will be used to let provider maintainers check in on their tests.
   However, provider maintainers that do not check in within 3 (?)
   weeks will not block the PR merge. We do this because (1) not
   everyone has an IDN to test with, (2) old code should work as well
   as before (bugs and all!).

5. Documentation updates: Not sure what updates are need but
suggestions welcome!

# Call for volunteers!

I am not an expert in IDN.  If someone would like to help out with
testing, coding, and so on, I would greatly appreciate it!
