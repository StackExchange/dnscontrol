package js

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

type _escLocalFS struct{}

var _escLocal _escLocalFS

type _escStaticFS struct{}

var _escStatic _escStaticFS

type _escDirectory struct {
	fs   http.FileSystem
	name string
}

type _escFile struct {
	compressed string
	size       int64
	modtime    int64
	local      string
	isDir      bool

	once sync.Once
	data []byte
	name string
}

func (_escLocalFS) Open(name string) (http.File, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	return os.Open(f.local)
}

func (_escStaticFS) prepare(name string) (*_escFile, error) {
	f, present := _escData[path.Clean(name)]
	if !present {
		return nil, os.ErrNotExist
	}
	var err error
	f.once.Do(func() {
		f.name = path.Base(name)
		if f.size == 0 {
			return
		}
		var gr *gzip.Reader
		b64 := base64.NewDecoder(base64.StdEncoding, bytes.NewBufferString(f.compressed))
		gr, err = gzip.NewReader(b64)
		if err != nil {
			return
		}
		f.data, err = ioutil.ReadAll(gr)
	})
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (fs _escStaticFS) Open(name string) (http.File, error) {
	f, err := fs.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.File()
}

func (dir _escDirectory) Open(name string) (http.File, error) {
	return dir.fs.Open(dir.name + name)
}

func (f *_escFile) File() (http.File, error) {
	type httpFile struct {
		*bytes.Reader
		*_escFile
	}
	return &httpFile{
		Reader:   bytes.NewReader(f.data),
		_escFile: f,
	}, nil
}

func (f *_escFile) Close() error {
	return nil
}

func (f *_escFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, nil
}

func (f *_escFile) Stat() (os.FileInfo, error) {
	return f, nil
}

func (f *_escFile) Name() string {
	return f.name
}

func (f *_escFile) Size() int64 {
	return f.size
}

func (f *_escFile) Mode() os.FileMode {
	return 0
}

func (f *_escFile) ModTime() time.Time {
	return time.Unix(f.modtime, 0)
}

func (f *_escFile) IsDir() bool {
	return f.isDir
}

func (f *_escFile) Sys() interface{} {
	return f
}

// _escFS returns a http.Filesystem for the embedded assets. If useLocal is true,
// the filesystem's contents are instead used.
func _escFS(useLocal bool) http.FileSystem {
	if useLocal {
		return _escLocal
	}
	return _escStatic
}

// _escDir returns a http.Filesystem for the embedded assets on a given prefix dir.
// If useLocal is true, the filesystem's contents are instead used.
func _escDir(useLocal bool, name string) http.FileSystem {
	if useLocal {
		return _escDirectory{fs: _escLocal, name: name}
	}
	return _escDirectory{fs: _escStatic, name: name}
}

// _escFSByte returns the named file from the embedded assets. If useLocal is
// true, the filesystem's contents are instead used.
func _escFSByte(useLocal bool, name string) ([]byte, error) {
	if useLocal {
		f, err := _escLocal.Open(name)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		f.Close()
		return b, err
	}
	f, err := _escStatic.prepare(name)
	if err != nil {
		return nil, err
	}
	return f.data, nil
}

// _escFSMustByte is the same as _escFSByte, but panics if name is not present.
func _escFSMustByte(useLocal bool, name string) []byte {
	b, err := _escFSByte(useLocal, name)
	if err != nil {
		panic(err)
	}
	return b
}

// _escFSString is the string version of _escFSByte.
func _escFSString(useLocal bool, name string) (string, error) {
	b, err := _escFSByte(useLocal, name)
	return string(b), err
}

// _escFSMustString is the string version of _escFSMustByte.
func _escFSMustString(useLocal bool, name string) string {
	return string(_escFSMustByte(useLocal, name))
}

var _escData = map[string]*_escFile{

	"/helpers.js": {
		local:   "js/helpers.js",
		size:    8460,
		modtime: 0,
		compressed: `
H4sIAAAAAAAA/7w5bW/bONLf/StmBTy19ESRnWTbO8jrw/k2yaK42A0c5y6AYQSMRNtsJVEg6bi5wvnt
B75IoiS7SYDL9kNqkfM+w5nh0NlwDFwwEgln0Ok8IgYRzZYwhB8dAACGV4QLhhgPYb7w1Vqc8fuc0UcS
49oyTRHJ1EJnZ2jFeIk2iRixFYchzBeDTme5ySJBaAYkI4KghPwHu55mVuN8iPtPJGhKIb93Ay1cS5Cd
JcoEb6cFKzdDKfbFU479FAvkGXHIEly56JXiyS8YDsEZjya3oytHM9qpv1J3hldSGUkuBEVUoYTqrw+S
eKj+GhGl9kGlcZBv+NpleOUNjCfEhmWKUEv484xfG3O4FSfNw1IAXKUCXaoNGA6H0KUPX3Ekuh58+ABu
l+T3Ec0eMeOEZrwLJNM0PMspciGoA8IQlpSlSNwL4e7Z9xqmiXn+dtPUnK6tE/P8JetkeHuuQkIbprSv
Vwa4QqzJUgKF1U8j1Y+d3I4oi3k4X/gyEq+rQJS7JtJms6sQ+r6iyDGTlgjni11duJzRCHN+jtiKu6lv
gtc2dq8nLQsYRWtIaUyWBDNf+pIIIBxQEAQ1WEM5hAgliQTaErE2dG1AxBh6CgsBpEobxskjTp5sKB0c
0hVshRXLTFBliBgJVELKs3EfEH5puLtpLWCKuHGNeoNyZwc44bjEH0mh9iBLC7gybr6qgGzTrttx/nVR
mrIGuDvE+IvScw/n+wB/FziLjeiBVN1P2xrYWGLN6Bacf4+mk8+TP0IjSek9nTc2Gd/kOWUCxyE4R1Cc
SzgCB3TAqnXDV8d1pceu0+n14LwZ0yH8zjASGBCcT24MnQBuOQaxxpAjhlIsMOOAeBHGgLJYCseDKi5b
hI2C6uxqdYaHT5YWtHQagSH0B0B+s5NwkOBsJdYDIEdHXmm9mh8t6DlZ+JZDd20Gp5IBYqtNijNRp245
R0KnMIQScE4WlVkPnMYqd+k0pAuMSUAGxPjj4nJ0ezW7AZOmOCDgWABdFqpXnEFQQHmePKkfSQLLjdgw
XNSvQNK7kKdeHWRBK+JbkiQQJRgxQNkT5Aw/Errh8IiSDeaSoe1Jg1WU2HYd3O+rF01p+1KZwrapV9RC
bZfZ7Mp99EK4wULF4Wx2pVjqKNVxaMmswev5udh0mS0EC4RIYAiPdX7nZQqusS18ULBXa/qIWAazcQ/I
ENcMEVQZvyGKFsaqzU5RvyYoxY4PfQ8kSMZ/p5tMxUkfUowyDjHNugJkc0aZKUJY+9sqKIGNnFFRxB0z
RCQ6ShJbu1ajYNC9okkoOoSCrGoSNlmMlyTDcbc6qxUEHJ/Yvc9L1rIq5lzKsJC5RNOqu3GkRSR5UXLH
JoXyIAi8SikDByS385RMaTCEFRYlWhWj/qn3sqwojqeKrxv7zsjxC2kkZa8u6Wj0amFL0HeWdzT6uchX
n0c3ptdFbIXFS3JX8KAR3lN4ycxIb6RraCBV+H0yGl+8QQUL/v1VUMx+qoJMjHezN8hfQr+/9LO72Uuy
306verfTq7P+Se9yOhpfAGJYJadojVEOPMcRWZIIOMawFiLnYa+33W6DEiSIaNozXVAP5aSXYrGmMe+Z
EngcZ7zHsTheUy54gHj+3fB9g81K6Pe32e306hU2O+ufvCT+Wf+krkGJ86cocdY/eUkP5fA3eMGCf38V
FLOXNBjfaXFyRigj4ul1ehRYUKI11InWOPomeyJ3Lu8VN4KRbOWD/D3ZpA/y7latL/yqHfTBGd8B/p7j
SHA4xMXxXmm0s1cYTfX8qnUr+Fj3GtuiUjTHB9t9PjRMWpqosoD6xZWOXF6LeeRVoxRU3QHgN41UfFst
hrpKuQrVajD23CxqBBqXCsXvFw0xJwvFWvaoXv2qV/E6cuC49Aw4R+TIkXdt2WBFlDEcCXVdczzrQmbH
1uQtdXXypxXVyc8rqhR8NL64uZj+62JqK2AL2wBoCP1C52d3riru6gMgRSo0/+/2xVY1YxIMZVx+3gv0
kJihnCyokv98ntBtCCc+rMlqHcKpL++q/0Ach3C28EFv/1psf1Tbn69D+LRYaDJqzOGcwDOcwjOcwfMA
foVn+AjPAM/wyeloByUkw/oa1bGjcihjEn6DhpD7blIKPodhE7a8l0oAJR0MgeSB+jkoT5H6rEW6NUfR
m40oL2jdBynKNYhf+ot4P4o52iY9jalwibfzgq+UZK7j2/GOE473Ey4wNfdB64hYSkmPlGrJj5picuEn
qqnttnKGZqme/P6fKWiIWyoqKQ4ryehWhofZL3nmQUK3nt9elgFZrRvpO5aB1W892FbBZ4bEdGt0gGdw
PKmGlMGoqgHN/gCcYlrxeXz9ZTq7n01Hk5vLL9OxPlQJkpbSUViNQMoj+HokX4jkVYlBz8ojGDaKTpOV
44Pzd6ckX5pV//vRbRyhbtjMF7aU3m7h1QqElLbucIYjM14QImn7WBvx+nb6x4VrGUgvGAXj4J8Y57fZ
t4xuMxjCEiUcF8n2y30LuVw7gC/YBtcyYrM2cJ8LxPZVkb2jHgU8UNOeg4Oeqk0oCmf7ri9h6pNt25Vq
qN+qPIaFzLZLk/RVlTVtEuJ8k2KZHFEcM8x5APpBQQARQZkoqs7KNbXIlt2QrY6sgWk/1cjw+2G/QRwu
Tb6Mh9Ae+1Sdmhr5m4cC83axf4If44jEGB4QxzHQTD9/FPDHcNmY43M9xxdrbLoJQFx9Ff1Ahfpl78xe
wtbm9gpWWy6Ez5cwvqsoa8srdxSKlQa3fdeKJ92MqYg5EE1gTWEl3Jwsanuve0qA1GU4shIvvGGmD1r9
IprKtKFGslx15ryNoHQPSmD48AGsJ4tqo1mTSokt3NprmYXaRty1lsoXCZmeWs8Rr4dqWMucoVS9A1Yv
m3fOHutJmkVcSDfuJdy2QkQzTmUbRFdu9ToyPvgs4vjlq4gPjnvzjeQ5yVa/eE5Tlb31Nw7MA0fxkBrV
nwoZjgY6FZMcqrfKskhxWDKaqtFF2OtxgaJv9BGzZUK3anKBen896X/8y6/93snpyadPfZnTHwkqEL6i
R8QjRnIRoAe6EQonIQ8MsafeQ0JyE3/BWqRWeb12Yyq8jvXcAkOIqQh4nhDhdoNuXQtX/TuK5/2F9/+n
Hz95R/LjZOFZX6e1r7OF13ghLdqZTVowJkv5pWa/5ejXs5/lFW+n9uRdRJK+2ypqbZRskzZSb6yz8/+d
fvy0p0CdyU76byqvHB/r82ENoKWIMEZiHSwTSpnk2ZN6VuFhUYcj6AZdOIJ4z7A6lib5bwAAAP//VPEp
CAwhAAA=
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
