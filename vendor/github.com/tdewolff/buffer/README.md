# Buffer [![GoDoc](http://godoc.org/github.com/tdewolff/buffer?status.svg)](http://godoc.org/github.com/tdewolff/buffer)

This package contains several buffer types used in https://github.com/tdewolff/parse for example.

## Installation
Run the following command

	go get github.com/tdewolff/buffer

or add the following import and run the project with `go get`
``` go
import "github.com/tdewolff/buffer"
```

## Reader
Reader is a wrapper around a `[]byte` that implements the `io.Reader` interface. It is a much thinner layer than `bytes.Buffer` provides and is therefore faster.

## Writer
Writer is a buffer that implements the `io.Writer` interface. It is a much thinner layer than `bytes.Buffer` provides and is therefore faster. It will expand the buffer when needed.

The reset functionality allows for better memory reuse. After calling `Reset`, it will overwrite the current buffer and thus reduce allocations.

## Lexer
Lexer is a read buffer specifically designed for building lexers. It keeps track of two positions: a start and end position. The start position is the beginning of the current token being parsed, the end position is being moved forward until a valid token is found. Calling `Shift` will collapse the positions to the end and return the parsed `[]byte`.

Moving the end position can go through `Move(int)` which also accepts negative integers. One can also use `Pos() int` to try and parse a token, and if it fails rewind with `Rewind(int)`, passing the previously saved position.

`Peek(int) byte` will peek forward (relative to the end position) and return the byte at that location. `PeekRune(int) (rune, int)` returns UTF-8 runes and its length at the given **byte** position. Upon an error `Peek` will return `0`, the **user must peek at every character** and not skip any, otherwise it may skip a `0` and panic on out-of-bounds indexing.

`Lexeme() []byte` will return the currently selected bytes, `Skip()` will collapse the selection. `Shift() []byte` is a combination of `Lexeme() []byte` and `Skip()`.

When the passed `io.Reader` returned an error, `Err() error` will return that error even if not at the end of the buffer.

## StreamLexer
StreamLexer behaves like Lexer but uses a buffer pool to read in chunks from `io.Reader`, retaining old buffers in memory that are still in use, and re-using old buffers otherwise. Calling `Free(n int)` frees up `n` bytes from the internal buffer(s). It holds an array of buffers to accommodate for keeping everything in-memory. Calling `ShiftLen() int` returns the number of bytes that have been shifted since the previous call to `ShiftLen`, which can be used to specify how many bytes need to be freed up from the buffer. If you don't need to keep returned byte slices around, call `Free(ShiftLen())` after every `Shift` call.

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
