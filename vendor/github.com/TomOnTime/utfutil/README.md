# utfutil

Utilities to make it easier to read text encoded as UTF-16.

## Dealing with UTF-16 files from Windows.

Ever have code that worked for years until you received a file from a MS-Windows system that just didn't work at all?  Looking at a hex dump you realize every other byte is \0.  WTF?  No, UTF.  More specifically UTF-16LE with an optional BOM.

What does all that mean?  Well, first you should read ["The Absolute Minimum Every Software Developer Absolutely, Positively Must Know About Unicode and Character Sets (No Excuses!)"](http://www.joelonsoftware.com/articles/Unicode.html) by Joel Spolsky.

Now you are an expert.  You can spend an afternoon trying to figure out how the heck to put all that together and use `golang.org/x/text/encoding/unicode` to decode UTF-16LE.  However I've already done that for you. Now you can take the easy way out change ioutil.ReadFile() to utfutil.ReadFile().  Everything will just work.

### utfutil.ReadFile() is the equivalent of ioutil.ReadFile()

OLD: Works with UTF8 and ASCII files:

```
		data, err := ioutil.ReadFile(filename)
```

NEW: Works if someone gives you a Windows UTF-16LE file occasionally but normally you are processing UTF8 files:

```
		data, err := utfutil.ReadFile(filename, utfutil.UTF8)
```

### utfutil.OpenFile() is the equivalent of os.Open().

OLD: Works with UTF8 and ASCII files:

```
		data, err := os.Open(filename)
```

NEW: Works if someone gives you a file with a BOM:

```
		data, err := utfutil.OpenFile(filename, utfutil.HTML5)
```

### utfutil.NewScanner() is for reading files line-by-line

It works like os.Open():

```
		s, err := utfutil.NewScanner(filename, utfutil.HTML5)
```


## Encoding hints:

What's that second argument all about?

Since it is impossible to guess 100% correctly if there is no BOM,
the functions take a 2nd parameter of type "EncodingHint" where you
specify the default encoding for BOM-less files.

```
UTF8        No BOM?  Assume UTF-8
UTF16LE     No BOM?  Assume UTF 16 Little Endian
UTF16BE     No BOM?  Assume UTF 16 Big Endian
WINDOWS = UTF16LE   (i.e. a reasonable guess if file is from MS-Windows)
POSIX   = UTF8      (i.e. a reasonable guess if file is from Unix or Unix-like systems)
HTML5   = UTF8      (i.e. a reasonable guess if file is from the web)
```

## Future Directions

If someone writes a golang equivalent of uchatdet, I'll add a hint
called "AUTO" which uses it. That would be awesome. Volunteers?
