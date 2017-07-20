// Package json is a JSON parser following the specifications at http://json.org/.
package json // import "github.com/tdewolff/parse/json"

import (
	"errors"
	"io"
	"strconv"

	"github.com/tdewolff/buffer"
)

// ErrBadComma is returned when an unexpected comma is encountered.
var ErrBadComma = errors.New("unexpected comma character outside an array or object")

// ErrNoComma is returned when no comma is present between two values.
var ErrNoComma = errors.New("expected comma character or an array or object ending")

// ErrBadObjectKey is returned when the object key is not a quoted string.
var ErrBadObjectKey = errors.New("expected object key to be a quoted string")

// ErrBadObjectDeclaration is returned when the object key is not followed by a colon character.
var ErrBadObjectDeclaration = errors.New("expected colon character after object key")

// ErrBadObjectEnding is returned when an unexpected right brace is encountered.
var ErrBadObjectEnding = errors.New("unexpected right brace character")

// ErrBadArrayEnding is returned when an unexpected right bracket is encountered.
var ErrBadArrayEnding = errors.New("unexpected right bracket character")

////////////////////////////////////////////////////////////////

// GrammarType determines the type of grammar
type GrammarType uint32

// GrammarType values.
const (
	ErrorGrammar GrammarType = iota // extra grammar when errors occur
	WhitespaceGrammar
	LiteralGrammar
	NumberGrammar
	StringGrammar
	StartObjectGrammar // {
	EndObjectGrammar   // }
	StartArrayGrammar  // [
	EndArrayGrammar    // ]
)

// String returns the string representation of a GrammarType.
func (gt GrammarType) String() string {
	switch gt {
	case ErrorGrammar:
		return "Error"
	case WhitespaceGrammar:
		return "Whitespace"
	case LiteralGrammar:
		return "Literal"
	case NumberGrammar:
		return "Number"
	case StringGrammar:
		return "String"
	case StartObjectGrammar:
		return "StartObject"
	case EndObjectGrammar:
		return "EndObject"
	case StartArrayGrammar:
		return "StartArray"
	case EndArrayGrammar:
		return "EndArray"
	}
	return "Invalid(" + strconv.Itoa(int(gt)) + ")"
}

////////////////////////////////////////////////////////////////

// State determines the current state the parser is in.
type State uint32

// State values.
const (
	ValueState State = iota // extra token when errors occur
	ObjectKeyState
	ObjectValueState
	ArrayState
)

// String returns the string representation of a State.
func (state State) String() string {
	switch state {
	case ValueState:
		return "Value"
	case ObjectKeyState:
		return "ObjectKey"
	case ObjectValueState:
		return "ObjectValue"
	case ArrayState:
		return "Array"
	}
	return "Invalid(" + strconv.Itoa(int(state)) + ")"
}

////////////////////////////////////////////////////////////////

// Parser is the state for the lexer.
type Parser struct {
	r     *buffer.Lexer
	state []State
	err   error

	needComma bool
}

// NewParser returns a new Parser for a given io.Reader.
func NewParser(r io.Reader) *Parser {
	return &Parser{
		r:     buffer.NewLexer(r),
		state: []State{ValueState},
	}
}

// Err returns the error encountered during tokenization, this is often io.EOF but also other errors can be returned.
func (p Parser) Err() error {
	err := p.r.Err()
	if err != nil {
		return err
	}
	return p.err
}

