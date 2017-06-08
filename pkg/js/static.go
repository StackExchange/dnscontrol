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
		size:    9402,
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
