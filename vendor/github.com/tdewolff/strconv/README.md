# Strconv [![GoDoc](http://godoc.org/github.com/tdewolff/strconv?status.svg)](http://godoc.org/github.com/tdewolff/strconv)

This package contains string conversion function and is written in [Go][1]. It is much alike the standard library's strconv package, but it is specifically tailored for the performance needs within the minify package.

For example, the floating-point to string conversion function is approximately twice as fast as the standard library, but it is not as precise.

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
