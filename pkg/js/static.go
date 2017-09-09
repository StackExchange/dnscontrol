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
		size:    14087,
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
