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

## Shifter
Shifter is a read buffer specifically for building lexers. It reads in chunks from an `io.Reader` and allows to keep track two positions: the start and end position. The start position is the beginning of the current token being parsed, the end position is being moved forward until a valid token is found. Calling `Shift` will collapse the positions to the end and return the parsed `[]byte`.

Moving the end position can go through `Move(int)` which also accepts negative integers or `MoveTo(int)` where the integer will be the new length of the selected bytes. `MoveTo(int)` is useful when you saved a previous position through `Pos() int` and want to return to that position.

`Peek(int) byte` will peek forward (relative to the end position, ie. the position set with Move/MoveTo) and return the byte at that location. `PeekRune(int) (rune, int)` returns UTF-8 runes and its length at the given **byte** position. Consecutive calls to Peek **may invalidate previously returned byte slices**. So if you need to use the content of a byte slice after the next call to `Peek(int) byte`, it needs to be copied in principal (see exception below).

`Bytes() []byte` will return the currently selected bytes, `Skip()` will collapse the selection. `Shift() []byte` is a combination of `Bytes() []byte` and `Skip()`.

When the internal `io.Reader` returned an error, `Err() error` will return that error (even if subsequent peeks  are still possible). If `Peek(int) byte` returns `0` when an error occurred. `IsEOF() bool` is a faster alternative than `Err() == io.EOF`, if it returns true it means the internal buffer will not be reallocated/overwritten. So returned byte slices need not be copied for use after subsequent `Peek(int) byte` calls. When the `io.Reader` provides the `Bytes() []byte` function (which `Reader` does in this package), it will use that buffer instead and thus `IsEOF()` returns always `true` (ie. copying returned slices is not needed).

## Lexer
Lexer is an improvement over Shifter in that it does not need the returned byte slices to be copied. Instead you can call `ShiftLen() int`, which returns the number of bytes that have been shifted since the previous call to `ShiftLen`, and use that to specify how many bytes need to be freed up from the buffer. Calling `Free(n int)` frees up `n` bytes from the internal buffer(s). It holds an array of buffers to accomodate for keeping everything in-memory. If you don't need to keep returned byte slices around, call `Free(ShiftLen())` after every `Shift` call.

The `MoveTo(int)` function has been renamed to `Rewind(int)` to fit its meaning better. Also `Bytes() []byte` has been renamed to `Lexeme() []byte` for the same reason.

## License
Released under the [MIT license](LICENSE.md).

[1]: http://golang.org/ "Go Language"
