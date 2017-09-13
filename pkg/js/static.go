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
		size:    13505,
		modtime: 0,
		compressed: `
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
AAA=
`,
	},

	"/": {
		isDir: true,
		local: "pkg/js",
	},
}
