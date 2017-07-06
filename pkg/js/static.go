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
		local:   "pkg/js/helpers.js",
		size:    11049,
		modtime: 0,
		compressed: `
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
Yoh/BQAA//+IcjlWKSsAAA==
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
