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
		size:    8855,
		modtime: 0,
		compressed: `
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
9Q1D/CcAAP//qJ7f15ciAAA=
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
