# utfutil

Utilities to make it easier to read text encoded as UTF-16.

## Dealing with UTF-16 files you receive from Windows.

Have you encountered this situation?  Code that has worked for years
suddenly breaks.  It turns out someone tried to use it with a file
that came from a MS-Windows system. Now this perfectly good code stops
working.

Looking at a hex dump you realize every other byte is \0.  WTF?
No, UTF.  More specifically UTF-16LE with an optional BOM.

What does all that mean?  Well, first you should read ["The Absolute Minimum Every Software Developer Absolutely, Positively Must Know About Unicode and Character Sets (No Excuses!)"](http://www.joelonsoftware.com/articles/Unicode.html) by Joel Spolsky.

Now you understand what the problem is, but how do you fix it?
Well, you can spend a week trying to figure out how to use
`golang.org/x/text/encoding/unicode` and you'll be able to
decode UTF-16LE files. (No offense to the authors of that
module. It is a fantastic module but if you aren't already
an expert in Unicode encoding, it is pretty difficult to use.)

If you don't have a week, you can just use this module.
Take the easy way out!  Just change `ioutil.ReadFile()` to
`utfutil.ReadFile()`.
Everything will just work.

The goal of `utfutl` is to provide replacement functions
that magically do the right thing. There is a demo
program that shows how to use it called [catutf](https://github.com/TomOnTime/utfutil/blob/master/catutf/main.go).


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

What's that second argument all about?    utfutil.UTF8?  utfutil.HTML5?

If a file has no BOM, it is impossible to guess the file encoding with
100% accuracy.  Therefore, the 2nd parameter is an
"EncodingHint" that specifies what to assume for BOM-less files.

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
