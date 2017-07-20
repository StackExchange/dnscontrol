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
		size:    11080,
		modtime: 0,
		compressed: `
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
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
