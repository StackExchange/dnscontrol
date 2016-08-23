package JsonConfigReader

import (
	"bytes"
	"io"
)

type state struct {
	r  io.Reader
	br *bytes.Reader
}

func isNL(c byte) bool {
	return c == '\n' || c == '\r'
}

func isWS(c byte) bool {
	return c == ' ' || c == '\t' || isNL(c)
}

func consumeComment(s []byte, i int) int {
	if i < len(s) && s[i] == '/' {
		s[i-1] = ' '
		for ; i < len(s) && !isNL(s[i]); i += 1 {
			s[i] = ' '
		}
	}
	return i
}

func prep(r io.Reader) (s []byte, err error) {
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, r)
	s = buf.Bytes()
	if err != nil {
		return
	}

	i := 0
	for i < len(s) {
		switch s[i] {
		case '"':
			i += 1
			for i < len(s) {
				if s[i] == '"' {
					i += 1
					break
				} else if s[i] == '\\' {
					i += 1
				}
				i += 1
			}
		case '/':
			i = consumeComment(s, i+1)
		case ',':
			j := i
			for {
				i += 1
				if i >= len(s) {
					break
				} else if s[i] == '}' || s[i] == ']' {
					s[j] = ' '
					break
				} else if s[i] == '/' {
					i = consumeComment(s, i+1)
				} else if !isWS(s[i]) {
					break
				}
			}
		default:
			i += 1
		}
	}
	return
}

// Read acts as a proxy for the underlying reader and cleans p
// of comments and trailing commas preceeding ] and }
// comments are delimitted by // up until the end the line
func (st *state) Read(p []byte) (n int, err error) {
	if st.br == nil {
		var s []byte
		if s, err = prep(st.r); err != nil {
			return
		}
		st.br = bytes.NewReader(s)
	}
	return st.br.Read(p)
}

// New returns an io.Reader acting as proxy to r
func New(r io.Reader) io.Reader {
	return &state{r: r}
}
