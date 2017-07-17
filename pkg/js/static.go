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
		size:    9590,
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
