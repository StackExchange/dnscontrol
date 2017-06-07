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
		size:    11833,
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
