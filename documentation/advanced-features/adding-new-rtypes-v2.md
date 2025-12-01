# Creating new DNS Resource Types (rtypes) (v4.28 and later)

Everyone is familiar with A, AAAA, CNAME, NS and other Rtypes.
However there are new record types being added all the time.
Each new record type requires special handling by
DNSControl.

Version v4.28.0 greatly simplified how to add new record types. As
a demonstration of this new method it added the "RP" type and
ported the existing "CLOUDFLAREAPI_SINGLE_REDIRECT" type.  All
other records still use the old method. The old and new
methods co-exist, though eventually we hope to migrate
everything to the new method.

# What's new?

* OLD: the RecordConfig struct keeps getting larger as new record types require more fields.
* NEW: the RecordConfig struct has a pointer to a struct describing the fields of the record.
* Benefit: Saves memory.

* OLD: helpers.js performs validation, executes builders, etc.  Since we don't have a test framework for Javascript, this is brittle and difficult to debug.
* NEW: helpers.js packs up the fields, whatever they are, and handes them off to Go code for processing. The Go test framework is available.
* Benefit: More testable, easier to  develop as you don't need to know 2 languages.

* OLD: Critical things like IDN processing, normalization (downcasing), and validation happen late in the pipeline by `pkg/normalize`.
* NEW: The factory that creates a RecordConfig performs all that.
* Benefit: No code has to be concerned with "has this RecordConfig been normalized/IDN-ized yet?". This makes it easier to write and debug code.

* OLD: Code that affects a Record Type is splattered all over the code base.
* NEW: Code related to a Record Type is all in one file.
* Benefit: Centralize all concerns about a record type in one file (with some exceptions that we hope to fix eventually).

# Overview

There are two parts to adding a new Record Type (rtype). First we activate it in the parser for dnsconfig.js. Then we update
any provider to be aware of the rtype.

Activate it in dnsconfig.js:

1. Update `pkg/js/helpers.js` (add just 1 line!)
2. Create a file in `pkg/rtype` (for example, `pkg/rtype/rp.go`) with a parser.

Update providers to be aware of the rtype:

1. Update the toRC() function (whatever it may be called).
1. Update the toNative() and any create/delete/change functions.


# Updating the parser

In these examples, the new type will be called `THING` (or `thing`).

Step 1: Update helpers.js

At the end of the file, add a single line named after the record type.

```
var RP = rawrecordBuilder('RP');
```

In this example, the first `RP` is the name of the function that users will
type in dnsconfig.js.  For example, `A("label", "10.2.3.4"),` (though the "A"
record type hasn't been ported to the new system yet).

Step 2: Create the parser

Create a file in `pkg/rtype/thing.go` named after the record type (all lowercase).

Copy `pkg/rtype/rp.go` as it is a good prototype.

Step 2a: Update init()

Update the init function to register your new type. That is, change `RP` to `THING`.

```
func init() {
	rtypecontrol.Register(&THING{})
}
```

STep 2b: Create the struct

Create a struct that will store the fields of this rtype.

If this is a standard rtype, borrow from miekg/dns:

```
type THING struct {
	dns.THING
}
```

If this is not a standard type, list the fields:

```
type THING struct {
	ThingField1 uint32
    ThingField2 string
    ThingField3 string
	ThingField4 uint32
}
```

Step 2c: Create Name()

Create a function `Name()` that outputs the
name of the type.

```
func (handle *THING) Name() string {
	return "THING"
}
```

Step 2d: Create FromArgs

The "FromArgs" function receives an array of `any` which can contain any type.  The "PaveArgs" function
will convert them to the types you need. For example, it will convert numbers to strings, or strings to numbers.

* "s": Convert to string
* "i": Convert to int16

`args[0]` is the label. You can skip it as that is already processed for you.  (If you want
to modify the label, see cfsingle.go as an example of how to do that.)

If THING takes 4 parameters (2 ints and 2 strings), you might pave the arguments as follows:

```
	if err := rtypecontrol.PaveArgs(args[1:], "issi"); err != nil {
```

Now you can fill the struct as needed:

```
	fields := &THING{
		dns.THING{
			ThingField1: args[1].(uint16),
			ThingField2:  args[2].(string),
			ThingField3:  args[3].(string),
			ThingField4:  args[4].(uint16),
		},
	}
```

Now call FromStruct to finsh up.

```
	return handle.FromStruct(dc, rec, args[0].(string), fields)
```

FromStruct does many things:

1. Installs the struct in .F
2. Converts any names to IDN and Unicode equivalents for future reference.
2. Performs any validation (for example, if ThingField1 has to be between 0 and 999)
2. Generate the .ZonefilePartial field: This is what the record outputs in a zonefile.
2. Generate the .Comparable field: This is an opaque string used to compare two RecordConfigs. If the strings are not an exact match, they are considered "not equal".

Create the CopyToLegacyFields function

This updates any of the legacy fields. The most important is the .target field, which we
usually store a copy of the .ZonefilePartial.

When we migrate other rtypes this will populate the legacy RecordConfig fields. For example, when
we migrate `SRV`, this function will populate the `Srv*` fields.  Then, eventually, we'll remove
those legacy fields.


TODO:

* `js/parse_test`
* Run the integration tests for BIND
* Write documentation

# Update providers

When a provider needs to create a THING, they have two choices

If you have the fields already in variables of the right type, use NewRecordConfigFromStruct:

```
rec, err = rtypecontrol.NewRecordConfigFromStruct(name, ttl, "THING", rtype.THING{a, b, c, d}, dc)
```

If you have the fields in variables that are strings that need to be converted, use `NewRecordConfigFromRaw()`:

```
rec, err := rtypecontrol.NewRecordConfigFromRaw("CF_REDIRECT", ttl, []any{pattern, target}, dc)
```

There's a good chance you know the zoneName but don't have a complete models.DomainConfig(). That's ok.
As long as you know the zone name, we can fake it: (this is a hack we'll figure out how to eliminate eventually).

```
dc := models.MakeFakeDomainConfig(zoneName)
```

# Tips for "builders"

A "builder" is a function that create other records.  For example, SPF_BUILDER() creates a `TXT()` record.

A good example of a builder is `providers/cloudflare/rtypes/cfsingleredirect/cfredirect.go`

Simply do the processing you need, then create the resulting rtype you want:

In cfredirect.go, CF_REDIRECT is a builder that generates a CLOUDFLAREAPI_SINGLE_REDIRECT, represented by the SingleRedirectConfig struct:

```
	sr := SingleRedirectConfig{}
	rec.Type = sr.Name() // This record is now a CLOUDFLAREAPI_SINGLE_REDIRECT
	err = sr.FromArgs(dc, rec, []any{name, code, srWhen, srThen})
	if err != nil {
		return err
	}
```
