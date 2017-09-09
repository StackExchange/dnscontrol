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
		size:    14060,
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
