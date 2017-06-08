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
		size:    10314,
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
