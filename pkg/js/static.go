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
		size:    13380,
		modtime: 0,
		compressed: `
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
vwEAAP//OlEbP0Q0AAA=
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
