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
		size:    11677,
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
