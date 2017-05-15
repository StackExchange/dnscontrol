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
		size:    8291,
		modtime: 0,
		compressed: `
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
s+3SJH1VZU2bhDgvUiyTI4pjhjkPQI9kBRARVImi7qxcU4ts7EZsfWQNTXfYLcPvuz3F3V+afBkPoX1x
rjs1NTQ1o1Yz/d09A41xRGIMj4jjGGimB8gl/QFctCahXE9C5Z1fdxOAuHoq+4Ga9Xrn1FPSNiafilZ7
LoTPF3D1UEvWnlfbURpWOdzeu0486WZMRcyeaAJrjiXpZmTeWHvbMBZSl+HISrzwE1NR0OaX0VSlDTXU
4qoz510GZXtQEcO7d2ANfeuFdk2qEFu8je8NFmuXcdt5Vc10ZXrqDHTfTtXyljlDqfqSUn8benB2eE/K
LONCbuNOwV0vRDTjVLZBdOXW8+WrvYNlx6/myj447t1XkuckW/3mOW1TdtbfODAj4vJTVNT82MJwNNSp
mORQf+2pihSHJaMprIXIw8NDLlD0lW4wWyb0KYhoeogO/3x89OFPvx8dHp8cn54eyZy+Iahk+II2iEeM
5CJAj7QQiichjwyx58PHhOQm/oK1SK3yeuPGVHg9a2ANI4ipCHieEOH2g37TClf9G8Szo7n3/ycfTr2B
fDiee9bTSePp/dxrfWMq25kiLRWTpXxS07NqeObZHzaVbqfx0bCMJH23VdK6LFmRtlJvrLPz/518ON1R
oN7LTvovKq8cHOjzYY3wJES4QmIdLBNKmdR5KO2sw8OSDgPoB30YQLxj3BebUIBPCS3iZYIYBpQQxDEP
9bwACzUNFzI7KIwki8mGxAVKym8Rgfpo/OlicXN7/fCvxfXFhSwq/agSucgZ/fbcD6FPl8v+dqggyiZC
voaYcNmZxG0xk/1SslKIJQZnu6Rc3F9e7pWzLJJESyqlDG4RSVZFVkuTK5gdlJ+DbHeEvdoGM6Smy6Wu
epkg1WcBcK05thc2AZpR/16vLQxf7b0dWrOu0n1qdnu1oUV6t/efAAAA//8jBMkCYyAAAA==
`,
	},

	"/": {
		isDir: true,
		local: "js",
	},
}