// Next returns the next Grammar. It returns ErrorGrammar when an error was encountered. Using Err() one can retrieve the error message.
func (p *Parser) Next() (GrammarType, []byte) {
	p.r.Free(p.r.ShiftLen())

	p.moveWhitespace()
	c := p.r.Peek(0)
	state := p.state[len(p.state)-1]
	if c == ',' {
		if state != ArrayState && state != ObjectKeyState {
			p.err = ErrBadComma
			return ErrorGrammar, nil
		}
		p.r.Move(1)
		p.moveWhitespace()
		p.needComma = false
		c = p.r.Peek(0)
	}
	p.r.Skip()

	if p.needComma && c != '}' && c != ']' && c != 0 {
		p.err = ErrNoComma
		return ErrorGrammar, nil
	} else if c == '{' {
		p.state = append(p.state, ObjectKeyState)
		p.r.Move(1)
		return StartObjectGrammar, p.r.Shift()
	} else if c == '}' {
		if state != ObjectKeyState {
			p.err = ErrBadObjectEnding
			return ErrorGrammar, nil
		}
		p.needComma = true
		p.state = p.state[:len(p.state)-1]
		if p.state[len(p.state)-1] == ObjectValueState {
			p.state[len(p.state)-1] = ObjectKeyState
		}
		p.r.Move(1)
		return EndObjectGrammar, p.r.Shift()
	} else if c == '[' {
		p.state = append(p.state, ArrayState)
		p.r.Move(1)
		return StartArrayGrammar, p.r.Shift()
	} else if c == ']' {
		p.needComma = true
		if state != ArrayState {
			p.err = ErrBadArrayEnding
			return ErrorGrammar, nil
		}
		p.state = p.state[:len(p.state)-1]
		if p.state[len(p.state)-1] == ObjectValueState {
			p.state[len(p.state)-1] = ObjectKeyState
		}
		p.r.Move(1)
		return EndArrayGrammar, p.r.Shift()
	} else if state == ObjectKeyState {
		if c != '"' || !p.consumeStringToken() {
			p.err = ErrBadObjectKey
			return ErrorGrammar, nil
		}
		n := p.r.Pos()
		p.moveWhitespace()
		if c := p.r.Peek(0); c != ':' {
			p.err = ErrBadObjectDeclaration
			return ErrorGrammar, nil
		}
		p.r.Move(1)
		p.state[len(p.state)-1] = ObjectValueState
		return StringGrammar, p.r.Shift()[:n]
	} else {
		p.needComma = true
		if state == ObjectValueState {
			p.state[len(p.state)-1] = ObjectKeyState
		}
		if c == '"' && p.consumeStringToken() {
			return StringGrammar, p.r.Shift()
		} else if p.consumeNumberToken() {
			return NumberGrammar, p.r.Shift()
		} else if p.consumeLiteralToken() {
			return LiteralGrammar, p.r.Shift()
		}
	}
	return ErrorGrammar, nil
}

// State returns the state the parser is currently in (ie. which token is expected).
func (p *Parser) State() State {
	return p.state[len(p.state)-1]
}

////////////////////////////////////////////////////////////////

/*
The following functions follow the specifications at http://json.org/
*/

func (p *Parser) moveWhitespace() {
	for {
		if c := p.r.Peek(0); c != ' ' && c != '\t' && c != '\r' && c != '\n' {
			break
		}
		p.r.Move(1)
	}
}

func (p *Parser) consumeLiteralToken() bool {
	c := p.r.Peek(0)
	if c == 't' && p.r.Peek(1) == 'r' && p.r.Peek(2) == 'u' && p.r.Peek(3) == 'e' {
		p.r.Move(4)
		return true
	} else if c == 'f' && p.r.Peek(1) == 'a' && p.r.Peek(2) == 'l' && p.r.Peek(3) == 's' && p.r.Peek(4) == 'e' {
		p.r.Move(5)
		return true
	} else if c == 'n' && p.r.Peek(1) == 'u' && p.r.Peek(2) == 'l' && p.r.Peek(3) == 'l' {
		p.r.Move(4)
		return true
	}
	return false
}

func (p *Parser) consumeNumberToken() bool {
	mark := p.r.Pos()
	if p.r.Peek(0) == '-' {
		p.r.Move(1)
	}
	c := p.r.Peek(0)
	if c >= '1' && c <= '9' {
		p.r.Move(1)
		for {
			if c := p.r.Peek(0); c < '0' || c > '9' {
				break
			}
			p.r.Move(1)
		}
	} else if c != '0' {
		p.r.Rewind(mark)
		return false
	} else {
		p.r.Move(1) // 0
	}
	if c := p.r.Peek(0); c == '.' {
		p.r.Move(1)
		if c := p.r.Peek(0); c < '0' || c > '9' {
			p.r.Move(-1)
			return true
		}
		for {
			if c := p.r.Peek(0); c < '0' || c > '9' {
				break
			}
			p.r.Move(1)
		}
	}
	mark = p.r.Pos()
	if c := p.r.Peek(0); c == 'e' || c == 'E' {
		p.r.Move(1)
		if c := p.r.Peek(0); c == '+' || c == '-' {
			p.r.Move(1)
		}
		if c := p.r.Peek(0); c < '0' || c > '9' {
			p.r.Rewind(mark)
			return true
		}
		for {
			if c := p.r.Peek(0); c < '0' || c > '9' {
				break
			}
			p.r.Move(1)
		}
	}
	return true
}

func (p *Parser) consumeStringToken() bool {
	// assume to be on "
	p.r.Move(1)
	for {
		c := p.r.Peek(0)
		if c == '"' {
			p.r.Move(1)
			break
		} else if c == '\\' && (p.r.Peek(1) != 0 || p.r.Err() == nil) {
			p.r.Move(1)
		} else if c == 0 {
			return false
		}
		p.r.Move(1)
	}
	return true
}
